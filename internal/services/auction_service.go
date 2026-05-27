package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Definindo os tipos de mensagens que podem ser enviadas pelos clientes para o servidor, como por exemplo, uma mensagem para realizar um lance em um leilão
type MessageKind int

const (
	PlaceBid MessageKind = iota
	// Ok/Success
	SuccessfullyPlacedBid

	// Errors
	FailedToPlaceBid
	InvalidJSON

	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserID  uuid.UUID   `json:"user_id,omitempty"`
}

// AuctionLobby é a estrutura que gerencia as salas de leilão, onde cada sala representa um leilão específico para um produto. Ela utiliza um mutex para garantir a segurança em operações concorrentes e um mapa para armazenar as salas de leilão, associando cada sala a um UUID que representa o ID do produto leiloado.
type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

// Client representa um cliente conectado a uma sala de leilão, contendo informações sobre a sala à qual o cliente está conectado, a conexão WebSocket para comunicação em tempo real, um canal para enviar mensagens ao cliente e o ID do usuário associado ao cliente.
type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

// AuctionRoom representa uma sala de leilão específica para um produto, contendo um ID único, um contexto para controle de tempo de vida, canais para broadcast de mensagens, registro e desregistro de clientes, e um mapa para armazenar os clientes conectados à sala.
type AuctionRoom struct {
	Id          uuid.UUID
	Context     context.Context
	Broadcast   chan Message
	Register    chan *Client
	Unregister  chan *Client
	Clients     map[uuid.UUID]*Client
	BidsService BidsService
}

// NewAuctionRoom é uma função que cria e retorna uma nova instância de AuctionRoom, inicializando os canais e o mapa de clientes, e associando o ID do produto ao leilão correspondente.
func NewAuctionRoom(ctx context.Context, id uuid.UUID, BidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Context:     ctx,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		BidsService: BidsService,
		Clients:     make(map[uuid.UUID]*Client),
	}
}

// NewClient é uma função que cria e retorna uma nova instância de Client, associando o cliente a uma sala de leilão específica, configurando a conexão WebSocket para comunicação em tempo real, inicializando o canal de envio de mensagens e associando o ID do usuário ao cliente.
func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message),
		UserId: userId,
	}
}

// registerClient é um método da estrutura AuctionRoom que registra um cliente na sala de leilão, adicionando o cliente ao mapa de clientes conectados e registrando a conexão do cliente para receber mensagens relacionadas ao leilão.
func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user Connected", "Client", c)
	r.Clients[c.UserId] = c
}

// unregisterClient é um método da estrutura AuctionRoom que desregistra um cliente da sala de leilão, removendo o cliente do mapa de clientes conectados e encerrando a conexão do cliente para que ele não receba mais mensagens relacionadas ao leilão.
func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("User disconnected", "Client", c)
	delete(r.Clients, c.UserId)
}

// broadcastMessage é um método da estrutura AuctionRoom que processa as mensagens recebidas dos clientes e as distribui para os clientes conectados à sala de leilão, realizando ações específicas com base no tipo de mensagem, como por exemplo, processar um lance realizado por um cliente e notificar os outros clientes sobre o novo lance.
func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "RoomID", r.Id, "message", m.Message, "user_id", m.UserID, "amount", m.Amount, "kind", m.Kind)
	switch m.Kind {
	case PlaceBid:
		bid, err := r.BidsService.PlaceBid(r.Context, r.Id, m.UserID, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := r.Clients[m.UserID]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error(), UserID: m.UserID}
				}
				return
			}
		}

		if client, ok := r.Clients[m.UserID]; ok {
			client.Send <- Message{Kind: SuccessfullyPlacedBid, Message: "Your bid was Successfully placed.", UserID: m.UserID}
		}

		for id, client := range r.Clients {
			newBidMessage := Message{Kind: NewBidPlaced, Message: "A new bid was placed", Amount: bid.BidAmount, UserID: m.UserID}
			if id == m.UserID {
				continue
			}
			client.Send <- newBidMessage
		}
	}
}

// Run é um método da estrutura AuctionRoom que inicia o loop principal da sala de leilão, onde ele aguarda por eventos como registro e desregistro de clientes, recebimento de mensagens para broadcast e o término do leilão, processando cada evento de acordo com a lógica definida para cada tipo de evento.
func (r *AuctionRoom) Run() {
	slog.Info("Auction has started.", "auctionID", r.Id)
	// Garantindo que os canais de broadcast, registro e desregistro sejam fechados quando a função Run terminar sua execução, para evitar vazamentos de recursos e garantir que os clientes sejam notificados sobre o término do leilão.
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		// select é utilizado para aguardar por múltiplos eventos simultaneamente, permitindo que a sala de leilão processe eventos de registro e desregistro de clientes, recebimento de mensagens para broadcast e o término do leilão de forma eficiente e responsiva.
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended.", "auctionID", r.Id)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinished, Message: "auction has been finished"}
			}
			return
		}
	}
}

const (
	maxMessageSize = 512                     // maxMessageSize define o tamanho máximo permitido para as mensagens enviadas pelos clientes através da conexão WebSocket, garantindo que mensagens muito grandes sejam rejeitadas para evitar sobrecarga do servidor e garantir uma comunicação eficiente entre o cliente e o servidor.
	readDeadline   = 60 * time.Second        // readDeadline define o tempo máximo permitido para que o servidor receba uma mensagem do cliente através da conexão WebSocket, garantindo que conexões inativas sejam encerradas e recursos sejam liberados de forma eficiente.
	pingPeriod     = (readDeadline * 9) / 10 // pingPeriod define o intervalo de tempo para enviar mensagens de ping para os clientes através da conexão WebSocket, garantindo que a conexão seja mantida ativa e detectando conexões inativas de forma eficiente.
	writeWait      = 10 * time.Second        // writeWait define o tempo máximo permitido para que o servidor envie uma mensagem para o cliente através da conexão WebSocket, garantindo que conexões inativas sejam encerradas e recursos sejam liberados de forma eficiente.

)

// ReadEventLoop é um método da estrutura Client que inicia um loop de leitura de eventos, onde ele aguarda por mensagens enviadas pelo cliente através da conexão WebSocket, processando cada mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.
func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)                  // Configurando o limite de tamanho para as mensagens recebidas através da conexão WebSocket, garantindo que mensagens muito grandes sejam rejeitadas para evitar sobrecarga do servidor e garantir uma comunicação eficiente entre o cliente e o servidor.
	c.Conn.SetReadDeadline(time.Now().Add(readDeadline)) // Configurando o tempo máximo permitido para que o servidor receba uma mensagem do cliente através da conexão WebSocket, garantindo que conexões inativas sejam encerradas e recursos sejam liberados de forma eficiente.
	c.Conn.SetPongHandler(func(string) error {           // Configurando um manipulador para mensagens de pong recebidas do cliente, que é utilizado para manter a conexão WebSocket ativa e detectar conexões inativas, garantindo que conexões inativas sejam encerradas de forma eficiente.
		c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for { // Iniciando um loop infinito para aguardar por mensagens enviadas pelo cliente através da conexão WebSocket, processando cada mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.
		var m Message              // Criando uma variável para armazenar a mensagem recebida do cliente, que será processada de acordo com a lógica definida para cada tipo de mensagem.
		m.UserID = c.UserId        // Associando o ID do usuário ao campo UserID da mensagem recebida, garantindo que as mensagens possam ser processadas de forma associada ao usuário que as enviou, permitindo uma comunicação eficiente e personalizada entre o cliente e o servidor.
		err := c.Conn.ReadJSON(&m) // Lendo uma mensagem JSON enviada pelo cliente através da conexão WebSocket e armazenando-a na variável m, processando a mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.
		if err != nil {            // Verificando se ocorreu um erro durante a leitura da mensagem JSON enviada pelo cliente, o que pode indicar que a mensagem não é um JSON válido ou que a conexão WebSocket foi encerrada de forma inesperada, garantindo que conexões inativas sejam encerradas de forma eficiente.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected Close error", "error", err)
				return
			}

			c.Room.Broadcast <- Message{ // Enviando uma mensagem de erro para o canal de broadcast da sala de leilão, indicando que a mensagem recebida do cliente não é um JSON válido, garantindo que os clientes sejam notificados sobre mensagens inválidas e permitindo uma comunicação eficiente entre o cliente e o servidor.
				Kind:    InvalidJSON,
				Message: "this message should be a valid json",
				UserID:  m.UserID,
			}

			continue
		}
		c.Room.Broadcast <- m // Enviando a mensagem recebida do cliente para o canal de broadcast da sala de leilão, permitindo que a mensagem seja processada e distribuída para os clientes conectados à sala de leilão de acordo com a lógica definida para cada tipo de mensagem.
	}
}

func (c *Client) WriteEventLoop() {
	// ticker é utilizado para enviar mensagens de ping para os clientes em intervalos regulares, garantindo que a conexão WebSocket seja mantida ativa e detectando conexões inativas de forma eficiente.
	ticker := time.NewTicker(pingPeriod)
	// Garantindo que o ticker seja parado e a conexão WebSocket seja fechada quando a função WriteEventLoop terminar sua execução, para evitar vazamentos de recursos e garantir que conexões inativas sejam encerradas de forma eficiente.
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send: // Aguardando por mensagens a serem enviadas para o cliente através do canal de envio de mensagens, processando cada mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.

			if !ok { // Verificando se o canal de envio de mensagens foi fechado, o que pode indicar que a conexão WebSocket foi encerrada de forma inesperada, garantindo que conexões inativas sejam encerradas de forma eficiente.
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing websocket conn",
				})
			}

			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteJSON(message) // Enviando uma mensagem JSON para o cliente através da conexão WebSocket, processando a mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.
			if err != nil {
				c.Room.Unregister <- c
				return
			}

		case <-ticker.C: // Aguardando pelo evento de ticker para enviar mensagens de ping para o cliente, garantindo que a conexão WebSocket seja mantida ativa e detectando conexões inativas de forma eficiente.
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil { // Enviando uma mensagem de ping para o cliente através da conexão WebSocket, processando a mensagem de acordo com a lógica definida para cada tipo de mensagem e garantindo que conexões inativas sejam encerradas de forma eficiente.
				slog.Error("Unexpected write error", "error", err)
				return
			}

		}
	}
}

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/jsonutils"
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	// Extraindo o ID do produto da URL, que é necessário para identificar a qual leilão o usuário deseja se inscrever
	fmt.Println("Full URL:", r.URL.Path)
	fmt.Println("Route pattern:", r.URL.RawPath)
	rawProductId := chi.URLParam(r, "product_id")

	fmt.Println("Id:", rawProductId+" - "+fmt.Sprintf("%T", rawProductId))
	// Convertendo o ID do produto de string para UUID, que é o formato utilizado internamente para identificar os produtos
	productId, err := uuid.Parse(rawProductId)
	// Verificando se ocorreu um erro durante a conversão do ID do produto, o que pode indicar que o ID fornecido não é um UUID válido
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "Invalid product id - must be a valid uuid",
		})
		return
	}

	// Verificando se o produto com o ID fornecido existe no banco de dados, utilizando o serviço de produtos para buscar o produto pelo ID
	_, err = a.ProductService.GetProductByID(r.Context(), productId)

	// Tratando o erro retornado pelo serviço de produtos, verificando se o erro é do tipo "produto não encontrado" ou se é um erro inesperado
	if err != nil {
		// Se o erro for do tipo "produto não encontrado", retornamos uma resposta HTTP 404 Not Found com uma mensagem indicando que não há um produto com o ID fornecido
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": "no product with given id",
			})
			return
		}
		// Se o erro for diferente de "produto não encontrado", retornamos uma resposta HTTP 500 Internal Server Error com uma mensagem genérica indicando que ocorreu um erro inesperado
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "unexpected error, try again later.",
		})
		return
	}

	// Extraindo o ID do usuário autenticado da sessão, que é necessário para associar a inscrição do usuário ao leilão correto
	userId, ok := a.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)

	// Verificando se o ID do usuário foi extraído corretamente da sessão, o que é essencial para garantir que apenas usuários autenticados possam se inscrever nos leilões
	if !ok {
		// Se o ID do usuário não puder ser extraído da sessão, retornamos uma resposta HTTP 500 Internal Server Error com uma mensagem genérica indicando que ocorreu um erro inesperado
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "unexpected error, try again later.",
		})
		return
	}
	// Aqui, a lógica para inscrever o usuário no leilão seria implementada, associando o ID do usuário ao ID do produto no banco de dados ou em uma estrutura de dados em memória

	a.ActionLobby.Lock()                       // Adquirindo um bloqueio para garantir que a leitura do mapa de salas do lobby seja feita de forma segura e sem condições de corrida, evitando problemas de concorrência ao acessar o mapa de salas do lobby
	room, ok := a.ActionLobby.Rooms[productId] // Buscando a sala de leilão correspondente ao ID do produto no mapa de salas do lobby
	a.ActionLobby.Unlock()                     // Liberando o bloqueio após a leitura do mapa de salas do lobby, permitindo que outras operações possam acessar o mapa de salas do lobby sem bloqueios desnecessários

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})
		return
	}

	// Após a inscrição do usuário no leilão, podemos estabelecer uma conexão WebSocket para enviar atualizações em tempo real sobre o leilão para o cliente
	conn, err := a.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("ws upgrade error:", err)
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "could not upgrade connection to a websocket protocol",
		})
		return
	}

	client := services.NewClient(room, conn, userId) // Criando um novo cliente para a sala de leilão, associando o cliente à sala de leilão correspondente e configurando a conexão WebSocket para comunicação em tempo real

	room.Register <- client // Registrando o cliente na sala de leilão, o que permitirá que o cliente receba mensagens relacionadas ao leilão e participe das interações em tempo real

	go client.ReadEventLoop()  // Iniciando a rotina de leitura de eventos do cliente, que ficará responsável por ler as mensagens enviadas pelo cliente através da conexão WebSocket e processá-las de acordo com a lógica do leilão
	go client.WriteEventLoop() // Iniciando a rotina de escrita de eventos do cliente, que ficará responsável por enviar mensagens para o cliente através da conexão WebSocket, permitindo que o cliente receba atualizações em tempo real sobre o leilão e as interações relacionadas
}

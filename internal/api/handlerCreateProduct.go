package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/jsonutils"
	"github.com/Mateus-R-De-Lima/GoBid/internal/services"
	"github.com/Mateus-R-De-Lima/GoBid/internal/usecase/product"
	"github.com/google/uuid"
)

func (a *Api) handlerCreateProduct(w http.ResponseWriter, r *http.Request) {

	data, problemns, err := jsonutils.DecodeJson[product.CreateProductRequest](r)
	fmt.Println("data : " + fmt.Sprint(data))
	fmt.Println("problems : " + fmt.Sprint(problemns))
	fmt.Println("err : " + fmt.Sprint(err))

	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problemns)
		return
	}

	userID, ok := a.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}

	productId, err := a.ProductService.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd,
	)

	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}

	ctx, cancel := context.WithDeadline(context.Background(), data.AuctionEnd)

	defer cancel()

	auctionRoom := services.NewAuctionRoom(ctx, productId, a.BidsService)

	go auctionRoom.Run()                         // Iniciando a sala de leilão em uma goroutine separada para processar eventos de forma assíncrona e permitir que a API continue respondendo a outras requisições enquanto o leilão está ativo
	a.ActionLobby.Lock()                         // Adquirindo um bloqueio para garantir que a adição da nova sala de leilão ao lobby seja feita de forma segura e sem condições de corrida, evitando problemas de concorrência ao acessar o mapa de salas do lobby
	a.ActionLobby.Rooms[productId] = auctionRoom // Adicionando a nova sala de leilão ao mapa de salas do lobby, associando o ID do produto ao leilão correspondente para que os clientes possam se inscrever e participar do leilão
	a.ActionLobby.Unlock()                       // Liberando o bloqueio após a adição da nova sala de leilão ao lobby, permitindo que outras operações possam acessar o mapa de salas do lobby sem bloqueios desnecessários

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"product_id": productId,
		"message":    "Auction has started with success",
	})

}

package api

import (
	"fmt"
	"net/http"

	"github.com/Mateus-R-De-Lima/GoBid/internal/jsonutils"
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

	id, err := a.ProductService.CreateProduct(
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

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"product_id": id,
		"message":    "product created with success",
	})

}

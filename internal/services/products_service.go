package services

import (
	"context"
	"errors"
	"time"

	"github.com/Mateus-R-De-Lima/GoBid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductsService(pool *pgxpool.Pool) *ProductsService {
	return &ProductsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductsService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName string,
	description string,
	basePrice float64,
	auctionEnd time.Time,
) (uuid.UUID, error) {

	productId, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: productName,
		Description: description,
		Baseprice:   basePrice,
		AuctionEnd:  auctionEnd,
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	return productId, nil
}

var ErrProductNotFound = errors.New("product not found")

func (ps *ProductsService) GetProductByID(ctx context.Context, productId uuid.UUID) (pgstore.Product, error) {

	product, err := ps.queries.GetProductById(ctx, productId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}

	return product, nil
}

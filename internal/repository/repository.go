package repository

import (
	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user domain.User) (int, error)
	SignUser(email string) (domain.User, error)
}
type Pvz interface {
	CreatePvz(pvz domain.PVZ) (domain.PVZ, error)
	GetPvz(input domain.GettingPvzParams) ([]domain.PvzSummary, error)
	CreateRecep(recep domain.ProductReception) (domain.ProductReception, error)
	AddProdToRecep(product domain.Product) (domain.Product, error)
	DeleteLastProduct(delProd uuid.UUID) error
	CloseReception(closeRec uuid.UUID) (domain.ProductReception, error)
}

type Repository struct {
	Authorization
	Pvz
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Pvz:           NewPvzPostgres(db),
	}
}

package usecase

import (
	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/bllooop/pvzservice/internal/repository"
	"github.com/google/uuid"
)

//go:generate mockgen -source=usecase.go -destination=mocks/mock.go -package=mocks
type Authorization interface {
	CreateUser(user domain.User) (int, error)
	SignUser(email, password string) (domain.User, error)
	GenerateToken(userId uuid.UUID, userRole int) (string, error)
	ParseToken(accessToken string) (string, int, error)
}
type Pvz interface {
	CreatePvz(pvz domain.PVZ) (domain.PVZ, error)
	GetPvz(input domain.GettingPvzParams) ([]domain.PvzSummary, error)
	CreateRecep(recep domain.ProductReception) (domain.ProductReception, error)
	AddProdToRecep(product domain.Product) (domain.Product, error)
	DeleteLastProduct(delProd uuid.UUID) error
	CloseReception(closeRec uuid.UUID) (domain.ProductReception, error)
}
type Usecase struct {
	Authorization
	Pvz
}

func NewUsecase(repo *repository.Repository) *Usecase {
	return &Usecase{
		Authorization: NewAuthUsecase(repo),
		Pvz:           NewPvzUsecase(repo),
	}
}

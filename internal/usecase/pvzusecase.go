package usecase

import (
	"context"

	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/bllooop/pvzservice/internal/repository"
	"github.com/google/uuid"
)

type PvzUsecase struct {
	repo repository.Pvz
}

func NewPvzUsecase(repo *repository.Repository) *PvzUsecase {
	return &PvzUsecase{
		repo: repo,
	}
}

func (s *PvzUsecase) CreatePvz(pvz domain.PVZ) (domain.PVZ, error) {
	return s.repo.CreatePvz(pvz)
}
func (s *PvzUsecase) GetPvz(input domain.GettingPvzParams) ([]domain.PvzSummary, error) {
	return s.repo.GetPvz(input)
}
func (s *PvzUsecase) CreateRecep(recep domain.ProductReception) (domain.ProductReception, error) {
	return s.repo.CreateRecep(recep)
}

func (s *PvzUsecase) AddProdToRecep(product domain.Product) (domain.Product, error) {
	return s.repo.AddProdToRecep(product)
}

func (s *PvzUsecase) DeleteLastProduct(delProd uuid.UUID) error {
	return s.repo.DeleteLastProduct(delProd)
}
func (s *PvzUsecase) CloseReception(closeRec uuid.UUID) (domain.ProductReception, error) {
	return s.repo.CloseReception(closeRec)
}

func (s *PvzUsecase) GetListOFpvz(ctx context.Context) ([]domain.PVZ, error) {
	return s.repo.GetListOFpvz(ctx)
}

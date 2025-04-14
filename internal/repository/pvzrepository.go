package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/bllooop/pvzservice/internal/domain"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/jmoiron/sqlx"
)

type PvzPostgres struct {
	db *sqlx.DB
}

func NewPvzPostgres(db *sqlx.DB) *PvzPostgres {
	return &PvzPostgres{
		db: db,
	}
}
func (r *PvzPostgres) beginTx() (*sqlx.Tx, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
func (r *PvzPostgres) GetListOFpvz(ctx context.Context) ([]domain.PVZ, error) {
	var pvzList []domain.PVZ
	query := fmt.Sprintf("SELECT * FROM %s", pvzTable)
	err := r.db.SelectContext(ctx, &pvzList, query)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка при выполнении запроса для получения списка ПВЗ")
		return nil, err
	}
	logger.Log.Debug().Any("query", query).Msg("Запрос данных о ПВЗ")
	return pvzList, nil
}
func (r *PvzPostgres) CreatePvz(pvz domain.PVZ) (domain.PVZ, error) {
	var pvzResponse domain.PVZ
	query := fmt.Sprintf(`INSERT INTO %s (registrationdate,city) VALUES ($1,$2) RETURNING id,registrationdate,city`, pvzTable)
	row := r.db.QueryRowx(query, pvz.DateRegister, pvz.City)
	logger.Log.Debug().Str("query", query).Msg("Выполнение запроса заведния ПВЗ")
	if err := row.Scan(&pvzResponse.Id, &pvzResponse.DateRegister, &pvzResponse.City); err != nil {
		return domain.PVZ{}, err
	}
	logger.Log.Debug().Any("pvz response", pvzResponse).Msg("Успешно заведно ПВЗ")
	return pvzResponse, nil
}
func (r *PvzPostgres) GetPvz(input domain.GettingPvzParams) ([]domain.PvzSummary, error) {
	tx, err := r.beginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var result []domain.PvzSummary

	conditions, args := buildConditions(input)
	conditionsOther := buildConditionsOther(input)
	offset := (input.Page - 1) * input.Limit
	args = append(args, input.Limit, offset)
	pvzs, err := r.queryPvzData(conditions, args)
	if err != nil {
		return nil, err
	}
	receptions, err := r.queryReceptionData(conditionsOther, args)
	if err != nil {
		return nil, err
	}

	products, err := r.queryProductData(conditionsOther, args)
	if err != nil {
		return nil, err
	}
	logger.Log.Debug().Any("receptions", receptions).Msg("Получены данные о приемках")
	logger.Log.Debug().Any("products", products).Msg("Получены данные о товарах")
	receptionMap := make(map[string][]domain.ProductReception)
	for _, reception := range receptions {
		receptionMap[reception.PVZId.String()] = append(receptionMap[reception.PVZId.String()], reception)
	}

	productMap := make(map[string][]domain.Product)
	for _, product := range products {
		productMap[product.ReceptionId.String()] = append(productMap[product.ReceptionId.String()], product)
	}

	for _, pvz := range pvzs {
		var receptionsWithProducts []domain.Receptions

		for _, reception := range receptionMap[pvz.Id.String()] {
			receptionsWithProducts = append(receptionsWithProducts, domain.Receptions{
				ReceptionInfo: reception,
				ProductInfo:   productMap[reception.Id.String()],
			})
		}

		result = append(result, domain.PvzSummary{
			PvzInfo:        pvz,
			ReceptionsInfo: receptionsWithProducts,
		})
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	logger.Log.Debug().Any("response", result).Msg("Успешно получены данные о ПВЗ")
	return result, nil
}

func buildConditions(input domain.GettingPvzParams) ([]string, []interface{}) {
	var conditions []string
	var args []interface{}

	if !input.Start.IsZero() {
		conditions = append(conditions, "registrationDate >= $1")
		args = append(args, input.Start)
	}

	if !input.End.IsZero() {
		conditions = append(conditions, "registrationDate <= $2")
		args = append(args, input.End)
	}

	return conditions, args
}
func buildConditionsOther(input domain.GettingPvzParams) []string {
	var conditionsOther []string
	if !input.Start.IsZero() {
		conditionsOther = append(conditionsOther, "date_received >= $1")
	}
	if !input.End.IsZero() {
		conditionsOther = append(conditionsOther, "date_received <= $2")
	}
	return conditionsOther
}

func (r *PvzPostgres) queryPvzData(conditions []string, args []interface{}) ([]domain.PVZ, error) {
	query := fmt.Sprintf("SELECT * FROM %s", pvzTable)
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	var pvz []domain.PVZ
	err := r.db.Select(&pvz, query, args...)
	logger.Log.Debug().Any("query", query).Msg("Запрос данных о ПВЗ")
	return pvz, err
}

func (r *PvzPostgres) queryReceptionData(conditionsOther []string, args []interface{}) ([]domain.ProductReception, error) {
	query := fmt.Sprintf("SELECT * FROM %s", receptionTable)
	if len(conditionsOther) > 0 {
		query += " WHERE " + strings.Join(conditionsOther, " AND ")
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	var receptions []domain.ProductReception
	err := r.db.Select(&receptions, query, args...)
	logger.Log.Debug().Any("query", query).Msg("Запрос данных о приемках")
	return receptions, err
}

func (r *PvzPostgres) queryProductData(conditionsOther []string, args []interface{}) ([]domain.Product, error) {
	query := fmt.Sprintf("SELECT * FROM %s", productTable)
	if len(conditionsOther) > 0 {
		query += " WHERE " + strings.Join(conditionsOther, " AND ")
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	var products []domain.Product
	err := r.db.Select(&products, query, args...)
	logger.Log.Debug().Any("query", query).Msg("Запрос данных о товарах")
	return products, err
}

func (r *PvzPostgres) DB() *sqlx.DB {
	return r.db
}

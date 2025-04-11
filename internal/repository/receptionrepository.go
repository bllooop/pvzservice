package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bllooop/pvzservice/internal/domain"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNoProductsToDelete = errors.New("нет товаров для удаления")

func (r *PvzPostgres) CreateRecep(recep domain.ProductReception) (domain.ProductReception, error) {
	tx, err := r.beginTx()
	if err != nil {
		return domain.ProductReception{}, err
	}
	defer tx.Rollback()

	lastStatus, _, err := r.getLastReceptionStatus(tx, *recep.PVZId)
	if err != nil {
		return domain.ProductReception{}, err
	}
	if lastStatus == "in_progress" {
		return domain.ProductReception{}, fmt.Errorf("Неверный запрос или есть незакрытая приемка")
	}
	createdRecep, err := r.insertReception(tx, recep)
	if err != nil {
		return domain.ProductReception{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ProductReception{}, err
	}
	logger.Log.Debug().Any("pvz response", createdRecep).Msg("Успешно создана приемка")
	return createdRecep, nil
}

func (r *PvzPostgres) AddProdToRecep(product domain.Product) (domain.Product, error) {
	tx, err := r.beginTx()
	if err != nil {
		return domain.Product{}, err
	}
	defer tx.Rollback()

	lastStatus, recepId, err := r.getLastReceptionStatus(tx, *product.PVZId)
	if lastStatus == "close" {
		return domain.Product{}, fmt.Errorf("Неверный запрос или нет активной приемки")
	}
	logger.Log.Debug().Any("reception id", recepId).Msg("id приемки")
	addedProduct, err := r.insertProduct(tx, product, recepId, *product.PVZId)
	if err != nil {
		return domain.Product{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Product{}, err
	}
	logger.Log.Debug().Any("pvz response", addedProduct).Msg("Успешно добавлен товар")
	return addedProduct, nil
}

func (r *PvzPostgres) DeleteLastProduct(delProd uuid.UUID) error {
	tx, err := r.beginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lastStatus, recepId, err := r.getLastReceptionStatus(tx, delProd)
	if err != nil {
		return err
	}
	if lastStatus == "close" {
		return fmt.Errorf("Неверный запрос, нет активной приемки или нет товаров для удаления")
	}
	logger.Log.Debug().Any("reception id", recepId).Msg("id приемки")
	if err = r.delLastProduct(tx, delProd, recepId); err != nil {
		if errors.Is(err, ErrNoProductsToDelete) {
			return fmt.Errorf("Неверный запрос, нет активной приемки или нет товаров для удаления")
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PvzPostgres) CloseReception(closeProd uuid.UUID) (domain.ProductReception, error) {
	tx, err := r.beginTx()
	if err != nil {
		return domain.ProductReception{}, err
	}
	defer tx.Rollback()
	lastStatus, recepId, err := r.getLastReceptionStatus(tx, closeProd)
	if lastStatus == "close" {
		return domain.ProductReception{}, fmt.Errorf("Неверный запрос, нет активной приемки или нет товаров для удаления")
	}
	if ok, err := r.checkIfAddedProducts(tx, closeProd, recepId); err != nil || !ok {
		logger.Log.Error().Err(err).Msg("Ошибка проверки добавления товаров")
		return domain.ProductReception{}, fmt.Errorf("Неверный запрос, нет активной приемки или нет товаров для удаления")
	}
	logger.Log.Debug().Any("reception id", recepId).Msg("id приемки")
	res, err := r.statusChange(tx, closeProd, recepId)
	if err != nil {
		return domain.ProductReception{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ProductReception{}, err
	}
	return res, nil
}
func (r *PvzPostgres) statusChange(tx *sqlx.Tx, pvzId uuid.UUID, recepId uuid.UUID) (domain.ProductReception, error) {
	var respRecep domain.ProductReception
	query := fmt.Sprintf(`UPDATE %s SET status_reception = 'close' WHERE pvz_id = $1 AND id = $2 RETURNING *`, receptionTable)
	logger.Log.Debug().Str("query", query).Msg("Удаление последнего товара")
	err := tx.QueryRowx(query, pvzId, recepId).Scan(&respRecep.Id, &respRecep.DateReceived, &respRecep.PVZId, &respRecep.Status)
	if err != nil {
		return domain.ProductReception{}, err
	}
	return respRecep, nil
}

func (r *PvzPostgres) checkIfAddedProducts(tx *sqlx.Tx, pvzId uuid.UUID, recepId uuid.UUID) (bool, error) {
	var amount int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE pvz_id = $1 AND reception_id = $2`, productTable)
	logger.Log.Debug().Str("query", query).Msg("Проверка добавления продуктов в приемку")
	err := tx.QueryRowx(query, pvzId, recepId).Scan(&amount)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return amount > 0, nil
}

func (r *PvzPostgres) getLastReceptionStatus(tx *sqlx.Tx, pvzId uuid.UUID) (string, uuid.UUID, error) {
	var status string
	var recepId uuid.UUID
	query := fmt.Sprintf(`SELECT status_reception,id FROM %s WHERE pvz_id = $1 ORDER BY date_received DESC LIMIT 1`, receptionTable)
	logger.Log.Debug().Str("query", query).Msg("Получение последнего статуса приёмки")
	err := tx.QueryRowx(query, pvzId).Scan(&status, &recepId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", uuid.Nil, nil
		}
		return "", uuid.Nil, err
	}
	return status, recepId, nil
}

func (r *PvzPostgres) delLastProduct(tx *sqlx.Tx, pvzId uuid.UUID, recepId uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = (
SELECT id FROM product
  WHERE pvz_id = $1 AND reception_id = $2
  ORDER BY date_received DESC
  LIMIT 1
)`, productTable)
	logger.Log.Debug().Str("query", query).Msg("Удаление последнего товара")
	res, err := tx.Exec(query, pvzId, recepId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoProductsToDelete
	}
	return nil
}

func (r *PvzPostgres) insertReception(tx *sqlx.Tx, recep domain.ProductReception) (domain.ProductReception, error) {
	query := fmt.Sprintf(`INSERT INTO %s (date_received, pvz_id, status_reception) VALUES ($1, $2, $3) RETURNING id, date_received, pvz_id, status_reception`, receptionTable)
	logger.Log.Debug().Str("query", query).Msg("Вставка новой приёмки")
	var res domain.ProductReception
	err := tx.QueryRowx(query, recep.DateReceived, recep.PVZId, recep.Status).
		Scan(&res.Id, &res.DateReceived, &res.PVZId, &res.Status)
	return res, err
}

func (r *PvzPostgres) insertProduct(tx *sqlx.Tx, product domain.Product, recepId uuid.UUID, pvzId uuid.UUID) (domain.Product, error) {
	query := fmt.Sprintf(`INSERT INTO %s (date_received, type_product, reception_id,pvz_id) VALUES ($1, $2, $3,$4) RETURNING id, date_received, type_product, reception_id`, productTable)
	logger.Log.Debug().Str("query", query).Msg("Добавление нового товара")
	var res domain.Product
	err := tx.QueryRowx(query, product.DateReceived, product.Type, recepId, pvzId).
		Scan(&res.Id, &res.DateReceived, &res.Type, &res.ReceptionId)
	return res, err
}

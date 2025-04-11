package repository

import (
	"fmt"

	logger "github.com/bllooop/pvzservice/pkg/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBname   string
	SSLMode  string
}

const (
	userListTable  = "userlist"
	pvzTable       = "pvz"
	receptionTable = "product_reception"
	productTable   = "product"
)

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	logger.Log.Info().Msg("Подключение к базе данных")
	constring := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBname, cfg.SSLMode)
	db, err := sqlx.Open("pgx", constring)
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBname, cfg.SSLMode)
	logger.Log.Debug().Str("conn", connStr).Msg("Обработка подключения к БД")
	if err != nil {
		return nil, err
	}
	return db, nil
}

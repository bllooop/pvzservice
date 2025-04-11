package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bllooop/pvzservice/internal/domain"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

func (r *AuthPostgres) CreateUser(user domain.User) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (email,password,role) VALUES ($1,$2,$3) RETURNING id`, userListTable)
	row := r.db.QueryRowx(query, user.Email, user.Password, user.Role)
	logger.Log.Debug().Str("query", query).Msg("Выполнение запроса регистрации")
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	logger.Log.Debug().Int("id", id).Msg("Успешно зарегестрирован пользователь")
	return id, nil
}

func (r *AuthPostgres) SignUser(email string) (domain.User, error) {
	var user domain.User
	query := fmt.Sprintf(`SELECT id,email,password FROM %s WHERE email=$1`, userListTable)
	res := r.db.QueryRowx(query, email)
	err := res.Scan(&user.Id, &user.Email, &user.Password)
	logger.Log.Debug().Str("query", query).Msg("Выполнение запроса авторизации")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.New("пользователь не найден")
		}
		return domain.User{}, err
	}
	logger.Log.Debug().Any("user", user).Msg("Успешно найден пользователь")
	return user, nil
}

func (r *AuthPostgres) DB() *sqlx.DB {
	return r.db
}

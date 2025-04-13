package repository

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestAuthPostgres_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	r := NewAuthPostgres(sqlxDB)

	tests := []struct {
		name    string
		mock    func()
		input   domain.User
		want    domain.User
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"email", "password"}).AddRow("email", "employee")
				mock.ExpectQuery("INSERT INTO userlist").
					WithArgs("email", "123", "employee").WillReturnRows(rows)
			},
			input: domain.User{
				Email:    "email",
				Password: "123",
				Role:     "employee",
			},
			want: domain.User{
				Email: "email",
				Role:  "employee",
			},
		},
		{
			name: "Пустые поля вводных данных",
			mock: func() {
			},
			input: domain.User{
				Email:    "",
				Password: "123",
				Role:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CreateUser(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthPostgres_SignUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	r := NewAuthPostgres(sqlxDB)

	type args struct {
		username string
	}

	tests := []struct {
		name    string
		mock    func()
		input   args
		want    domain.User
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
					AddRow(userID, "test", "password", "employee")
				mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM %s", userListTable)).
					WithArgs("test").WillReturnRows(rows)
			},
			input: args{"test"},
			want: domain.User{
				Id:       userID,
				Email:    "test",
				Password: "password",
				Role:     "employee",
			},
		},
		{
			name: "Пользователь не найден",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password", "role"})
				mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM %s", userListTable)).
					WithArgs("not").WillReturnRows(rows)
			},
			input:   args{"not"},
			wantErr: true,
		},
		{
			name: "Пустые поля вводных данных",
			mock: func() {
			},
			input:   args{""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.SignUser(tt.input.username)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

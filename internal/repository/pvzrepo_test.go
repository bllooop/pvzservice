package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPvzPostgres_CreatePvz(t *testing.T) {
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
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
	r := NewPvzPostgres(sqlxDB)

	tests := []struct {
		name    string
		mock    func()
		input   domain.PVZ
		want    domain.PVZ
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "registrationDate", "city"}).AddRow(&userID, fixedTime, "Москва")
				mock.ExpectQuery("INSERT INTO pvz").
					WithArgs(&fixedTime, "Москва").WillReturnRows(rows)
			},
			input: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "Москва",
			},
			want: domain.PVZ{
				Id:           &userID,
				DateRegister: &fixedTime,
				City:         "Москва",
			},
		},
		{
			name: "Ошибка БД",
			mock: func() {
				mock.ExpectQuery("INSERT INTO pvz").
					WithArgs(&fixedTime, "Москва").
					WillReturnError(errors.New("ошибка бд"))
			},
			input: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "Москва",
			},
			want:    domain.PVZ{},
			wantErr: true,
		},
		{
			name: "Ошибка Scan",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "registrationDate"}).AddRow(uuid.New(), time.Now())
				mock.ExpectQuery("INSERT INTO pvz").
					WithArgs(&fixedTime, "Москва").
					WillReturnRows(rows)
			},
			input: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "Москва",
			},
			want:    domain.PVZ{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CreatePvz(tt.input)
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

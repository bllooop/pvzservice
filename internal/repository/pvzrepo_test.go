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
				rows := sqlmock.NewRows([]string{"id", "registrationdate", "city"}).AddRow(&userID, fixedTime, "Москва")
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
				rows := sqlmock.NewRows([]string{"id", "registrationdate"}).AddRow(uuid.New(), time.Now())
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

func TestPvzPostgres_getPvz(t *testing.T) {
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
	typ := "электроника"
	stat := "in_progress"
	tests := []struct {
		name    string
		mock    func()
		input   domain.GettingPvzParams
		want    []domain.PvzSummary
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				mock.ExpectBegin()
				pvzRows := sqlmock.NewRows([]string{"id", "registrationdate", "city"}).AddRow(userID, fixedTime, "Москва")
				mock.ExpectQuery("SELECT \\* FROM pvz").WillReturnRows(pvzRows)
				recepRows := sqlmock.NewRows([]string{"id", "date_received", "pvz_id", "status_reception"}).AddRow(userID, fixedTime, userID, stat)
				mock.ExpectQuery("SELECT \\* FROM product_reception").WillReturnRows(recepRows)
				prodRows := sqlmock.NewRows([]string{"id", "date_received", "type_product", "reception_id", "pvz_id"}).AddRow(userID, fixedTime, typ, userID, userID)
				mock.ExpectQuery("SELECT \\* FROM product").WillReturnRows(prodRows)
				mock.ExpectCommit()
			},
			input: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			want: []domain.PvzSummary{
				{
					PvzInfo: domain.PVZ{
						Id:           &userID,
						DateRegister: &fixedTime,
						City:         "Москва",
					},
					ReceptionsInfo: []domain.Receptions{
						{
							ReceptionInfo: domain.ProductReception{
								Id:           &userID,
								PVZId:        &userID,
								DateReceived: &fixedTime,
								Status:       &stat,
							},
							ProductInfo: []domain.Product{
								{
									Id:           &userID,
									ReceptionId:  &userID,
									Type:         typ,
									DateReceived: &fixedTime,
									PVZId:        &userID,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ошибка при запросе PVZ",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT \\* FROM pvz").
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			input: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			want:    []domain.PvzSummary{},
			wantErr: true,
		},
		{
			name: "Ошибка при запросе приемок",
			mock: func() {
				mock.ExpectBegin()
				pvzRows := sqlmock.NewRows([]string{"id", "registrationdate", "city"}).AddRow(userID, fixedTime, "Москва")
				mock.ExpectQuery("SELECT \\* FROM pvz").
					WillReturnRows(pvzRows)

				mock.ExpectQuery("SELECT \\* FROM product_reception").
					WillReturnError(errors.New("reception error"))
				mock.ExpectRollback()

			},
			input: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			want:    []domain.PvzSummary{},
			wantErr: true,
		},
		{
			name: "Ошибка при запросе товаров",
			mock: func() {
				mock.ExpectBegin()
				pvzRows := sqlmock.NewRows([]string{"id", "registrationdate", "city"}).
					AddRow(userID, fixedTime, "Москва")
				mock.ExpectQuery("SELECT \\* FROM pvz").
					WillReturnRows(pvzRows)

				recepRows := sqlmock.NewRows([]string{"id", "date_received", "pvz_id", "status_reception"}).AddRow(userID, fixedTime, userID, stat)
				mock.ExpectQuery("SELECT \\* FROM product_reception").
					WillReturnRows(recepRows)

				mock.ExpectQuery("SELECT \\* FROM product").
					WillReturnError(errors.New("product error"))
				mock.ExpectRollback()
			},
			input: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			want:    []domain.PvzSummary{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.GetPvz(tt.input)
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

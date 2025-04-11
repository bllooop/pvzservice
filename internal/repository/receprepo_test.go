package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPvzPostgres_CreateRecep(t *testing.T) {
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
	stat := "in_progress"

	tests := []struct {
		name    string
		mock    func()
		input   domain.ProductReception
		want    domain.ProductReception
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnError(sql.ErrNoRows)
				rows := sqlmock.NewRows([]string{"id", "date_received", "pvz_id", "status_reception"}).AddRow(userID, fixedTime, userID, stat)
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s", receptionTable)).
					WithArgs(&fixedTime, &userID, &stat).WillReturnRows(rows)
				mock.ExpectCommit()
			},
			input: domain.ProductReception{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
			want: domain.ProductReception{
				Id:           &userID,
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
		},
		{
			name: "Ошибка БД",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s", receptionTable)).
					WithArgs(&fixedTime, &userID, &stat).WillReturnError(errors.New("ошибка бд"))
				mock.ExpectRollback()
			},
			input: domain.ProductReception{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
			want:    domain.ProductReception{},
			wantErr: true,
		},
		{
			name: "Ошибка Scan",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnError(sql.ErrNoRows)
				rows := sqlmock.NewRows([]string{"id", "date_received"}).AddRow(uuid.New(), time.Now())
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s", receptionTable)).
					WithArgs(&fixedTime, &userID, &stat).
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
			input: domain.ProductReception{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
			want:    domain.ProductReception{},
			wantErr: true,
		},
		{
			name: "Есть незакрытая приемка",
			mock: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows)
				mock.ExpectRollback()
			},
			input: domain.ProductReception{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
			want:    domain.ProductReception{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CreateRecep(tt.input)
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

func TestPvzPostgres_AddProdToRecep(t *testing.T) {
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

	tests := []struct {
		name    string
		mock    func()
		input   domain.Product
		want    domain.Product
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				rows := sqlmock.NewRows([]string{"id", "date_received", "type_product", "reception_id"}).AddRow(userID, fixedTime, typ, userID)
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s (.+)", productTable)).
					WithArgs(&fixedTime, &typ, &userID, &userID).WillReturnRows(rows)
				mock.ExpectCommit()
			},
			input: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Type:         typ,
			},
			want: domain.Product{
				Id:           &userID,
				DateReceived: &fixedTime,
				ReceptionId:  &userID,
				Type:         typ,
			},
		},
		{
			name: "Ошибка БД",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s (.+)", productTable)).
					WithArgs(&fixedTime, &typ, &userID, &userID).WillReturnError(errors.New("ошибка бд"))
				mock.ExpectRollback()
			},
			input: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Type:         typ,
			},
			want:    domain.Product{},
			wantErr: true,
		},
		{
			name: "Ошибка Scan",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				rows := sqlmock.NewRows([]string{"id", "date_received"}).AddRow(uuid.New(), time.Now())
				mock.ExpectQuery(fmt.Sprintf("INSERT INTO %s (.+)", productTable)).
					WithArgs(&fixedTime, &typ, &userID, &userID).
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
			input: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Type:         typ,
			},
			want:    domain.Product{},
			wantErr: true,
		},
		{
			name: "Нет активной приемки",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			input: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Type:         typ,
			},
			want:    domain.Product{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.AddProdToRecep(tt.input)
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

func TestPvzPostgres_deleteLast(t *testing.T) {
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
		input   uuid.UUID
		want    error
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				//rows := sqlmock.NewRows([]string{"pvz_id", "reception_id"}).AddRow(userID, userID)
				mock.ExpectExec(fmt.Sprintf("DELETE FROM %s ", productTable)).
					WithArgs(&userID, &userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			input: userID,
			want:  nil,
		},
		{
			name: "Ошибка БД",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				mock.ExpectExec(fmt.Sprintf("DELETE FROM %s ", productTable)).
					WithArgs(&userID, &userID).WillReturnError(errors.New("ошибка бд"))
				mock.ExpectRollback()
			},
			input:   userID,
			want:    err,
			wantErr: true,
		},
		/*{
			name: "Ошибка Scan",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				//rows := sqlmock.NewRows([]string{"id", "date_received"}).AddRow(uuid.New(), time.Now())
				mock.ExpectExec(fmt.Sprintf("DELETE FROM %s ", productTable)).
					WithArgs(&userID, &userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectRollback()
			},
			input:   userID,
			want:    err,
			wantErr: true,
		},*/
		{
			name: "Нет активной приемки",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			input:   userID,
			want:    err,
			wantErr: true,
		},
		{
			name: "Нет товаров для удаления",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).
					WillReturnRows(sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("open", userID))

				mock.ExpectExec(fmt.Sprintf("DELETE FROM %s (.+)", productTable)).
					WithArgs(&userID, &userID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectRollback()
			},
			input:   userID,
			want:    fmt.Errorf("Неверный запрос, нет активной приемки или нет товаров для удаления"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := r.DeleteLastProduct(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPvzPostgres_closeLast(t *testing.T) {
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
	stat := "close"

	tests := []struct {
		name    string
		mock    func()
		input   uuid.UUID
		want    domain.ProductReception
		wantErr bool
	}{
		{
			name: "Ok",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				mock.ExpectQuery(fmt.Sprintf(`COUNT\(\*\) FROM %s (.+)`, productTable)).
					WithArgs(&userID, &userID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				rows := sqlmock.NewRows([]string{"id", "date_received", "pvz_id", "status_reception"}).AddRow(userID, fixedTime, userID, stat)
				mock.ExpectQuery(fmt.Sprintf("UPDATE %s SET (.+)", receptionTable)).
					WithArgs(&userID, &userID).WillReturnRows(rows)
				mock.ExpectCommit()
			},
			input: userID,
			want: domain.ProductReception{
				Id:           &userID,
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Status:       &stat,
			},
		},
		{
			name: "Ошибка БД",
			mock: func() {
				mock.ExpectBegin()
				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(&userID).WillReturnRows(rows2)
				mock.ExpectQuery(fmt.Sprintf(`COUNT\(\*\) FROM %s (.+)`, productTable)).
					WithArgs(&userID, &userID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				mock.ExpectQuery(fmt.Sprintf("UPDATE %s SET (.+)", receptionTable)).
					WithArgs(&userID, &userID).WillReturnError(errors.New("ошибка бд"))
				mock.ExpectRollback()
			},
			input:   userID,
			want:    domain.ProductReception{},
			wantErr: true,
		},
		{
			name: "Приемка уже закрыта",
			mock: func() {
				mock.ExpectBegin()

				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("close", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(userID).
					WillReturnRows(rows2)
				mock.ExpectRollback()
			},
			input:   userID,
			want:    domain.ProductReception{},
			wantErr: true,
		},
		{
			name: "Ошибка проверки наличия товаров",
			mock: func() {
				mock.ExpectBegin()

				rows2 := sqlmock.NewRows([]string{"status_reception", "id"}).AddRow("in_progress", userID)
				mock.ExpectQuery(fmt.Sprintf("SELECT status_reception,id FROM %s (.+)", receptionTable)).
					WithArgs(userID).
					WillReturnRows(rows2)

				mock.ExpectQuery(fmt.Sprintf(`COUNT\(\*\) FROM %s (.+)`, productTable)).
					WithArgs(userID, userID).
					WillReturnError(fmt.Errorf("Ошибка проверки добавления товаров"))

				mock.ExpectRollback()
			},
			input:   userID,
			want:    domain.ProductReception{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.CloseReception(tt.input)
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

package domain

import (
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	Id           *uuid.UUID `json:"id" db:"id"`
	DateRegister *time.Time `json:"registrationDate,omitempty" db:"registrationdate"`
	City         string     `json:"city" binding:"required,oneof=Москва Санкт-Петербург Казань"`
}

type ProductReception struct {
	Id           *uuid.UUID `json:"id" db:"id"`
	DateReceived *time.Time `json:"dateTime,omitempty" db:"date_received"`
	PVZId        *uuid.UUID `json:"pvzId" db:"pvz_id"`
	Status       *string    `json:"status,omitempty" db:"status_reception"`
}

type Product struct {
	Id           *uuid.UUID `json:"id" db:"id"`
	DateReceived *time.Time `json:"dateTime,omitempty" db:"date_received"`
	Type         string     `json:"type" db:"type_product" binding:"required,oneof=электроника одежда обувь"`
	ReceptionId  *uuid.UUID `json:"receptionId" db:"reception_id"`
	PVZId        *uuid.UUID `json:"pvzId,omitempty" db:"pvz_id"`
}

type PvzSummary struct {
	PvzInfo        PVZ          `json:"pvz"`
	ReceptionsInfo []Receptions `json:"receptions"`
}

type Receptions struct {
	ReceptionInfo ProductReception `json:"reception"`
	ProductInfo   []Product        `json:"products"`
}
type GettingPvzParams struct {
	Start time.Time
	End   time.Time
	Page  int
	Limit int
}

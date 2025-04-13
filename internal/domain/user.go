package domain

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID `json:"-" db:"id"`
	Email    string    `json:"email"`
	Password string    `json:"password,omitempty"`
	Role     string    `json:"role" binding:"required,oneof=employee moderator"`
}

type SignInInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type DummyLogin struct {
	Role string `json:"role" binding:"required,oneof=employee moderator"`
}

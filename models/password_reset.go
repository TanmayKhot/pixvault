package models

import (
	"database/sql"
	"fmt"
	"time"
)

type PasswordReset struct {
	ID     int
	UserID int
	// Token is set only when a PasswordReset is being created
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

const (
	// DefaultResetDuration is the default time that a PasswordReset is valid for
	DefaultResetDuration = 1 * time.Hour
)

type PasswordResetService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each password reset token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be
	// used.
	// This logic is similar to sessions code
	BytesPerToken int

	// Duration is the amount of time that a PasswordReset is valid for
	Duration time.Duration
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	// This method will create the password reset token
	return nil, fmt.Errorf("TODO: Implement PasswordResetService.Create")
}

func (service *PasswordResetService) Consume(token string) (*User, error) {
	// This method will consume the token and return a user associated with it
	// If the token is invalid then it will return an error
	return nil, fmt.Errorf("TODO: Implement PasswordResetService.Consume")
}

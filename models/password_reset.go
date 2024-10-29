package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/TanmayKhot/pixvault/rand"
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

func (service *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	// This method will create the password reset token

	// 1. Verify if there exists a user with given email
	email = strings.ToLower(email)
	var userID int
	row := service.DB.QueryRow(`
	SELECT id FROM users WHERE email = $1;`, email)
	err := row.Scan(&userID)
	if err != nil {
		//TODO: Consider returning a specific error when the user doesnt exist
		// For eg: It could either be db connection issue OR the user doesn't exist in db
		return nil, fmt.Errorf("Create: %w", err)
	}
	// Build the token
	bytesPerToken := service.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = service.DB.QueryRow(`
	INSERT INTO password_resets (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3) ON CONFLICT (user_id) DO 
	UPDATE
	SET token_hash =$2, expires_at= $3
	RETURNING id;`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return &pwReset, nil
}

func (service *PasswordResetService) Consume(token string) (*User, error) {
	// 1. Validate the token used for pw reset and check its expiry
	// 2. If it is a valid token, update the user info (password)
	// 3. Delete the token so that it cannot be used again

	tokenHash := service.hash(token)
	var user User
	var pwReset PasswordReset
	row := service.DB.QueryRow(`
	SELECT 
		pr.id,
		pr.expires_at,
		u.id,
		u.email,
		u.password_hash
	FROM password_resets pr
	JOIN users u ON u.id = pr.user_id
	WHERE pr.token_hash = $1;`, tokenHash)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt,
		&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		return nil, fmt.Errorf("Consume: %w", err)
	}

	// Validate the token expiry
	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("Token expired: %w", err)
	}

	err = service.delete(user.ID)
	if err != nil {
		return nil, fmt.Errorf("comsume: %w", err)
	}

	return &user, nil
}

func (service *PasswordResetService) delete(id int) error {
	_, err := service.DB.Exec(`
	DELETE FROM password_resets
	WHERE user_id = $1;`, id)

	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

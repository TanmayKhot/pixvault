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

type EmailSignin struct {
	ID        int
	UserID    int
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type EmailSigninService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each password reset token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be
	// used.
	// This logic is similar to sessions code
	BytesPerToken int

	//Duration is the amount of time the Email Signin token is valid for
	Duration time.Duration
}

func (service *EmailSigninService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (service *EmailSigninService) Create(email string) (*EmailSignin, error) {
	// This method will create emailSignin token

	// 1. Verify if there exists a user with given email
	email = strings.ToLower(email)
	var userID int

	row := service.DB.QueryRow(`
	SELECT id FROM users WHERE email = $1;`, email)
	err := row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}

	bytesPerToken := service.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}

	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	emailSignin := EmailSignin{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = service.DB.QueryRow(`
	INSERT INTO email_signin (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
	UPDATE 
	SET token_hash=$2, expires_at=$3
	RETURNING id;`, emailSignin.UserID, emailSignin.TokenHash, emailSignin.ExpiresAt)

	err = row.Scan(&emailSignin.ID)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return &emailSignin, nil
}

func (service *EmailSigninService) delete(id int) error {
	_, err := service.DB.Exec(`
	DELETE FROM email_signin
	WHERE user_id = $1;`, id)

	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	return nil
}

func (service *EmailSigninService) Consume(token string) (*User, error) {
	tokenHash := service.hash(token)
	var user User
	var emailSignin EmailSignin
	row := service.DB.QueryRow(`
	SELECT 
		em.id,
		em.expires_at,
		u.id,
		u.email
	FROM email_signin em
	JOIN users u ON u.id = em.user_id
	WHERE em.token_hash = $1`, tokenHash)
	err := row.Scan(&emailSignin.ID, &emailSignin.ExpiresAt, &user.ID, &user.Email)

	if err != nil {
		return nil, fmt.Errorf("Consume: %w", err)
	}

	if time.Now().After(emailSignin.ExpiresAt) {
		return nil, fmt.Errorf("Token expired: %w", err)
	}

	err = service.delete(user.ID)
	if err != nil {
		return nil, fmt.Errorf("Consume: %w", err)
	}

	return &user, nil
}

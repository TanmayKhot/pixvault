package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/TanmayKhot/pixvault/rand"
)

/*
Functioning
1. We will first hash the session token passed into the function.
2. Next, we will query for a session with that token hash.
3. As long as a user is found, we will use the user ID from the result to query for a user with that ID.
4. Finally, we will return the user associated with the session token.
*/

type Session struct {
	ID     int
	UserID int
	// Token is only set when creating a new session. When looking up a session
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token.
	Token     string
	TokenHash string
}

const (
	MinBytesPerToken = 32
)

type SessionService struct {
	DB            *sql.DB
	BytesPerToken int
}

// Token is only set when creating a new session. When looking up a session
// this will be left empty, as we only store the hash of a session token
// in our database and we cannot reverse it into a raw token.
func (ss *SessionService) Create(userId int) (*Session, error) {
	// Generate a token
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	// Create a session
	session := Session{
		UserID:    userId,
		Token:     token,
		TokenHash: ss.hash(token),
	}

	// Store the session in our DB
	row := ss.DB.QueryRow(
		`UPDATE sessions 
		SET token_hash = $2
		WHERE user_id = $1
		RETURNING id;`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)
	if err == sql.ErrNoRows {
		// If no rows exist then we need to create a new session for the user
		row = ss.DB.QueryRow(`
		INSERT INTO sessions (user_id, token_hash)
		VALUES ($1, $2)
		RETURNING id;`, session.UserID, session.TokenHash)
		err = row.Scan(&session.ID)
	}
	// If creating a new session fails then we check for other errors
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return &session, nil
}

// Returns a user based on the session token extracted from the cookie
func (ss *SessionService) User(token string) (*User, error) {
	/*
		Operation logic
		1. Hash the session token
		2. Query the database for the session with that token
		3. Get the userID from session
		4. Query the database for that user
		5. Return the user
	*/

	var user User

	// 1. Hash the session token
	tokenHash := ss.hash(token)

	// 2. Query for the session with that hash
	row := ss.DB.QueryRow(`
	SELECT user_ID
	FROM sessions
	WHERE token_hash = $1;`, tokenHash)
	err := row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	// 3. Return the user based on userID
	row = ss.DB.QueryRow(`
	SELECT email, password_hash
	FROM users
	WHERE id = $1`, user.ID)
	err = row.Scan(&user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	return &user, nil
}

// Delete a session
func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	_, err := ss.DB.Exec(`
	DELETE FROM sessions
	WHERE token_hash = $1;`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

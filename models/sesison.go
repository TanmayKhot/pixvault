package models

import "database/sql"

type Session struct {
	ID     int
	UserID int
	// Token is only set when creating a new session. When looking up a session
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token.
	Token     string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
}

// Token is only set when creating a new session. When looking up a session
// this will be left empty, as we only store the hash of a session token
// in our database and we cannot reverse it into a raw token.
func (ss *SessionService) Create(userId int) (*Session, error) {
	return nil, nil
}

// Returns a user based on the session token extracted from the cookie
func (ss *SessionService) User(token string) (*User, error) {
	return nil, nil
}

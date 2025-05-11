package models

import "errors"

var (
	ErrNotFound           = errors.New("Resource not found")
	ErrInvalidID          = errors.New("ID provided was invalid")
	ErrEmailTaken         = errors.New("Email address is already in use")
	ErrInvalidCredentials = errors.New("Invalid email or password")
)

package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Returns a byte slice containing n random numbers
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	nRead, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if nRead < n {
		return nil, fmt.Errorf("bytes: 	didn't read enough bytes")
	}
	return b, nil
}

// Return a random string of length n
func String(n int) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Generate session token to track user sessions
const SessionTokenBytes = 32

func SessionToken() (string, error) {
	return String(SessionTokenBytes)
}

package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateSecureToken creates a cryptographically secure random token string.
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding to avoid special characters.
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashToken creates a SHA-256 hash of a token.
func HashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

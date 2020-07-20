package akhttpd

import "context"

// PublicKeyRepository represents an object that can fetch public keys
type PublicKeyRepository interface {
	// GetKey gets a key for a given username
	GetKey(context context.Context, username string) (keys []string, err error)

	// IsNotFound checks if a given error represents a not found error
	IsNotFound(err error) bool
}

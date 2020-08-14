package akhttpd

import (
	"context"

	"golang.org/x/crypto/ssh"
)

// KeyRepository is an object that can fetch pulic keys from a provided repository
type KeyRepository interface {
	// GetKeys fetches the keys for the provided username
	// When fetching keys fails, should return a value of type UserNotFoundError.
	GetKeys(context context.Context, username string) (keys []ssh.PublicKey, err error)
}

// UserNotFoundError is an error that indicates a user was not found
// This implements the causer interface of the "github.com/pkg/errors" package.
type UserNotFoundError struct {
	error
}

// Cause returns the underlying error of this UserNotFoundError
func (u UserNotFoundError) Cause() error {
	return u.error
}

// Unwrap provides compatibility with go 1.13+ errors
func (u UserNotFoundError) Unwrap() error {
	return u.error
}

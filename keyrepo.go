package akhttpd

import (
	"context"

	"golang.org/x/crypto/ssh"
)

// KeyRepository is an object that can fetch ssh keys for a given username from a remote source.
// Any implementation is assumed safe for concurrent access and may internally cache responses.
type KeyRepository interface {
	// GetKeys resolves and returns the keys for the provided username.
	//
	// When this function determines that a user does not exist, it returns an error of type UserNotFoundError.
	// It may return other error types for undefined errors
	GetKeys(context context.Context, username string) (keys []ssh.PublicKey, err error)
}

// UserNotFoundError indicates that a KeyRepository was unable to find the provided user and is thus unable to return keys for it.
//
// This type implements github.com/pkg/errors.Causer and go 1.13+ errors.
type UserNotFoundError struct {
	error
}

// Cause returns the error that caused this error.
func (u UserNotFoundError) Cause() error {
	return u.error
}

// Unwrap unwraps this error
func (u UserNotFoundError) Unwrap() error {
	return u.error
}

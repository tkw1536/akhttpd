package repo

import (
	"context"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Blocklisted represents a KeyRepository that blocks a list of user for legal reasons
type Blocklisted struct {
	Repository KeyRepository

	Blocked []string // set of case-insensitive usernames that are blocked
}

// GetKeys resolves and returns the keys for the provided username.
func (b *Blocklisted) GetKeys(context context.Context, username string) (string, []ssh.PublicKey, error) {
	// check if the user is blacklisted
	for _, user := range b.Blocked {
		if strings.EqualFold(user, username) {
			return "", nil, UserNotAvailableError{user: username}
		}
	}

	// then call the normal function
	return b.Repository.GetKeys(context, username)
}

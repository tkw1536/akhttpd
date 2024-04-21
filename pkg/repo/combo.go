package repo

import (
	"context"

	"golang.org/x/crypto/ssh"
)

// Combo combines an array of KeyRepositories by trying each in order.
// Keys for a specific user are always returned from a specific repository, and never combined.
type Combo []KeyRepository

// GetKeys resolves and returns the keys for the provided username.
// When a user cannot be found, returns the appropriate error of the last user
func (c Combo) GetKeys(context context.Context, username string) (source string, keys []ssh.PublicKey, err error) {
	for _, r := range c {
		source, keys, err = r.GetKeys(context, username)

		// upon encountering a regular (not not-found error) or nil error, we can immediately return
		if _, isNotFound := err.(UserNotFoundError); err == nil || !isNotFound {
			break
		}
	}

	// return the last error!
	return
}

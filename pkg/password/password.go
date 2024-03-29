package password

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const PasswordCharSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var passwordCharCount = big.NewInt(int64(len(PasswordCharSet)))

// Password returns a randomly generated string with the provided length.
// It consists of alphanumeric characters only.
//
// When an error occurs, it is guaranteed to return "", err.
// [rand.Reader] is used as the source of randomness.
func Password(length int) (string, error) {
	if length < 0 {
		panic("length < 0")
	}

	// create a buffer to write the string to!
	var password strings.Builder
	password.Grow(length)

	for i := 0; i < length; i++ {

		// grab a random bIndex!
		bIndex, err := rand.Int(rand.Reader, passwordCharCount)
		if err != nil {
			return "", err
		}

		// and use that index!
		index := int(bIndex.Int64())
		if err := password.WriteByte(PasswordCharSet[index]); err != nil {
			return "", err
		}
	}

	// return the password!
	return password.String(), nil
}

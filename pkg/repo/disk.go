package repo

import (
	"context"
	"io/fs"
	"os"

	"golang.org/x/crypto/ssh"
)

// Disk reads keys from files named on disk
type Disk struct {
	FS fs.FS
}

func (d Disk) path(username string) (path string) {
	return username
}

func (d Disk) GetKeys(context context.Context, username string) (source string, keys []ssh.PublicKey, err error) {
	bytes, err := fs.ReadFile(d.FS, d.path(username))
	if os.IsNotExist(err) {
		return "", nil, UserNotFoundError{error: err}
	}
	if err != nil {
		return "", nil, err
	}
	return "disk", parseKeys(bytes), nil
}

func parseKeys(in []byte) (keys []ssh.PublicKey) {
	var err error
	var key ssh.PublicKey
	for {
		key, _, _, in, err = ssh.ParseAuthorizedKey(in)
		if err != nil {
			break
		}
		keys = append(keys, key)
	}
	return
}

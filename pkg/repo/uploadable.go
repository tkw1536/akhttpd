package repo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/tkw1536/akhttpd/pkg/password"
	"github.com/tkw1536/akhttpd/pkg/wshandler"

	"golang.org/x/crypto/ssh"

	_ "embed"
)

// spellchecker:words akhttpd wshandler userkeys

// UploadableKeys is an object that allows callers to upload keys to the server temporarily.
type UploadableKeys struct {
	Prefix  string // Prefix is the prefix for new users
	counter uint64 // internal counter for usernames

	AuthPassword, AuthUser string

	WriteSuffix func(w io.Writer) error

	lock sync.RWMutex
	data map[string][]ssh.PublicKey
}

var errUserKeysNotConfigured = UserNotFoundError{errors.New("User is not configured in UserKeys")}

// GetKeys fetches keys from GitHub for the provided username.
//
// If this function determines that a user does not exist, returns a UserNotFoundError.
func (uk *UploadableKeys) GetKeys(context context.Context, username string) (string, []ssh.PublicKey, error) {
	uk.lock.RLock()
	defer uk.lock.RUnlock()

	// check if we have the keys
	keys, ok := uk.data[username]
	if !ok {
		return "", nil, errUserKeysNotConfigured
	}

	return "userkeys", keys, nil
}

// Register registers a new set of keys from the user.
// The delete function will delete the user from the cache.
func (uk *UploadableKeys) Register(keys ...ssh.PublicKey) (username string, cleanup func()) {
	uk.lock.Lock()
	defer uk.lock.Unlock()

	// create a username that does not yet exist
	for {
		username = uk.username()
		if _, ok := uk.data[username]; ok {
			continue
		}
		break
	}

	if uk.data == nil {
		uk.data = make(map[string][]ssh.PublicKey)
	}
	uk.data[username] = keys

	return username, func() {
		uk.lock.Lock()
		defer uk.lock.Unlock()

		delete(uk.data, username)
	}
}

// username generates a new username
func (uk *UploadableKeys) username() string {
	hash, err := password.Password(10)
	if err != nil { // fallback to a counter-based approach
		return fmt.Sprintf("%s%d", uk.Prefix, atomic.AddUint64(&uk.counter, 1))
	}
	return uk.Prefix + hash
}

//go:embed uploadable.min.html
var uploadableHTML []byte

var authenticateHeader = `Basic realm="akhttpd UserUpload"`
var authorizedResponse = []byte("Unauthorized")

func (uk *UploadableKeys) auth(w http.ResponseWriter, r *http.Request) bool {
	// no auth required!
	if uk.AuthPassword == "" && uk.AuthUser == "" {
		return true
	}

	// check that the authentication matches
	user, pass, ok := r.BasicAuth()
	if ok && user == uk.AuthUser && pass == uk.AuthPassword {
		return true
	}

	// return an unauthorized response
	w.Header().Add("WWW-Authenticate", authenticateHeader)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(authorizedResponse)
	return false
}

func (uk *UploadableKeys) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !uk.auth(w, r) {
		return
	}

	// if an upgrade to the websocket was requested, serve a websocket!
	if r.Header.Get("Upgrade") == "websocket" {
		wshandler.Handle(w, r, func(messenger wshandler.WebSocket) {
			key, ok := messenger.Read() // wait for any kind of message
			if !ok {
				return
			}

			// read a private key from the connection!
			pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				return
			}

			// register the key
			username, cleanup := uk.Register(pk)
			defer cleanup()

			// write the username back!
			if !messenger.Write(username) {
				return
			}

			messenger.Wait()
		})
		return
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(uploadableHTML)
	if uk.WriteSuffix != nil {
		uk.WriteSuffix(w)
	}
}

package repo

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"

	_ "embed"

	"github.com/tkw1536/pkglib/lazy"
	"github.com/tkw1536/pkglib/password"
	"github.com/tkw1536/pkglib/websocketx"
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

	server lazy.Lazy[*websocketx.Server]
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
	hash, err := password.Generate(rand.Reader, 10, password.DefaultCharSet)
	if err != nil { // fallback to a counter-based approach
		return fmt.Sprintf("%s%d", uk.Prefix, atomic.AddUint64(&uk.counter, 1))
	}
	return uk.Prefix + hash
}

func (uk *UploadableKeys) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !uk.auth(w, r) {
		return
	}

	uk.server.Get(func() *websocketx.Server {
		return &websocketx.Server{
			Handler:  uk.handleWS,
			Fallback: http.HandlerFunc(uk.handleHTTP),
		}
	}).ServeHTTP(w, r)
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

func (uk *UploadableKeys) handleWS(conn *websocketx.Connection) {
	key, ok := <-conn.Read()
	if !ok {
		return
	}

	// read a private key from the connection!
	pk, _, _, _, err := ssh.ParseAuthorizedKey(key.Body)
	if err != nil {
		return
	}

	// register the key
	username, cleanup := uk.Register(pk)
	defer cleanup()

	// Write the username back
	conn.WriteText(username)

	// and wait for the connection to be closed
	// by the client
	<-conn.Context().Done()
}

func (uk *UploadableKeys) handleHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(uploadableHTML)
	if uk.WriteSuffix != nil {
		uk.WriteSuffix(w)
	}
}

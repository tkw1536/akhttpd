package akhttpd

import (
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/mpolden/echoip/useragent"
)

// Formatter is an object that can write ssh keys to an http.ResponseWriter.
type Formatter interface {
	// WriteTo writes the ssh keys, which are associated with the given user, into w.
	// Returns the number of bytes written and an error.
	WriteTo(handler Handler, username string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error)
}

// fmtContext is an object that is internally used to format values for the templates
type fmtContext struct {
	User string
	Time time.Time
	Keys []string
}

// newFmtContext returns a new format context
func newFmtContext(username string, keys []ssh.PublicKey) (ctx fmtContext, err error) {
	ctx.User = username
	ctx.Time = time.Now().UTC()
	ctx.Keys = make([]string, 0, len(keys))

	// format all the keys
	for _, k := range keys {
		key := ssh.MarshalAuthorizedKey(k)
		ctx.Keys = append(ctx.Keys, string(key))
	}
	return
}

// MagicFormatter calls either FormatterAuthorizedKeys or FormatterHTML
type MagicFormatter struct {
	html           FormatterHTML
	authorizedKeys FormatterAuthorizedKeys
}

func (m MagicFormatter) WriteTo(handler Handler, username string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	if m.isCliRequest(r) {
		return m.authorizedKeys.WriteTo(handler, username, keys, r, w)
	}
	return m.html.WriteTo(handler, username, keys, r, w)
}

func (MagicFormatter) isCliRequest(r *http.Request) bool {
	agent := useragent.Parse(r.UserAgent())
	switch agent.Product {
	case "curl", "HTTPie", "httpie-go", "Wget", "fetch libfetch", "Go", "Go-http-client", "ddclient", "Mikrotik", "xh":
		return true
	default:
		return false
	}
}

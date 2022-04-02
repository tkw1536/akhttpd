package format

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
	WriteTo(username, source string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error)
}

// fmtContext is an object that is internally used to format values for the templates
type fmtContext struct {
	User   string
	Source string
	Time   time.Time
	Keys   []string
}

// newFmtContext returns a new format context
func newFmtContext(username, source string, keys []ssh.PublicKey) (ctx fmtContext, err error) {
	ctx.User = username
	ctx.Source = source
	ctx.Time = time.Now().UTC()
	ctx.Keys = make([]string, 0, len(keys))

	// format all the keys
	for _, k := range keys {
		key := ssh.MarshalAuthorizedKey(k)
		ctx.Keys = append(ctx.Keys, string(key))
	}
	return
}

// Magic calls either FormatterAuthorizedKeys or FormatterHTML
type Magic struct {
	HTML           HTML
	AuthorizedKeys AuthorizedKeys
}

func (m Magic) WriteTo(username, source string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	if m.isCliRequest(r) {
		return m.AuthorizedKeys.WriteTo(username, source, keys, r, w)
	}
	return m.HTML.WriteTo(username, source, keys, r, w)
}

func (Magic) isCliRequest(r *http.Request) bool {
	agent := useragent.Parse(r.UserAgent())
	switch agent.Product {
	case "curl", "HTTPie", "httpie-go", "Wget", "fetch libfetch", "Go", "Go-http-client", "ddclient", "Mikrotik", "xh":
		return true
	default:
		return false
	}
}

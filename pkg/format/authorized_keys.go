package format

import (
	"net/http"
	"text/template"

	_ "embed"

	"github.com/tkw1536/akhttpd/pkg/count"
	"golang.org/x/crypto/ssh"
)

// spellchecker:words akhttpd

// AuthorizedKeys is a struct that formats ssh keys as an authorized_keys file.
// It implements Formatter.
type AuthorizedKeys struct{}

//go:embed authorized_keys.tpl
var tplAuthorizedKeys string
var fmtAuthorizedKeys = template.Must(template.New("authorized_keys").Parse(tplAuthorizedKeys))

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted in authorized_keys format and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (AuthorizedKeys) WriteTo(username, source string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, source, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys\"")
	headers.Add("Content-Type", "text/plain")

	return count.Count(w, func(cw *count.Writer) error {
		return fmtAuthorizedKeys.Execute(cw, ctx)
	})
}

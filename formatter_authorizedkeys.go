package akhttpd

import (
	"net/http"
	"text/template"

	_ "embed"

	"golang.org/x/crypto/ssh"
)

// FormatterAuthorizedKeys is a zero-size struct that formats ssh keys as an authorized_keys file.
// It implements Formatter.
type FormatterAuthorizedKeys struct{}

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted in authorized_keys format and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (FormatterAuthorizedKeys) WriteTo(handler Handler, username string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys\"")
	headers.Add("Content-Type", "text/plain")

	ww := &CountWriter{Writer: w}
	return ww.StateWith(fmtAuthorizedKeys.Execute(ww, ctx))
}

//go:embed resources/templates/authorized_keys.tpl
var tplAuthorizedKeys string
var fmtAuthorizedKeys = template.Must(template.New("authorized_keys").Parse(tplAuthorizedKeys))

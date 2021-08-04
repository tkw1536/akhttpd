package akhttpd

import (
	"net/http"
	"text/template"

	_ "embed"

	"golang.org/x/crypto/ssh"
)

// FormatterShellScript is a zero-size struct that formats ssh keys as a shell script updating an authorized_keys file.
// It implements Formatter.
type FormatterShellScript struct{}

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted as a shell script that updates or creates the file '.ssh/authorized_keys' and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (FormatterShellScript) WriteTo(handler Handler, username string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys.sh\"")
	headers.Add("Content-Type", "text/x-shellscript")

	ww := &CountWriter{Writer: w}
	return ww.StateWith(fmtShellTemplate.Execute(ww, ctx))
}

//go:embed resources/templates/authorized_keys.sh.tpl
var tplShellTemplate string
var fmtShellTemplate = template.Must(template.New("authorized_keys.sh").Parse(tplShellTemplate))

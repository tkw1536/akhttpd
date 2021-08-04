package akhttpd

import (
	"net/http"
	"text/template"

	_ "embed"

	"golang.org/x/crypto/ssh"
)

// FormatterShellScript is a zero-size struct that formats ssh keys as a user-facing html page.
// It implements Formatter.
type FormatterHTML struct{}

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted as a shell script that updates or creates the file '.ssh/authorized_keys' and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (FormatterHTML) WriteTo(handler Handler, username string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Type", "text/html")

	ww := &CountWriter{Writer: w}

	err = fmtHTMLTemplate.Execute(ww, ctx)
	if err == nil {
		err = handler.writeSuffix(ww)
	}
	return ww.StateWith(err)
}

//go:embed resources/templates/authorized_keys.min.html.tpl
var tplHTMLTemplate string
var fmtHTMLTemplate = template.Must(template.New("authorized_keys.html").Parse(tplHTMLTemplate))

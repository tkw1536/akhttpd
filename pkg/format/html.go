package format

import (
	"io"
	"net/http"
	"text/template"

	_ "embed"

	"github.com/tkw1536/akhttpd/pkg/count"
	"golang.org/x/crypto/ssh"
)

// FormatterShellScript is a zero-size struct that formats ssh keys as a user-facing html page.
// It implements Formatter.
type HTML struct {
	// Suffix is called to write a suffix to the html response
	Suffix func(w io.Writer) error
}

//go:embed html.min.tpl
var tplHTMLTemplate string
var fmtHTMLTemplate = template.Must(template.New("authorized_keys.html").Parse(tplHTMLTemplate))

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted as a shell script that updates or creates the file '.ssh/authorized_keys' and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (h HTML) WriteTo(username, source string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, source, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Type", "text/html")

	return count.Count(w, func(cw *count.Writer) error {
		// write the html itself
		if err := fmtHTMLTemplate.Execute(cw, ctx); err != nil {
			return err
		}

		// write the html suffix (if any)
		if h.Suffix != nil {
			return h.Suffix(cw)
		}
		return nil
	})
}

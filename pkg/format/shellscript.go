package format

import (
	"net/http"
	"text/template"

	_ "embed"

	"github.com/tkw1536/akhttpd/pkg/count"
	"golang.org/x/crypto/ssh"
)

// spellchecker:words akhttpd

// ShellScript is a zero-size struct that formats ssh keys as a shell script updating an authorized_keys file.
// It implements Formatter.
type ShellScript struct{}

//go:embed shellscript.tpl
var tplShellTemplate string
var fmtShellTemplate = template.Must(template.New("authorized_keys.sh").Parse(tplShellTemplate))

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted as a shell script that updates or creates the file '.ssh/authorized_keys' and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (ShellScript) WriteTo(username, source string, keys []ssh.PublicKey, r *http.Request, w http.ResponseWriter) (int, error) {
	ctx, err := newFmtContext(username, source, keys)
	if err != nil {
		return 0, err
	}

	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys.sh\"")
	headers.Add("Content-Type", "text/x-shellscript")

	return count.Count(w, func(cw *count.Writer) error {
		return fmtShellTemplate.Execute(cw, ctx)
	})
}

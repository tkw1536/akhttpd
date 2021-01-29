package akhttpd

import (
	"bytes"
	"net/http"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh"
)

// Formatter is an object that can write ssh keys to an http.ResponseWriter.
type Formatter interface {
	// WriteTo writes the ssh keys, which are associated with the given user, into w.
	// Returns the number of bytes written and an error.
	WriteTo(username string, keys []ssh.PublicKey, w http.ResponseWriter) (int, error)
}

// fmtContext is an object that is internally used to format values for the templates
type fmtContext struct {
	User string
	Time time.Time

	Keys string
}

// newFmtContext returns a new format context
func newFmtContext(username string, keys []ssh.PublicKey) (ctx fmtContext, err error) {
	ctx.User = username
	ctx.Time = time.Now().UTC()

	var buffer bytes.Buffer
	for _, k := range keys {
		if _, err = buffer.Write(ssh.MarshalAuthorizedKey(k)); err != nil {
			return
		}
	}
	ctx.Keys = buffer.String()
	return
}

// FormatterAuthorizedKeys is a zero-size struct that formats ssh keys as an authorized_keys file.
// It implements Formatter.
type FormatterAuthorizedKeys struct{}

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted in authorized_keys format and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (FormatterAuthorizedKeys) WriteTo(username string, keys []ssh.PublicKey, w http.ResponseWriter) (int, error) {
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

var fmtAuthorizedKeys = template.Must(template.New("authorized_keys").Parse(`
# authorized_keys for {{ .User }}, generated {{ .Time }}
{{.Keys}}
`))

// FormatterShellScript is a zero-size struct that formats ssh keys as a shell script updating an authorized_keys file.
// It implements Formatter.
type FormatterShellScript struct{}

// WriteTo writes the ssh keys, which are associated with the given user, into w.
// They will be formatted as a shell script that updates or creates the file '.ssh/authorized_keys' and include an appropriate Content-Disposition header.
// Returns the number of bytes written in the body of w and an error.
func (FormatterShellScript) WriteTo(username string, keys []ssh.PublicKey, w http.ResponseWriter) (int, error) {
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

var fmtShellTemplate = template.Must(template.New("authorized_keys.sh").Parse(`#!/bin/sh
# authorized_keys.sh for {{ .User }}, generated {{ .Time }}
# This script will setup the authorized_keys file with the right keys
# Warning: This will overwrite any existing keys. 

set -e

SSH_DIR="{{ "$HOME" }}/.ssh"
AK_FILE="{{ "$SSH_DIR" }}/authorized_keys"

echo "Creating and fixing permissions of '{{ "$SSH_DIR" }}' ..."
mkdir -p "{{ "$SSH_DIR" }}"
chmod 700 "{{ "$SSH_DIR" }}"
echo "Writing '{{ "$AK_FILE" }}' ..."
cat > "{{ "$AK_FILE" }}" <<AUTHORIZEDKEYS
{{.Keys}}
AUTHORIZEDKEYS
echo "Fixing permissions of '{{ "$AK_FILE" }}'"
chmod 644 "{{ "$AK_FILE" }}"
`))

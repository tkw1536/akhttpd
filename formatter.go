package akhttpd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Formatter is an object that can Format ssh keys into an http response
type Formatter interface {
	Format(username string, keys []ssh.PublicKey, w http.ResponseWriter)
}

func init() {
	var _ Formatter = (*AuthorizedKeysFormatter)(nil)
	var _ Formatter = (*ShellKeysFormatter)(nil)
}

// AuthorizedKeysFormatter is a Formatter that writes keys in authorized_keys format
type AuthorizedKeysFormatter struct{}

// Format formats the public keys of the provided user in authorized_keys format.
func (AuthorizedKeysFormatter) Format(username string, keys []ssh.PublicKey, w http.ResponseWriter) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys\"")
	headers.Add("Content-Type", "text/plain")

	fmt.Fprintf(w, "# authorized_keys for %s, generated %s\n", username, time.Now().UTC())
	for _, key := range keys {
		w.Write(ssh.MarshalAuthorizedKey(key))
	}
}

// ShellKeysFormatter is a Formatter that is intended to be run by /bin/sh to echo things into .ssh/authorized_keys
type ShellKeysFormatter struct{}

// Format is a Formatter that is intended to be run by /bin/sh
func (ShellKeysFormatter) Format(username string, keys []ssh.PublicKey, w http.ResponseWriter) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"authorized_keys.sh\"")
	headers.Add("Content-Type", "text/x-shellscript")

	fmt.Fprint(w, shellKeysPrefix)
	fmt.Fprintf(w, "# authorized_keys.sh for %s, generated %s", username, time.Now().UTC())
	fmt.Fprintf(w, shellKeysPreKeys)
	for _, key := range keys {
		w.Write(ssh.MarshalAuthorizedKey(key))
	}
	fmt.Fprintf(w, shellKeysPostKeys)

}

const shellKeysScript = `#!/bin/sh
{{head}}
# This script will setup the authorized_keys file with the right keys
# Warning: This will overwrite any existing keys. 

set -e

SSH_DIR="$HOME/.ssh"
AK_FILE="$SSH_DIR/authorized_keys"

echo "Creating and fixing permissions of '$SSH_DIR' ..."
mkdir -p "$SSH_DIR"
chmod 700 "$SSH_DIR"
echo "Writing '$AK_FILE' ..."
cat > "$AK_FILE" <<AUTHORIZEDKEYS
{{content}}
AUTHORIZEDKEYS
echo "Fixing permissions of '$AK_FILE'"
chmod 644 "$AK_FILE"
`

var shellKeysPrefix string
var shellKeysPreKeys string
var shellKeysPostKeys string

func init() {
	splits := []string{"", shellKeysScript}
	splits = strings.SplitN(splits[1], "{{head}}", 2)
	shellKeysPrefix = splits[0]
	splits = strings.SplitN(splits[1], "{{content}}", 2)
	shellKeysPreKeys = splits[0]
	shellKeysPostKeys = splits[1]
}

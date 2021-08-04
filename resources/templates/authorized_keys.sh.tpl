#!/bin/sh
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
{{ range .Keys }}{{.}}{{ end }}
AUTHORIZEDKEYS
echo "Fixing permissions of '{{ "$AK_FILE" }}'"
chmod 644 "{{ "$AK_FILE" }}"
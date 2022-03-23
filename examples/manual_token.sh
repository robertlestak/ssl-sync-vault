#!/bin/bash

# ask the user to provide the vault token, allowing them to log in using any other method
# such as oidc, iam, or browser-based login
echo "Please enter your vault token:"
read -s VAULT_TOKEN
export VAULT_TOKEN

# sync the certificate from vault, and restart nginx when the cert is renewed
sslsync \
	-vault-path kv-name/path/to/value \
	-fullchain /etc/nginx/certs/certchain.pem \
	-key /etc/nginx/certs/certkey.pem \
	-complete-cmd "service nginx reload"
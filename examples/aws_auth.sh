#!/bin/bash

# use the instance's IAM role to authenticate to vault and retrieve the token
export VAULT_TOKEN=`vault login -method=aws -token-only role=vault-role-name`

# sync the certificate from vault, and restart nginx when the cert is renewed
sslsync \
	-vault-path kv-name/path/to/value \
	-fullchain /etc/nginx/certs/certchain.pem \
	-key /etc/nginx/certs/certkey.pem \
	-complete-cmd "service nginx reload"
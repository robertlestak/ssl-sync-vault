#!/bin/bash

# an example of using an AzureAD app to create a JWT token
# which is then passed to vault to authenticate

TENANT_ID="<TENANT_ID>"
CLIENT_ID="<CLIENT_ID>"
CLIENT_SECRET="<CLIENT_SECRET>"
SCOPE=api://$CLIENT_ID/.default
CLIENT_NAME="<CLIENT_NAME>"
ROLE_NAME="<ROLE_NAME>"

# create a JWT token using the AzureAD app
JWT_ACCESS_TOKEN=$(curl -X POST \
	https://login.microsoftonline.com/$TENANT_ID/oauth2/v2.0/token \
	-H "Host: login.microsoftonline.com" \
	-H "Content-Type: application/x-www-form-urlencoded" \
	-d "client_id=$CLIENT_ID&scope=$SCOPE&client_secret=$CLIENT_SECRET&grant_type=client_credentials" | jq -r '.access_token')

# create a vault token using the JWT token
export VAULT_TOKEN=$(curl -X POST -d "{\"jwt\": \"$JWT_ACCESS_TOKEN\", \"role\": \"$ROLE_NAME\"}" \
	$VAULT_ADDR/v1/auth/$CLIENT_NAME/login | jq -r '.auth.client_token')

# sync the certificate from vault, and restart nginx when the cert is renewed
sslsync \
	-vault-path kv-name/path/to/value \
	-fullchain /etc/nginx/certs/certchain.pem \
	-key /etc/nginx/certs/certkey.pem \
	-complete-cmd "service nginx reload"
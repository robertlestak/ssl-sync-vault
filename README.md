# vault-ssl-sync

"Cloud native" systems can utilize systems such as [cert-manager](https://cert-manager.io/) and [cert-manager-sync](https://github.com/robertlestak/cert-manager-sync) to automatically provision and manage TLS certificates for their services.

However more "conventional" systems and VMs still play a vital role in modern infrastructure and should not be forgotten.

This project aims to provide a generic tool to sync certificates between these two different systems.

Certificates can be provisioned and their lifecycle managed through more modern automation tools, and then synchronize their certificates to HashiCorp Vault using [cert-manager-sync](https://github.com/robertlestak/cert-manager-sync).

This script can then be run as a cronjob on the target systems and will connect to vault, retrieve the latest certificate, and update the certificate on the target system on changes. It can then optionally run a command on the target system when the certificate is updated, such as restarting a service.

## Authentication

Currently this script does not implement all of the supported Vault authentication methods. Instead, this relies on the environment to provide a valid `VAULT_TOKEN` with which it can use to authenticate to Vault. This is not ideal, but does mean that it can support all of the auth and integration capabilities out of the box.

In practice this means that you will need to also need to either use the REST API or install the corresponding [Vault binary](https://www.vaultproject.io/downloads) for your platform. See `examples` for more.

## Usage

You can either use environment variables (see `.env-sample`) or command line flags to configure the script.

### Flags

```bash
-complete-cmd string
        command to run when sync is complete
  -fullchain string
        full chain file path
  -key string
        key file path
  -vault-path string
        vault kv path
```

### Scheduling

By design, automated certificate systems can re-issue and push down new certificates on a relatively frequent basis. It is recommended to run this script as a cronjob on the target system to ensure that the certificate is always up to date. 

An example daily crontab entry:

```
    0 0 * * * /path/to/cert-manager-sync.sh
```
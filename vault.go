package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"

	log "github.com/sirupsen/logrus"
)

type VaultSecret struct {
	Addr       string
	Role       string
	AuthMethod string
	Path       string
	KubeToken  string      // auto-filled
	Client     *api.Client // auto-filled
	Token      string      // auto-filled
}

// NewClients creates and returns a new vault client with a valid token or error
func (s *VaultSecret) NewClient() (*api.Client, error) {
	l := log.WithFields(log.Fields{
		"vaultAddr": s.Addr,
		"action":    "vault.NewClient",
	})
	l.Printf("vault.NewClient")
	config := &api.Config{
		Address: s.Addr,
	}
	var err error
	s.Client, err = api.NewClient(config)
	if err != nil {
		l.Printf("vault.NewClient error: %v\n", err)
		return s.Client, err
	}
	if os.Getenv("KUBE_TOKEN") != "" {
		l.Printf("vault.NewClient using KUBE_TOKEN")
		fd, err := ioutil.ReadFile(os.Getenv("KUBE_TOKEN"))
		if err != nil {
			l.Printf("vault.NewClient error: %v\n", err)
			return s.Client, err
		}
		s.KubeToken = string(fd)
	}
	_, terr := s.NewToken()
	if terr != nil {
		l.Printf("vault.NewClient error: %v\n", terr)
		return s.Client, terr
	}
	return s.Client, err
}

// Login creates a vault token with the k8s auth provider
func (s *VaultSecret) Login() (string, error) {
	l := log.WithFields(log.Fields{
		"vaultAddr":  s.Addr,
		"action":     "vault.Login",
		"role":       s.Role,
		"authMethod": s.AuthMethod,
	})
	l.Printf("vault.Login")
	options := map[string]interface{}{
		"role": s.Role,
	}
	if s.KubeToken != "" {
		options["jwt"] = s.KubeToken
	}
	path := fmt.Sprintf("auth/%s/login", s.AuthMethod)
	secret, err := s.Client.Logical().Write(path, options)
	if err != nil {
		l.Printf("vault.Login(%s) error: %v\n", s.AuthMethod, err)
		return "", err
	}
	s.Token = secret.Auth.ClientToken
	l.Printf("vault.Login(%s) success\n", s.AuthMethod)
	s.Client.SetToken(s.Token)
	return s.Token, nil
}

// NewToken generate a new token for session. If LOCAL env var is set and the token is as well, the login is
// skipped and the token is used instead.
func (s *VaultSecret) NewToken() (string, error) {
	l := log.WithFields(log.Fields{
		"vaultAddr": s.Addr,
		"action":    "vault.NewToken",
	})
	l.Printf("vault.NewToken")
	if os.Getenv("VAULT_TOKEN") != "" {
		l.Printf("vault.NewToken using local token")
		s.Token = os.Getenv("VAULT_TOKEN")
		s.Client.SetToken(s.Token)
		return s.Token, nil
	}
	l.Printf("vault.NewToken calling Login")
	return s.Login()
}

// GetKVSecret retrieves a kv secret from vault
func (s *VaultSecret) GetKVSecret() (map[string]interface{}, error) {
	l := log.WithFields(log.Fields{
		"vaultAddr": s.Addr,
		"action":    "vault.GetKVSecret",
		"path":      s.Path,
	})
	l.Printf("vault.GetKVSecret")
	var secrets map[string]interface{}
	if s.Path == "" {
		return secrets, errors.New("secret path required")
	}
	ss := strings.Split(s.Path, "/")
	if len(ss) < 2 {
		l.Errorf("vault.GetKVSecret error: secret path must be in the form of /path/to/secret")
		return secrets, errors.New("Secret path must be in kv/path/to/secret format")
	}
	kv := ss[0]
	kp := strings.Join(ss[1:], "/")
	c := s.Client.Logical()
	secret, err := c.Read(kv + "/data/" + kp)
	if err != nil {
		l.Errorf("vault.GetKVSecret(%s) c.Read error: %v", s, err)
		return secrets, err
	}
	if secret == nil || secret.Data == nil {
		l.Errorf("vault.GetKVSecret(%s) secret is nil", s)
		return nil, errors.New("Secret not found")
	}
	return secret.Data["data"].(map[string]interface{}), nil
}

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SyncConfig struct {
	VaultKVPath         string
	VaultCertField      string
	VaultKeyField       string
	VaultChainField     string
	VaultAddr           string
	FilePathFullChain   string
	FilePathKey         string
	SyncCompleteCommand string
	FullChainContents   string
	KeyContents         string
}

var (
	sc = SyncConfig{}
)

func (s *SyncConfig) writeFiles() error {
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "writeFiles",
	})
	l.Info("starting")
	// check if the parent dir exists
	cpd := filepath.Dir(s.FilePathFullChain)
	kpd := filepath.Dir(s.FilePathKey)
	if _, err := os.Stat(cpd); os.IsNotExist(err) {
		l.Infof("creating dir: %s", cpd)
		err := os.MkdirAll(cpd, 0755)
		if err != nil {
			l.Errorf("error: %v", err)
			return err
		}
	}
	if _, err := os.Stat(kpd); os.IsNotExist(err) {
		l.Infof("creating dir: %s", kpd)
		err := os.MkdirAll(kpd, 0755)
		if err != nil {
			l.Errorf("error: %v", err)
			return err
		}
	}
	// write the files
	l.Infof("writing file: %s", s.FilePathFullChain)
	err := ioutil.WriteFile(s.FilePathFullChain, []byte(s.FullChainContents), 0644)
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	l.Infof("writing file: %s", s.FilePathKey)
	err = ioutil.WriteFile(s.FilePathKey, []byte(s.KeyContents), 0644)
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}

	return nil
}

func (s *SyncConfig) filesChanged() bool {
	var c bool
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "filesChanged",
	})
	l.Info("starting")
	// check if the files exist
	if _, err := os.Stat(s.FilePathFullChain); os.IsNotExist(err) {
		l.Infof("file does not exist: %s", s.FilePathFullChain)
		c = true
		return c
	}
	if _, err := os.Stat(s.FilePathKey); os.IsNotExist(err) {
		l.Infof("file does not exist: %s", s.FilePathKey)
		c = true
		return c
	}
	// read the files
	l.Infof("reading file: %s", s.FilePathFullChain)
	fc, err := ioutil.ReadFile(s.FilePathFullChain)
	if err != nil {
		l.Errorf("error: %v", err)
		return true
	}
	l.Infof("reading file: %s", s.FilePathKey)
	kc, err := ioutil.ReadFile(s.FilePathKey)
	if err != nil {
		l.Errorf("error: %v", err)
		return true
	}
	if s.FullChainContents != string(fc) {
		l.Infof("full chain contents changed")
		return true
	}
	if s.KeyContents != string(kc) {
		l.Infof("key contents changed")
		return true
	}
	l.Infof("no changes")
	return c
}

func (s *SyncConfig) runPostSyncCmd() error {
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "runPostSyncCmd",
	})
	l.Info("starting")
	var cmd string
	var args []string
	if s.SyncCompleteCommand == "" {
		l.Infof("no command to run")
		return nil
	}
	smcd := strings.Split(s.SyncCompleteCommand, " ")
	cmd = smcd[0]
	if len(smcd) > 1 {
		args = smcd[1:]
	}
	l.Infof("running command: %s", cmd)
	rcmd := exec.Command(cmd, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	rcmd.Stdout = &stdout
	rcmd.Stderr = &stderr
	err := rcmd.Run()
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	l.Infof("stdout: %s", stdout.String())
	l.Infof("stderr: %s", stderr.String())
	return nil
}

func (s *SyncConfig) readCertData(cert map[string]interface{}) error {
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "readCertData",
	})
	l.Info("starting")
	var certB64 string
	var keyB64 string
	var chainB64 string
	var ok bool
	if s.VaultCertField == "" {
		l.Error("no cert field")
		return errors.New("no cert field")
	}
	if s.VaultKeyField == "" {
		l.Error("no key field")
		return errors.New("no key field")
	}
	certB64, ok = cert[s.VaultCertField].(string)
	if !ok {
		l.Error("error: certificate not found")
		return errors.New("certificate not found")
	}
	keyB64, ok = cert[s.VaultKeyField].(string)
	if !ok {
		l.Error("error: key not found")
		return errors.New("key not found")
	}
	// base64 decode the cert and key
	certBytes, err := base64.StdEncoding.DecodeString(certB64)
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	s.FullChainContents = string(certBytes)
	if s.VaultChainField != "" {
		chainB64, ok = cert[s.VaultChainField].(string)
		if !ok {
			l.Error("error: chain not found")
			return errors.New("chain not found")
		}
		chainBytes, err := base64.StdEncoding.DecodeString(chainB64)
		if err != nil {
			l.Errorf("error: %v", err)
			return err
		}
		s.FullChainContents = s.FullChainContents + string(chainBytes)
	}
	s.KeyContents = string(keyBytes)
	return nil
}

func (s *SyncConfig) Sync() error {
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "Sync",
	})
	l.Info("starting")
	s.envConfig()
	// authenticate to vault
	vs := VaultSecret{
		Addr: s.VaultAddr,
		Path: s.VaultKVPath,
	}
	_, err := vs.NewClient()
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	// get the cert and key from vault
	cert, err := vs.GetKVSecret()
	if err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	// read the cert data
	if err := s.readCertData(cert); err != nil {
		l.Errorf("error: %v", err)
		return err
	}
	if s.filesChanged() {
		l.Infof("files changed")
		if err := s.writeFiles(); err != nil {
			l.Errorf("error: %v", err)
			return err
		}
		// run the sync complete command
		if err := s.runPostSyncCmd(); err != nil {
			l.Errorf("error: %v", err)
			return err
		}
	}
	return nil
}

func (s *SyncConfig) envConfig() {
	if os.Getenv("VAULT_ADDR") != "" {
		s.VaultAddr = os.Getenv("VAULT_ADDR")
	}
	if os.Getenv("VAULT_KV_PATH") != "" {
		s.VaultKVPath = os.Getenv("VAULT_KV_PATH")
	}
	if os.Getenv("FULL_CHAIN_FILE") != "" {
		s.FilePathFullChain = os.Getenv("FULL_CHAIN_FILE")
	}
	if os.Getenv("KEY_FILE") != "" {
		s.FilePathKey = os.Getenv("KEY_FILE")
	}
	if os.Getenv("SYNC_COMPLETE_COMMAND") != "" {
		s.SyncCompleteCommand = os.Getenv("SYNC_COMPLETE_COMMAND")
	}
	if os.Getenv("VAULT_CERT_FIELD") != "" {
		s.VaultCertField = os.Getenv("VAULT_CERT_FIELD")
	}
	if os.Getenv("VAULT_CHAIN_FIELD") != "" {
		s.VaultChainField = os.Getenv("VAULT_CHAIN_FIELD")
	}
	if os.Getenv("VAULT_KEY_FIELD") != "" {
		s.VaultKeyField = os.Getenv("VAULT_KEY_FIELD")
	}
}

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func main() {
	l := log.WithFields(log.Fields{
		"pkg": "main",
		"fn":  "main",
	})
	l.Info("starting")
	flag.StringVar(&sc.VaultKVPath, "vault-path", "", "vault kv path")
	flag.StringVar(&sc.FilePathFullChain, "fullchain", "", "full chain file path")
	flag.StringVar(&sc.FilePathKey, "key", "", "key file path")
	flag.StringVar(&sc.SyncCompleteCommand, "complete-cmd", "", "command to run when sync is complete")
	flag.StringVar(&sc.VaultCertField, "vault-cert-field", "", "vault cert field")
	flag.StringVar(&sc.VaultKeyField, "vault-key-field", "", "vault key field")
	flag.StringVar(&sc.VaultChainField, "vault-chain-field", "", "vault chain field")
	flag.Parse()

	if sc.VaultKVPath == "" {
		// print usage
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := sc.Sync()
	if err != nil {
		l.Errorf("error: %v", err)
		os.Exit(1)
	}
}

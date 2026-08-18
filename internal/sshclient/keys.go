package sshclient

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the path to the SSH keys directory (~/.config/baredeploy/keys),
// creating it if necessary. Returns an error if the home directory cannot be
// resolved or the directory cannot be created.
func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	dir := filepath.Join(configDir, "baredeploy", "keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating keys directory: %w", err)
	}
	return dir, nil
}

// PathForHost returns the path to the provisioned key for the given host.
// Returns an empty string if no key exists for that host.
//
// The key path is ~/.config/baredeploy/keys/
func PathForHost(user string, host string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	keyPath := filepath.Join(configDir, "baredeploy", "keys", filename(user, host))
	if _, err := os.Stat(keyPath); err != nil {
		return ""
	}
	return keyPath
}

// SaveKey stores PEM-encoded key data to the keys directory under the given host name.
// Returns the path where the key was saved.
func SaveKey(user string, host string, pemData []byte) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(dir, filename(user, host))
	if err := os.WriteFile(keyPath, pemData, 0600); err != nil {
		return "", fmt.Errorf("saving private key: %w", err)
	}

	return keyPath, nil
}

// Addrs formats the SSH address for the given user, host, and port.
func Addr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func filename(user string, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

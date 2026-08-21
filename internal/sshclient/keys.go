package sshclient

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GenerateKeyPair creates a new ed25519 keypair and returns the public
// key formatted as an authorized_keys line and the private key encoded
// as PEM.
func GenerateKeyPair() (pubKeyLine string, privPEM []byte, err error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", nil, fmt.Errorf("generating keypair: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return "", nil, fmt.Errorf("marshaling private key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", nil, fmt.Errorf("creating public key: %w", err)
	}

	pubKeyLine = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
	return pubKeyLine, pem.EncodeToMemory(block), nil
}

// AuthorizeKey ensures ~/.ssh exists on the remote host with correct
// permissions and appends the public key to authorized_keys idempotently.
func AuthorizeKey(ctx context.Context, conn *Connection, pubKeyLine string) error {
	if _, err := conn.Run(ctx, "mkdir", "-p", "~/.ssh"); err != nil {
		return fmt.Errorf("creating .ssh directory: %w", err)
	}
	if _, err := conn.Run(ctx, "chmod", "700", "~/.ssh"); err != nil {
		return fmt.Errorf("setting .ssh permissions: %w", err)
	}

	escaped := strings.ReplaceAll(pubKeyLine, "'", `'"'"'`)
	injectCmd := fmt.Sprintf("grep -qF '%s' ~/.ssh/authorized_keys || echo '%s' >> ~/.ssh/authorized_keys", escaped, escaped)
	if _, err := conn.Run(ctx, "sh", "-c", injectCmd); err != nil {
		return fmt.Errorf("injecting public key: %w", err)
	}

	if _, err := conn.Run(ctx, "chmod", "600", "~/.ssh/authorized_keys"); err != nil {
		return fmt.Errorf("setting authorized_keys permissions: %w", err)
	}

	return nil
}

// Dir returns the path to the SSH keys directory (~/.config/baredeploy/keys),
// creating it if necessary. Returns an error if the home directory cannot be
// resolved or the directory cannot be created.
func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config directory: %w", err)
	}

	dir := filepath.Join(configDir, "baredeploy", "keys")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating keys directory: %w", err)
	}
	return dir, nil
}

// ResolveKey returns the SSH private key path to use for user@host.
// An explicitly provided path takes precedence; otherwise the key
// previously provisioned by `baredeploy authorize` is used.
func ResolveKey(explicitKey, user, host string) (string, error) {
	if explicitKey != "" {
		return explicitKey, nil
	}

	key := PathForHost(user, host)
	if key == "" {
		return "", fmt.Errorf("no SSH key found for %s@%s (run `baredeploy authorize` first or pass --key)", user, host)
	}

	return key, nil
}

// PathForHost returns the path to the provisioned key for the given host.
// Returns an empty string if no key exists for that host.
func PathForHost(user, host string) string {
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

// SaveKey stores PEM-encoded key data to the keys directory for the given
// user and host. Returns the path where the key was saved.
func SaveKey(user, host string, pemData []byte) (string, error) {
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

// Addr formats the SSH address for the given host and port.
func Addr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func filename(user, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

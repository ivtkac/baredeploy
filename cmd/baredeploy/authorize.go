package main

import (
	"crypto/ed25519"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/ivtkac/baredeploy/internal/sshclient"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func authorizeCmd() *cobra.Command {
	var user string
	var port int
	var password string

	cmd := &cobra.Command{
		Use:   "authorize <host>",
		Short: "Generate a temporary SSH key and inject it into the rescue system",
		Long: `Generates a new ed25519 keypair, connects to the target host via
password authentication, and adds the public key to ~/.ssh/authorized_keys.
The private key is saved locally for subsequent baredeploy commands.

The user is prompted for the rescue system password if --password is not provided.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			addr := sshclient.Addr(host, port)

			if password == "" {
				fmt.Printf("Rescue system password for %s@%s: ", user, host)
				pwd, err := readPassword()
				if err != nil {
					return fmt.Errorf("reading password: %w", err)
				}
				password = pwd
			}

			pub, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				return fmt.Errorf("generating keypair: %w", err)
			}

			privPEM, err := encodePrivateKey(priv)
			if err != nil {
				return err
			}
			pubKeyLine := encodePublicKey(pub)

			conn, err := sshclient.ConnectByPassword(addr, user, password)
			if err != nil {
				return err
			}
			defer conn.Cleanup()

			if err := injectSSHKey(conn, cmd, pubKeyLine); err != nil {
				return err
			}

			keyPath, err := sshclient.SaveKey(user, host, privPEM)
			if err != nil {
				return err
			}

			fmt.Println("Key provisioned successfully.")
			fmt.Printf("Private key saved to: %s\n", keyPath)
			fmt.Println("Subsequent commands will use it automatically.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH username")
	cmd.Flags().IntVarP(&port, "port", "p", 22, "SSH port")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Rescue system password (promt if empty)")

	return cmd
}

// encodePrivateKey marshals an Ed25519 private key to PEM bytes.
func encodePrivateKey(priv ed25519.PrivateKey) ([]byte, error) {
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, fmt.Errorf("marshaling private key: %w", err)
	}
	return pem.EncodeToMemory(block), nil
}

// encodePublicKey formats an Ed25519 public key for authorized_keys.
func encodePublicKey(pub ed25519.PublicKey) string {
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		panic(fmt.Sprintf("creating public key: %v", err)) // should never fail
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
}

// injectSSHKey ensures ~/.ssh exists with correct permissions and appends
// the public key to authorized_keys idempotently.
func injectSSHKey(connection *sshclient.Connection, cmd *cobra.Command, pubKeyLine string) error {
	ctx := cmd.Context()

	if _, err := connection.Run(ctx, "mkdir", "-p", "~/.ssh"); err != nil {
		return fmt.Errorf("creating .ssh directory: %w", err)
	}
	if _, err := connection.Run(ctx, "chmod", "700", "~/.ssh"); err != nil {
		return fmt.Errorf("setting .ssh permissions: %w", err)
	}

	escaped := strings.ReplaceAll(pubKeyLine, "'", "'\"'\"'")
	injectCmd := fmt.Sprintf("grep -qF '%s' ~/.ssh/authorized_keys || echo '%s' >> ~/.ssh/authorized_keys", escaped, escaped)
	if _, err := connection.Run(ctx, "sh", "-c", injectCmd); err != nil {
		return fmt.Errorf("injecting public key: %w", err)
	}

	if _, err := connection.Run(ctx, "chmod", "600", "~/.ssh/authorized_keys"); err != nil {
		return fmt.Errorf("setting authorized_keys permissions: %w", err)
	}

	return nil
}

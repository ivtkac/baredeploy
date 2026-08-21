package main

import (
	"fmt"

	"github.com/ivtkac/baredeploy/internal/sshclient"
	"github.com/spf13/cobra"
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

			if password == "" {
				fmt.Printf("Rescue system password for %s@%s: ", user, host)
				pwd, err := readPassword()
				if err != nil {
					return fmt.Errorf("reading password: %w", err)
				}
				password = pwd
			}

			pubKeyLine, privPEM, err := sshclient.GenerateKeyPair()
			if err != nil {
				return err
			}

			conn, err := sshclient.ConnectByPassword(sshclient.Addr(host, port), user, password)
			if err != nil {
				return err
			}
			defer conn.Cleanup()

			if err := sshclient.AuthorizeKey(cmd.Context(), conn, pubKeyLine); err != nil {
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
	cmd.Flags().StringVarP(&password, "password", "P", "", "Rescue system password (prompt if empty)")

	return cmd
}

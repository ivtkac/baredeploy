// Package ssh provides a remote command executor over SSH.
//
// It implements [runner.Executor]
package sshclient

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ivtkac/baredeploy/internal/runner"
	"golang.org/x/crypto/ssh"
)

type Connection struct {
	client *ssh.Client
}

// ConnectByPassword connects to the remote host using password authentication.
//
//	addr  — "host:port" or just "host" (port 22 assumed)
//	user  — SSH username (e.g. "root")
//	pass  — SSH password
func ConnectByPassword(addr, user, pass string) (*Connection, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return dial(addr, config)
}

// ConnectByKey connects to the remote host using an SSH private key.
//
//	addr  — "host:port" or just "host" (port 22 assumed)
//	user  — SSH username (e.g. "root")
//	key   — path to the private key file
func ConnectByKey(addr, user, keyPath string) (*Connection, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key file %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key file %s: %w", keyPath, err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		// Key verification is deferred to the user's explicit trust
		// when running `baredeploy authorize` for the first time.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return dial(addr, config)
}

func dial(addr string, config *ssh.ClientConfig) (*Connection, error) {
	if !strings.Contains(addr, ":") {
		addr = addr + ":22"
	}

	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s@%s: %w", config.User, addr, err)
	}

	return &Connection{client: conn}, nil
}

// Run executes `name args...` on the remote host via SSH.
//
// The remote shell is invoked as `name arg1 arg2...` (no shell escaping).
// For complex commands, use `sh -c "..."` as the name with args.
func (c *Connection) Run(ctx context.Context, name string, args ...string) (runner.Result, error) {
	cmd := buildCommand(name, args...)

	session, err := c.client.NewSession()
	if err != nil {
		return runner.Result{}, fmt.Errorf("creating SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr strings.Builder
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		<-done
		return runner.Result{ExitCode: 124}, fmt.Errorf("command timed out: %s", cmd)
	case err := <-done:
		res := runner.Result{
			Stdout: []byte(stdout.String()),
			Stderr: []byte(stderr.String()),
		}
		if err != nil {
			res.ExitCode = 1
			return res, fmt.Errorf("command %s failed on remote host: %w: %s", cmd, err, runner.Trimmed(stderr.String()))
		}
		return res, nil
	}
}

// Cleanup closes the SSH connection.
// It is safe to call multiple times.
func (c *Connection) Cleanup() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}

func buildCommand(name string, args ...string) string {
	if len(args) == 0 {
		return name
	}

	parts := make([]string, 0, len(args)+1)
	parts = append(parts, name)
	parts = append(parts, args...)

	return strings.Join(parts, " ")
}

// ResolveKey returns the SSH key for the given path.
func ResolveKey(explicitKey, user string, host string) (string, error) {
	if explicitKey != "" {
		return explicitKey, nil
	}

	key := PathForHost(user, host)
	if key == "" {
		return "", fmt.Errorf("no SSH key found for %s%s", user, host)
	}

	return key, nil
}

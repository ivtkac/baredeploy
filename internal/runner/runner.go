// Package runner defines the command-execution abstraction used across
// baredeploy and provides a local implementation of it.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Executor runs a command and returns its result. Implementations exist
// for the local machine ([Local]) and for remote hosts over SSH
// (sshclient.Connection).
type Executor interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

// DefaultTimeout bounds how long any single command may run unless the
// caller's context already carries a deadline.
const DefaultTimeout = 10 * time.Second

// Result holds the output of a command execution.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Local is an [Executor] that runs commands on the local machine.
type Local struct{}

// Run executes `name args...` directly (i.e. without a shell) and
// returns the captured output.
func (Local) Run(ctx context.Context, name string, args ...string) (Result, error) {
	ctx, cancel := WithDefaultTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	res := Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return res, fmt.Errorf("%s %v: timed out", name, args)
	}
	if err != nil {
		return res, fmt.Errorf("%s %v: %w: %s", name, args, err, Trimmed(stderr.String()))
	}
	return res, nil
}

// WithDefaultTimeout applies [DefaultTimeout] to the context unless it
// already carries a deadline.
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, DefaultTimeout)
}

// Trimmed truncates a string to a safe length for display in errors.
func Trimmed(s string) string {
	const max = 2000
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

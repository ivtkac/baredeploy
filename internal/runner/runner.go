package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Executor interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

// DefaultTimeout bounds how long any single discovery command may run.
// Discovery commands are expected to complete within this time.
const DefaultTimeout = 10 * time.Second

// Result holds the output of a command execution.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Run executes `name args...` directly (i.e. without shell) and
// returns the result to output. If the command fails, a error is outputed to stderr.
func Run(ctx context.Context, name string, args ...string) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
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
		return res, fmt.Errorf("%s %v: timed out after %s", name, args, DefaultTimeout)
	}
	if err != nil {
		return res, fmt.Errorf("%s %v: %w: %s", name, args, err, Trimmed(stderr.String()))
	}
	return res, nil
}

// LookPath reports whether a required tool is available on the system.
func LookPath(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("required tool %s not found in PATH: %w", name, err)
	}
	return nil
}

// Trimmed truncates a string to MAX characters for safe display.
func Trimmed(s string) string {
	const MAX = 2000
	if len(s) > MAX {
		return s[:MAX] + "..."
	}
	return s
}

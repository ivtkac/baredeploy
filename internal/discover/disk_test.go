package discover

import (
	"context"
	"testing"

	"github.com/ivtkac/baredeploy/internal/runner"
)

// fakeExecutor returns canned stdout for any command.
type fakeExecutor struct {
	stdout string
}

func (f fakeExecutor) Run(ctx context.Context, name string, args ...string) (runner.Result, error) {
	return runner.Result{Stdout: []byte(f.stdout)}, nil
}

const lsblkJSON = `{
	"blockdevices": [
		{"name": "sda", "type": "disk", "size": 512110190592, "mountpoint": null},
		{"name": "sdb", "type": "disk", "size": 512110190592, "mountpoint": "/mnt"},
		{"name": "sdc", "type": "disk", "size": 0, "mountpoint": null},
		{"name": "sr0", "type": "rom", "size": 1073741312, "mountpoint": null}
	]
}`

func TestTargetDisks(t *testing.T) {
	targets, err := TargetDisks(context.Background(), fakeExecutor{stdout: lsblkJSON})
	if err != nil {
		t.Fatalf("TargetDisks() error: %v", err)
	}

	if len(targets) != 1 {
		t.Fatalf("got %d targets, want 1: %+v", len(targets), targets)
	}
	if targets[0].Name != "sda" {
		t.Errorf("got target %q, want %q", targets[0].Name, "sda")
	}
}

func TestDevices(t *testing.T) {
	out, err := Devices(context.Background(), fakeExecutor{stdout: lsblkJSON})
	if err != nil {
		t.Fatalf("Devices() error: %v", err)
	}
	if len(out.BlockDevices) != 4 {
		t.Errorf("got %d devices, want 4", len(out.BlockDevices))
	}
}

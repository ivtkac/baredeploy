package discover

import (
	"context"

	lsblk "github.com/ivtkac/baredeploy/internal/exec"
	"github.com/ivtkac/baredeploy/internal/runner"
)

func DiscoverDevices(ex runner.Executor) (*lsblk.Output, error) {
	res, err := ex.Run(context.Background(), "lsblk",
		"-J", "-b", "-o", lsblk.Columns())
	if err != nil {
		return nil, err
	}
	return lsblk.Parse(res.Stdout)
}

func FindTargetDisks(ex runner.Executor) ([]lsblk.BlockDevice, error) {
	out, err := DiscoverDevices(ex)
	if err != nil {
		return nil, err
	}
	var targets []lsblk.BlockDevice
	for _, dev := range out.BlockDevices {
		if dev.Type == "disk" && dev.MountPoint == nil && dev.Size != nil {
			targets = append(targets, dev)
		}
	}
	return targets, nil
}

// Package discover inspects the target system's hardware.
package discover

import (
	"context"

	"github.com/ivtkac/baredeploy/internal/lsblk"
	"github.com/ivtkac/baredeploy/internal/runner"
)

// Devices lists all block devices on the target system.
func Devices(ctx context.Context, ex runner.Executor) (*lsblk.Output, error) {
	res, err := ex.Run(ctx, "lsblk", "-J", "-b", "-o", lsblk.Columns())
	if err != nil {
		return nil, err
	}
	return lsblk.Parse(res.Stdout)
}

// TargetDisks lists disks eligible as OS installation targets:
// unmounted disks with a non-zero size.
func TargetDisks(ctx context.Context, ex runner.Executor) ([]lsblk.BlockDevice, error) {
	out, err := Devices(ctx, ex)
	if err != nil {
		return nil, err
	}

	var targets []lsblk.BlockDevice
	for _, dev := range out.BlockDevices {
		if isTargetDisk(dev) {
			targets = append(targets, dev)
		}
	}
	return targets, nil
}

func isTargetDisk(dev lsblk.BlockDevice) bool {
	return dev.Type == "disk" && dev.MountPoint == nil && dev.Size != nil && *dev.Size > 0
}

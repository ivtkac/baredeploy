package main

import (
	"fmt"
	"os"

	"github.com/ivtkac/baredeploy/internal/discover"
	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	var rf remoteFlags

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover hardware",
	}

	rf.addFlags(cmd)
	cmd.AddCommand(discoverDevicesCmd(&rf))
	cmd.AddCommand(discoverDisksCmd(&rf))

	return cmd
}

func discoverDevicesCmd(rf *remoteFlags) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List all block devices on the target system",
		RunE: func(cmd *cobra.Command, args []string) error {
			ex, cleanup, err := rf.executor()
			if err != nil {
				return err
			}
			defer cleanup()

			out, err := discover.Devices(cmd.Context(), ex)
			if err != nil {
				return err
			}

			if jsonOutput {
				return writeJSON(os.Stdout, out)
			}

			printDeviceTree(os.Stdout, out.BlockDevices)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func discoverDisksCmd(rf *remoteFlags) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "disks",
		Short: "List disks eligible for OS installation",
		Long: `Shows unmounted disks with non-zero size that can be used as
installation targets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ex, cleanup, err := rf.executor()
			if err != nil {
				return err
			}
			defer cleanup()

			targets, err := discover.TargetDisks(cmd.Context(), ex)
			if err != nil {
				return err
			}

			if jsonOutput {
				return writeJSON(os.Stdout, targets)
			}
			if len(targets) == 0 {
				fmt.Fprintln(os.Stderr, "No eligible install targets found.")
				return nil
			}

			printDiskTargets(os.Stdout, targets)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/ivtkac/baredeploy/internal/discover"
	lsblk "github.com/ivtkac/baredeploy/internal/exec"
	"github.com/ivtkac/baredeploy/internal/runner"
	"github.com/ivtkac/baredeploy/internal/sshclient"
	"github.com/spf13/cobra"
)

type remoteFlags struct {
	host string
	user string
	port int
	key  string
}

func (rf *remoteFlags) addFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&rf.host, "host", "H", "", "Remote host (IP or hostname)")
	cmd.PersistentFlags().StringVarP(&rf.user, "user", "u", "root", "SSH username")
	cmd.PersistentFlags().IntVarP(&rf.port, "port", "p", 22, "SSH port")
	cmd.PersistentFlags().StringVar(&rf.key, "key", "", "Path to SSH private key")
}

func (rf *remoteFlags) executor() (runner.Executor, func(), error) {
	if rf.host == "" {
		return &localExecutor{}, func() {}, nil
	}

	addr := sshclient.Addr(rf.host, rf.port)
	keyPath, err := sshclient.ResolveKey(rf.key, rf.user, rf.host)
	if err != nil {
		return nil, nil, err
	}

	client, err := sshclient.ConnectByKey(addr, rf.user, keyPath)
	if err != nil {
		return nil, nil, err
	}
	return client, client.Cleanup, nil
}

type localExecutor struct{}

func (e *localExecutor) Run(ctx context.Context, name string, args ...string) (runner.Result, error) {
	return runner.Run(ctx, name, args...)
}

func discoverCmd() *cobra.Command {
	var rf remoteFlags

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover hardware",
	}

	rf.addFlags(cmd)
	cmd.AddCommand(discoverDevicesCmd(&rf))
	cmd.AddCommand(findTargetDisksCmd(&rf))

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

			out, err := discover.DiscoverDevices(ex)
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

func findTargetDisksCmd(rf *remoteFlags) *cobra.Command {
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

			targets, err := discover.FindTargetDisks(ex)
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

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newTable(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func printDeviceTree(w io.Writer, devices []lsblk.BlockDevice) {
	tw := newTable(w)
	fmt.Fprintln(tw, "NAME\tSIZE\tMOUNTPOINT")
	for _, dev := range devices {
		writeDeviceRows(tw, dev, "", true)
	}
	tw.Flush()
}

func writeDeviceRows(tw *tabwriter.Writer, dev lsblk.BlockDevice, prefix string, isLast bool) {
	branch := prefix

	switch {
	case prefix == "":
	case isLast:
		branch += "└─"
	default:
		branch += "├─"
	}

	fmt.Fprintf(tw, "%s\t%s\t%s\n", branch+dev.Name, formatSize(dev.Size), orDash(dev.MountPoint))

	childPrefix := prefix
	switch {
	case prefix == "":
		childPrefix = "  "
	case isLast:
		childPrefix += "  "
	default:
		childPrefix += "│ "
	}

	for i, child := range dev.Children {
		writeDeviceRows(tw, child, childPrefix, i == len(dev.Children)-1)
	}
}

func printDiskTargets(w io.Writer, targets []lsblk.BlockDevice) {
	tw := newTable(w)
	fmt.Fprintln(tw, "NAME\tSIZE\tMODEL\tVENDOR")
	for _, dev := range targets {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			dev.Name, formatSize(dev.Size), orDash(dev.Model), orDash(dev.Vendor))
	}
	tw.Flush()
}

func formatSize(size *uint64) string {
	if size == nil || *size == 0 {
		return "-"
	}
	return humanSize(*size)
}

func humanSize(bytes uint64) string {
	const unit = 1000
	units := []string{"KB", "MB", "GB", "TB"}

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/div, units[exp])
}

func orDash(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}

package main

import (
	"github.com/ivtkac/baredeploy/internal/runner"
	"github.com/ivtkac/baredeploy/internal/sshclient"
	"github.com/spf13/cobra"
)

// remoteFlags holds the flags shared by every command that can target
// either the local machine or a remote host over SSH.
type remoteFlags struct {
	host string
	user string
	port int
	key  string
}

func (rf *remoteFlags) addFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&rf.host, "host", "H", "", "Remote host (IP or hostname); empty runs locally")
	cmd.PersistentFlags().StringVarP(&rf.user, "user", "u", "root", "SSH username")
	cmd.PersistentFlags().IntVarP(&rf.port, "port", "p", 22, "SSH port")
	cmd.PersistentFlags().StringVar(&rf.key, "key", "", "Path to SSH private key")
}

// executor returns an Executor for the selected target: the local
// machine when --host is empty, otherwise an SSH connection.
// The returned cleanup func must always be called.
func (rf *remoteFlags) executor() (runner.Executor, func(), error) {
	if rf.host == "" {
		return runner.Local{}, func() {}, nil
	}

	keyPath, err := sshclient.ResolveKey(rf.key, rf.user, rf.host)
	if err != nil {
		return nil, nil, err
	}

	conn, err := sshclient.ConnectByKey(sshclient.Addr(rf.host, rf.port), rf.user, keyPath)
	if err != nil {
		return nil, nil, err
	}
	return conn, conn.Cleanup, nil
}

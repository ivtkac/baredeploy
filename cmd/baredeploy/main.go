package main

import (
	"os"

	"github.com/ivtkac/baredeploy/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:     "baredeploy",
		Short:   "baredeploy is a tool for deploying bare metal servers",
		Version: version.String(),
	}

	root.AddCommand(discoverCmd())
	root.AddCommand(authorizeCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

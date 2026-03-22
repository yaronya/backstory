package main

import (
	"fmt"
	"os"

	"github.com/yaronya/backstory/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "backstory",
		Short: "Shared team memory for AI coding agents",
	}
	root.AddCommand(
		cli.NewInitCmd(),
		cli.NewSyncCmd(),
		cli.NewIndexCmd(),
		cli.NewSearchCmd(),
		cli.NewInjectCmd(),
		cli.NewCaptureCmd(),
		cli.NewStatusCmd(),
		cli.NewEditCmd(),
		cli.NewAddCmd(),
		cli.NewStaleCmd(),
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

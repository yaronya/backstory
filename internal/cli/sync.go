package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync decisions repo with remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory sync: not yet implemented")
			return nil
		},
	}
}

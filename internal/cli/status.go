package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show decisions repo status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory status: not yet implemented")
			return nil
		},
	}
}

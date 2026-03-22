package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory edit: not yet implemented")
			return nil
		},
	}
}

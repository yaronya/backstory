package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a decisions repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory init: not yet implemented")
			return nil
		},
	}
}

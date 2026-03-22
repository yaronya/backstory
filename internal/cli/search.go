package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search decisions by query",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory search: not yet implemented")
			return nil
		},
	}
}

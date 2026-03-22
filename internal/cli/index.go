package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Index decisions for search",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory index: not yet implemented")
			return nil
		},
	}
}

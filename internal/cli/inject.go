package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inject",
		Short: "Inject relevant decisions into agent context",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory inject: not yet implemented")
			return nil
		},
	}
}

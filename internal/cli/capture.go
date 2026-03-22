package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCaptureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "capture",
		Short: "Capture a new decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory capture: not yet implemented")
			return nil
		},
	}
}

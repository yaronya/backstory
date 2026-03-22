package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Index decisions for search",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			return rebuildIndex(repoPath)
		},
	}
}

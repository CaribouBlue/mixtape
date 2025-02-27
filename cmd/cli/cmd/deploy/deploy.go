package deploy

import (
	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the app in full or parts",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	DeployCmd.AddCommand(secretCmd)
}

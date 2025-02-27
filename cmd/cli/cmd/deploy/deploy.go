package deploy

import (
	"github.com/CaribouBlue/mixtape/cmd/cli/config"
	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the app in full or parts",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	DeployCmd.PersistentFlags().StringVarP(&flagDockerContext, "docker-context", "c", config.GetConfigValue(config.ConfDockerContext), "The docker context to use")

	DeployCmd.AddCommand(secretCmd)
	DeployCmd.AddCommand(appCmd)
}

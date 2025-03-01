package db

import (
	"github.com/CaribouBlue/mixtape/cmd/cli/config"
	"github.com/spf13/cobra"
)

var DbCmd = &cobra.Command{
	Use:   "db",
	Short: "Work with an app database",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var (
	flagDbPath string
)

func init() {
	DbCmd.PersistentFlags().StringVarP(&flagDbPath, "db-path", "p", config.GetConfigValue(config.ConfDbPath), "The path to the database")

	DbCmd.AddCommand(setupCmd)
	DbCmd.AddCommand(loadTestDataCmd)
}

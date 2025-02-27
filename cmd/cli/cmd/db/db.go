package db

import (
	"github.com/spf13/cobra"
)

var DbCmd = &cobra.Command{
	Use:   "db",
	Short: "Work with an app database",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	DbCmd.AddCommand(setupCmd)
	DbCmd.AddCommand(loadTestDataCmd)
}

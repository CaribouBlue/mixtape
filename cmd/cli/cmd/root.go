package cmd

import (
	"fmt"
	"os"

	"github.com/CaribouBlue/mixtape/cmd/cli/cmd/db"
	"github.com/CaribouBlue/mixtape/cmd/cli/cmd/deploy"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mixtape-cli",
	Short: "A CLI tool for managing the Mixtape application",
}

func init() {
	rootCmd.AddCommand(deploy.DeployCmd)
	rootCmd.AddCommand(db.DbCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

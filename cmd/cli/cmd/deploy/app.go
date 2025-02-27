package deploy

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app [NAME]",
	Short: "Deploy the app",
	Long:  "Deploy the app to a docker stack, defaults to mixtape if no name is provided.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stackName := "mixtape"
		if len(args) > 0 {
			stackName = args[0]
		}

		initialDockerContext, err := GetCurrentDockerContext()
		if err != nil {
			log.Fatalln("Failed to get current docker context:", err)
		}
		defer func() {
			err := SetDockerContext(initialDockerContext)
			if err != nil {
				log.Fatalln("Failed to set docker context to", initialDockerContext, ":", err)
			}
		}()

		err = SetDockerContext(flagDockerContext)
		if err != nil {
			log.Fatalln("Failed to set docker context to", flagDockerContext, ":", err)
		}
		log.Println("Using Docker context", flagDockerContext)

		withRegistryAuth := ""
		if flagWithRegistryAuth {
			withRegistryAuth = "--with-registry-auth"
		}
		osCmd := exec.Command("docker", "stack", "deploy", "-c", flagComposeFile, stackName, withRegistryAuth)
		osCmd.Stdout = os.Stdout
		osCmd.Stderr = os.Stderr
		err = osCmd.Run()
		if err != nil {
			log.Fatalln("Failed to run", "`"+osCmd.String()+"`", "with", err)
		}
		log.Println("Deployed stack", stackName)
	},
}

var (
	flagComposeFile      string
	flagWithRegistryAuth bool
)

func init() {
	appCmd.Flags().StringVarP(&flagComposeFile, "compose-file", "f", "./compose.yaml", "The docker compose file to use")
	appCmd.Flags().BoolVarP(&flagWithRegistryAuth, "with-registry-auth", "w", true, "Pass registry auth to the stack")
}

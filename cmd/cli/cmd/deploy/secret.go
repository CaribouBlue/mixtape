package deploy

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret NAME FILE",
	Short: "Deploy an app secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		secretTarget := args[0]
		secretFile := args[1]

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

		osCmd := exec.Command("docker", "secret", "ls", "--format", "{{.Name}}", "--filter", fmt.Sprintf("Name=%s", secretTarget))
		output, err := osCmd.CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		staleSecretName := strings.Trim(string(output), "\n")

		useTempSecret := false
		if staleSecretName == "" {
			log.Println("No secret found with name", secretTarget, "creating new secret")
		} else {
			log.Println("Found secret", staleSecretName)
			useTempSecret = true
		}

		newSecretName := secretTarget
		if useTempSecret {
			newSecretName = secretTarget + "-temp"
		}

		err = CreateDockerSecret(newSecretName, secretFile)
		if err != nil {
			log.Fatalln("Failed to create secret", newSecretName, "with", err)
		}
		log.Println("Created new secret", newSecretName)

		err = UpdateDockerServiceSecret(flagServiceName, staleSecretName, newSecretName, secretTarget)
		if err != nil {
			log.Fatalln("Failed to update service with", err)
		}
		log.Println("Updated service with new secret", newSecretName)

		err = RemoveDockerSecret(staleSecretName)
		if err != nil {
			log.Fatalln("Failed to remove stale secret", staleSecretName, "with", err)
		}
		log.Println("Removed stale secret", staleSecretName)

		if useTempSecret {
			tempSecretName := newSecretName
			secretName := secretTarget

			err = CreateDockerSecret(secretName, secretFile)
			if err != nil {
				log.Fatalln("Failed to create secret", secretName, "with", err)
			}
			log.Println("Created new secret", secretName)

			err = UpdateDockerServiceSecret(flagServiceName, tempSecretName, secretName, secretTarget)
			if err != nil {
				log.Fatalln("Failed to update service with", err)
			}
			log.Println("Updated service with new secret", secretName)

			err = RemoveDockerSecret(tempSecretName)
			if err != nil {
				log.Fatalln("Failed to remove stale secret", tempSecretName, "with", err)
			}
			log.Println("Removed temp secret", tempSecretName)
		}
	},
}

var (
	flagDockerContext string
	flagServiceName   string
)

func init() {
	secretCmd.Flags().StringVarP(&flagServiceName, "service-name", "n", "mixtape_app", "The name of the service to update")
}

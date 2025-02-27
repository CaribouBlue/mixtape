package deploy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/CaribouBlue/mixtape/cmd/cli/config"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret NAME FILE",
	Short: "Deploy an app secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		secretName := args[0]
		secretFile := args[1]

		osCmd := exec.Command("docker", "context", "show")
		output, err := osCmd.CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		initialDockerContext := strings.Trim(string(output), "\n")

		osCmd = exec.Command("docker", "context", "use", flagDockerContext)
		err = osCmd.Run()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		log.Println("Using Docker context", flagDockerContext)

		osCmd = exec.Command("docker", "secret", "ls", "--format", "{{.Name}}", "--filter", fmt.Sprintf("Name=%s", secretName))
		output, err = osCmd.CombinedOutput()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		currentSecretName := strings.Trim(string(output), "\n")

		if currentSecretName == "" {
			log.Fatalln("No secret found with name", secretName)
		} else {
			log.Println("Found secret", currentSecretName)
		}

		secretNameEnding := currentSecretName[len(currentSecretName)-2:]
		if secretNameEnding == "-x" {
			secretNameEnding = "y"
		} else {
			secretNameEnding = "x"
		}
		newSecretName := fmt.Sprintf("%s-%s", secretName, secretNameEnding)

		osCmd = exec.Command("docker", "secret", "create", newSecretName, secretFile)
		err = osCmd.Run()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		log.Println("Created new secret", newSecretName)

		osCmd = exec.Command("docker", "service", "update", flagServiceName, "--secret-rm", currentSecretName, "--secret-add", "source="+newSecretName+",target="+secretName)
		osCmd.Stdout = os.Stdout
		err = osCmd.Run()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		log.Printf("Updated service %s to use new secret %s as %s", flagServiceName, newSecretName, secretName)

		osCmd = exec.Command("docker", "secret", "rm", currentSecretName)
		err = osCmd.Run()
		if err != nil {
			log.Fatalln("Failed to run", "`"+strings.Join(osCmd.Args, " ")+"`", "with", err)
		}
		log.Println("Removed old secret", currentSecretName)

		if flagDockerContext != initialDockerContext {
			osCmd = exec.Command("docker", "context", "use", initialDockerContext)
			output, err = osCmd.CombinedOutput()
			if err != nil {
				log.Println(string(output))
				log.Fatal(err)
			}
			log.Default().Println("Returned docker context to initial value", initialDockerContext)
		}
	},
}

var (
	flagDockerContext string
	flagServiceName   string
)

func init() {
	secretCmd.Flags().StringVarP(&flagDockerContext, "docker-context", "c", config.GetConfigValue(config.ConfDockerContext), "The docker context to use")
	secretCmd.Flags().StringVarP(&flagServiceName, "service-name", "n", "mixtape_app", "The name of the service to update")
}

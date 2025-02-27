package deploy

import (
	"errors"
	"os/exec"
	"strings"
)

func GetCurrentDockerContext() (string, error) {
	osCmd := exec.Command("docker", "context", "show")
	outBuilder := strings.Builder{}
	errBuilder := strings.Builder{}
	osCmd.Stdout = &outBuilder
	osCmd.Stderr = &errBuilder
	err := osCmd.Run()
	if err != nil {
		return "", errors.New(errBuilder.String())
	}
	currentContext := strings.Trim(string(outBuilder.String()), "\n")
	return currentContext, nil
}

func SetDockerContext(context string) error {
	osCmd := exec.Command("docker", "context", "use", flagDockerContext)
	outBuilder := strings.Builder{}
	errBuilder := strings.Builder{}
	osCmd.Stdout = &outBuilder
	osCmd.Stderr = &errBuilder
	err := osCmd.Run()
	if err != nil {
		return errors.New(errBuilder.String())
	}
	return nil
}

func CreateDockerSecret(secretName, secretFile string) error {
	osCmd := exec.Command("docker", "secret", "create", secretName, secretFile)
	outBuilder := strings.Builder{}
	errBuilder := strings.Builder{}
	osCmd.Stdout = &outBuilder
	osCmd.Stderr = &errBuilder
	err := osCmd.Run()
	if err != nil {
		return errors.New(errBuilder.String())
	}
	return nil
}

func RemoveDockerSecret(secretName string) error {
	osCmd := exec.Command("docker", "secret", "rm", secretName)
	outBuilder := strings.Builder{}
	errBuilder := strings.Builder{}
	osCmd.Stdout = &outBuilder
	osCmd.Stderr = &errBuilder
	err := osCmd.Run()
	if err != nil {
		return errors.New(errBuilder.String())
	}
	return nil
}

func UpdateDockerServiceSecret(serviceName, rmSecretName, addSecretName, addSecretTarget string) error {
	osCmd := exec.Command("docker", "service", "update", serviceName, "--secret-rm", rmSecretName, "--secret-add", "source="+addSecretName+",target="+addSecretTarget)
	outBuilder := strings.Builder{}
	errBuilder := strings.Builder{}
	osCmd.Stdout = &outBuilder
	osCmd.Stderr = &errBuilder
	err := osCmd.Run()
	if err != nil {
		return errors.New(errBuilder.String())
	}
	return nil
}

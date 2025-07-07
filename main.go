package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/oullin/infra/api"
	"github.com/oullin/infra/pkg"
	"log"
	"os"
)

func main() {
	var err error
	err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	deployment, err := api.NewDeployment(api.DeploymentRequest{
		Command:        api.DeployCommand,
		ConfigFileName: pkg.Trim(os.Getenv("API_CONFIG_FILE_NAME")),
		ConfigFilePath: pkg.Trim(os.Getenv("API_CONFIG_FILE_PATH")),
	}, getValidator())

	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	if err := deployment.ReadDBSecrets(); err != nil {
		log.Fatal("Error reading DB secrets:", err)
	}

	fmt.Println("Username: ", deployment)
}

func getValidator() *validator.Validate {
	return validator.New(
		validator.WithRequiredStructEnabled(),
	)
}

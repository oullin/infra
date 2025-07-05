package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/oullin/infra/api"
	"log"
	"os"
)

func main() {
	var err error
	err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	deployer, err := api.NewDeployment(getValidator(), api.DeploymentRequest{
		SecretsDir:   os.Getenv("API_SECRETS_DIRECTORY"),
		ProjectDir:   os.Getenv("API_DIRECTORY"),
		CaddyLogsDir: os.Getenv("CADDY_LOGS_DIRECTORY"),
	})

	if err != nil {
		log.Fatal(err)
	}

	if err = deployer.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deployment completed.")
}

func getValidator() *validator.Validate {
	return validator.New(
		validator.WithRequiredStructEnabled(),
	)
}

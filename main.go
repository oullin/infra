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

	validate := validator.New(validator.WithRequiredStructEnabled())

	deployer := api.NewDeployment(validate, api.DeploymentRequest{
		SecretsDir: os.Getenv("API_SECRETS_DIRECTORY"),
		ProjectDir: os.Getenv("API_DIRECTORY"),
	})

	if err = deployer.Run(); err != nil {
		log.Fatal("Error running the deployment:", err)
	}

	fmt.Println(deployer)
}

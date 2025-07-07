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

	if err = godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	var ApiDeployment api.Deployment

	ApiDeployment, err = api.NewDeployment(api.DeploymentRequest{
		Command:        api.DeployCommand,
		ConfigFileName: pkg.Trim(api.ConfigFIleName),
		ConfigFilePath: pkg.Trim(os.Getenv("API_CONFIG_FILE_PATH")),
	}, getValidator())

	if err != nil {
		log.Fatal("Error create the deployment runner: ", err)
	}

	if err = ApiDeployment.ReadDBSecrets(); err != nil {
		log.Fatal("Error reading DB secrets:", err)
	}

	fmt.Println("Username: ", ApiDeployment)
}

func getValidator() *validator.Validate {
	return validator.New(
		validator.WithRequiredStructEnabled(),
	)
}

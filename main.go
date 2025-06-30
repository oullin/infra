package main

import (
	"fmt"
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

	deployer := api.NewAPIDeployment(api.DeploymentRequest{
		SecretsDir: os.Getenv("API_SECRETS_DIRECTORY"),
		ApiDir:     os.Getenv("API_DIRECTORY"),
	})

	if err = deployer.Run(); err != nil {
		log.Fatal("Error running the deployment:", err)
	}

	fmt.Println(deployer)
}

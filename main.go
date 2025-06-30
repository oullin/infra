package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/oullin/infra/cli"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	deployer := cli.NewAPIDeployment(os.Getenv("API_SECRETS_DIRECTORY"), os.Getenv("API_DIRECTORY"))

	fmt.Println(*deployer)
}

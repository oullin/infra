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

var validate *validator.Validate

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	validate = validator.New(
		validator.WithRequiredStructEnabled(),
	)
}

func main() {
	var err error
	var deployment api.Deployment

	deployment, err = api.NewDeployment(
		pkg.Trim(os.Getenv("API_CONFIG_FILE_PATH")),
		validate,
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = api.ParseDbCredentials(deployment); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Username: ", deployment)
}

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

var env *pkg.Env
var validate *validator.Validate

func init() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	validate = validator.New(
		validator.WithRequiredStructEnabled(),
	)

	wd, _ := os.Getwd()

	fmt.Println("Init: ", wd)

	env = &pkg.Env{
		ProjectRoot:       pkg.Trim(wd),
		AppEnv:            pkg.Trim(os.Getenv("APP_ENV")),
		ApiConfigFilePath: pkg.Trim(os.Getenv("API_CONFIG_FILE_PATH")),
	}

	if err := validate.Struct(env); err != nil {
		panic("Invalid app env: " + err.Error())
	}
}

func main() {
	deployment, err := api.NewDeployment(*env, *validate)

	if err != nil {
		log.Fatal(err)
	}

	if err = api.ParseDBSecrets(&deployment); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done ...")
}

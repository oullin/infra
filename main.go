package main

import (
    "fmt"
    "log"
    "os"

    "github.com/go-playground/validator/v10"
    "github.com/joho/godotenv"
    "github.com/oullin/infra/api"
    "github.com/oullin/infra/pkg"
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

    wd, err := os.Getwd()
    if err != nil {
        panic("Error getting working directory: " + err.Error())
    }

    env = &pkg.Env{
        ProjectRoot:       pkg.Trim(wd),
        AppEnv:            pkg.Trim(os.Getenv("APP_ENV")),
        ApiProjectRoot:    pkg.Trim(os.Getenv("API_DIRECTORY")),
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

    if err = deployment.ParseDBSecrets(); err != nil {
        log.Fatal(err)
    }

    if err = deployment.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Println(pkg.MagentaColour + "\nDeployment request sent successfully ..." + pkg.Reset)
}

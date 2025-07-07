package api

import "github.com/spf13/viper"

const DeployCommand = "make build:deploy"

type DeploymentRequest struct {
	ConfigFileName string
	ConfigFilePath string
	Command        string
}

type Deployment struct {
	Viper             *viper.Viper
	DBSecrets         *DBSecrets
	DeploymentRequest DeploymentRequest
}

type DBSecrets struct {
	DbName       string `validate:"required"`
	DbNameFile   string `validate:"required"`
	UserName     string `validate:"required"`
	UserNameFile string `validate:"required"`
	Password     string `validate:"required"`
	PasswordFile string `validate:"required"`
}

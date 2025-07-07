package api

import (
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
)

const ConfigFIleName = "api"
const ConfigFIleType = "yaml"
const DeployCommand = "build:deploy"
const DBNameFileName = "database.secrets.pg_dbname"
const DBUserNameFileName = "database.secrets.pg_username"
const DBPasswordFileName = "database.secrets.pg_password"

type Deployment struct {
	Env            *pkg.Env
	Viper          *viper.Viper `validate:"required"`
	DBSecrets      *DBSecrets
	ConfigFileName string `validate:"required"`
	ConfigFilePath string `validate:"required"`
	Command        string `validate:"required"`
}

type DBSecrets struct {
	DbName       string `validate:"required"`
	DbNameFile   string `validate:"required"`
	UserName     string `validate:"required"`
	UserNameFile string `validate:"required"`
	Password     string `validate:"required"`
	PasswordFile string `validate:"required"`
}

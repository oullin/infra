package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
)

func NewDeployment(env pkg.Env, validator validator.Validate) (Deployment, error) {
	var deployment Deployment

	fmt.Println("NewDeployment: ", env.ProjectRoot)

	request := DeploymentRequest{
		Command:        DeployCommand,
		ConfigFileName: pkg.Trim(ConfigFIleName),
		ConfigFilePath: env.GetApiConfigFilePath(),
	}

	if err := validator.Struct(request); err != nil {
		return deployment, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
	}

	viper.AddConfigPath(env.GetApiConfigFilePath())
	viper.SetConfigName(request.ConfigFileName)
	viper.SetConfigType(ConfigFIleType)

	if err := viper.ReadInConfig(); err != nil {
		return deployment, fmt.Errorf("[api] error reading config file: %w", err)
	}

	deployment.DeploymentRequest = &request
	deployment.Viper = viper.GetViper()
	deployment.DBSecrets = nil
	deployment.Env = &env

	if err := validator.Struct(deployment); err != nil {
		return deployment, fmt.Errorf("invalid deployment runner [%#v]: %v", request, err)
	}

	return deployment, nil
}

func ParseDBSecrets(deployment *Deployment) error {
	dbSecrets := DBSecrets{}

	namespace, fullPath := deployment.GetDbNamePair()
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the db name file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.DbName = dbName
		dbSecrets.DbNameFile = namespace
	}

	namespace, fullPath = deployment.GetDbUserNamePair()
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the username file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.UserName = dbName
		dbSecrets.UserNameFile = namespace
	}

	namespace, fullPath = deployment.GetDbPasswordPair()
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the password file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.Password = dbName
		dbSecrets.PasswordFile = namespace
	}

	return nil
}

package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
	"strings"
)

func NewDeployment(env pkg.Env, validator validator.Validate) (Deployment, error) {
	var deployment Deployment

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

func (d *Deployment) ParseDBSecrets() error {
	dbSecrets := DBSecrets{}

	namespace, fullPath := d.GetDirectoryPair(DBNameFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the db name file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.DbName = dbName
		dbSecrets.DbNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBUserNameFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the username file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.UserName = dbName
		dbSecrets.UserNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBPasswordFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[parser] error reading the password file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.Password = dbName
		dbSecrets.PasswordFile = namespace
	}

	d.DBSecrets = &dbSecrets

	return nil
}

// GetDirectoryPair (namespace, fullPath)
func (d *Deployment) GetDirectoryPair(seed string) (string, string) {
	namespace := strings.Trim(d.Viper.GetString(seed), "/")

	if d.Env.IsProduction() {
		return namespace, d.Viper.GetString(seed)
	}

	fullPath := d.Env.GetProjectRoot() + "/" + namespace

	return namespace, fullPath
}

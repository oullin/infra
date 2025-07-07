package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
)

func NewDeployment(configFilePath string, validator *validator.Validate) (Deployment, error) {
	var deployment Deployment

	request := DeploymentRequest{
		Command:        DeployCommand,
		ConfigFileName: pkg.Trim(ConfigFIleName),
		ConfigFilePath: pkg.Trim(configFilePath),
	}

	if err := validator.Struct(request); err != nil {
		return deployment, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
	}

	viper.SetConfigName(request.ConfigFileName)
	viper.AddConfigPath(request.ConfigFilePath)
	viper.SetConfigType(ConfigFIleType)

	if err := viper.ReadInConfig(); err != nil {
		return deployment, fmt.Errorf("[api] error reading config file: %w", err)
	}

	deployment.DeploymentRequest = &request
	deployment.Viper = viper.GetViper()
	deployment.DBSecrets = nil

	if err := validator.Struct(deployment); err != nil {
		return deployment, fmt.Errorf("invalid deployment runner [%#v]: %v", request, err)
	}

	return deployment, nil
}

func Build(deployment Deployment) error {
	return nil
}

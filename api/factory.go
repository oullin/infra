package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

func NewDeployment(request DeploymentRequest, validator *validator.Validate) (Deployment, error) {
	var deployment Deployment

	if err := validator.Struct(request); err != nil {
		return deployment, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
	}

	viper.SetConfigName(request.ConfigFileName)
	viper.AddConfigPath(request.ConfigFilePath)
	viper.SetConfigType(ConfigFIleType)

	if err := viper.ReadInConfig(); err != nil {
		return deployment, fmt.Errorf("error reading config file: %w", err)
	}

	deployment.DeploymentRequest = &request
	deployment.Viper = viper.GetViper()
	deployment.DBSecrets = nil

	return deployment, nil
}

func Build(deployment Deployment) error {
	return nil
}

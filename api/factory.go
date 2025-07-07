package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

func NewDeployment(request DeploymentRequest, validator *validator.Validate) (Deployment, error) {
	if err := validator.Struct(request); err != nil {
		return Deployment{}, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
	}

	//viper.SetConfigName(request.ConfigFileName)
	//viper.SetConfigName(request.ConfigFilePath)
	//viper.SetConfigFile("yaml")
	//viper.GetViper()

	return Deployment{
		DeploymentRequest: request,
		Viper:             nil,
		DBSecrets:         nil,
	}, nil
}

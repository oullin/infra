package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

func NewDeployment(request DeploymentRequest, validator *validator.Validate) (*Deployment, error) {
	if err := validator.Struct(request); err != nil {
		return nil, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
	}

	viper.SetConfigName(request.ConfigFileName)
	viper.SetConfigName(request.ConfigFilePath)
	viper.SetConfigFile("yaml")
	viper.GetViper()

	return &Deployment{
		Viper:     viper.GetViper(),
		DBSecrets: DBSecrets{},
	}, nil
}

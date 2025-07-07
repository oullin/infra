package api

import (
	"fmt"
	"github.com/spf13/viper"
)

func (d *Deployment) ReadDBSecrets() error {
	//viper := d.Viper
	//secrets := DBSecrets{}

	//fmt.Println(d.)
	fmt.Println(d.DeploymentRequest.ConfigFileName)
	fmt.Println(d.DeploymentRequest.ConfigFilePath)

	viper.SetConfigName(d.DeploymentRequest.ConfigFileName)
	viper.AddConfigPath(d.DeploymentRequest.ConfigFilePath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	fmt.Println(viper.GetStringMap("database"))
	panic("\n-------------")

	//file := viper.GetString("database.secrets.pg_dbname")
	//fmt.Println("File: ", file)
	//
	//value, err := pkg.GetFileContent(file)
	//
	//if err != nil {
	//	return err
	//}
	//
	//secrets.DbNameFile = file
	//secrets.DbName = value
	//
	//d.DBSecrets = &secrets

	return nil
}

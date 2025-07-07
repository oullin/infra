package api

import (
	"fmt"
)

func (d *Deployment) ReadDBSecrets() error {
	viper := d.Viper
	//secrets := DBSecrets{}

	fmt.Println(viper.GetStringMap("database"))
	fmt.Println(viper.GetString("database.secrets.pg_dbname"))

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

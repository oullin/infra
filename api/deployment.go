package api

import (
	"strings"
)

// GetDbNamePair (namespace, fullPath)
func (d Deployment) GetDbNamePair() (string, string) {
	namespace := strings.Trim(d.Viper.GetString(DBNameFileName), "/")

	if d.Env.IsProduction() {
		return namespace, d.Viper.GetString(DBNameFileName)
	}

	fullPath := d.Env.GetProjectRoot() + "/" + namespace

	return namespace, fullPath
}

// GetDbUserNamePair (namespace, fullPath)
func (d Deployment) GetDbUserNamePair() (string, string) {
	namespace := strings.Trim(d.Viper.GetString(DBUserNameFileName), "/")

	if d.Env.IsProduction() {
		return namespace, d.Viper.GetString(DBUserNameFileName)
	}

	fullPath := d.Env.GetProjectRoot() + "/" + namespace

	return namespace, fullPath
}

// GetDbPasswordPair (namespace, fullPath)
func (d Deployment) GetDbPasswordPair() (string, string) {
	namespace := strings.Trim(d.Viper.GetString(DBPasswordFileName), "/")

	if d.Env.IsProduction() {
		return namespace, d.Viper.GetString(DBPasswordFileName)
	}

	fullPath := d.Env.GetProjectRoot() + "/" + namespace

	return namespace, fullPath
}

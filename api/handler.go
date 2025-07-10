package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NewDeployment(env pkg.Env, validator validator.Validate) (Deployment, error) {
	deployment := Deployment{
		ConfigFilePath: env.GetApiConfigFilePath(),
		ConfigFileName: pkg.Trim(ConfigFileName),
		Command:        DeployCommand,
		Env:            &env,
		DBSecrets:      nil,
	}

	viper.SetConfigType(ConfigFileType)
	viper.SetConfigName(deployment.ConfigFileName)
	viper.AddConfigPath(deployment.ConfigFilePath)

	deployment.Viper = viper.GetViper()

	if err := deployment.Viper.ReadInConfig(); err != nil {
		return deployment, fmt.Errorf("[api] error reading config file: %w", err)
	}

	if err := validator.Struct(deployment); err != nil {
		return deployment, fmt.Errorf("[api] invalid deployment runner [%#v]: %v", deployment, err)
	}

	return deployment, nil
}

func (d *Deployment) ParseDBSecrets() error {
	dbSecrets := DBSecrets{}

	namespace, fullPath := d.GetDirectoryPair(DBNameFileName)
	if value, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the db name file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.DbName = value
		dbSecrets.DbNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBUserNameFileName)
	if value, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the username file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.UserName = value
		dbSecrets.UserNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBPasswordFileName)
	if value, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the password file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.Password = value
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

	fullPath := filepath.Join(d.Env.ProjectRoot, namespace)

	return namespace, fullPath
}

func (d *Deployment) Run() error {
	cmd := exec.Command("make", d.GetCommandArgs()...)

	// Pass the parent environment to the child process.
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error: 'make %v' command failed: %v\n", d.GetCommandArgs(), err)
	}

	return nil
}

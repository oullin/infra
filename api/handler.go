package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/oullin/infra/pkg"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
)

func NewDeployment(env pkg.Env, validator validator.Validate) (Deployment, error) {
	deployment := Deployment{
		ConfigFilePath: env.GetApiConfigFilePath(),
		ConfigFileName: pkg.Trim(ConfigFIleName),
		Command:        DeployCommand,
		Env:            &env,
		DBSecrets:      nil,
	}

	viper.SetConfigType(ConfigFIleType)
	viper.SetConfigName(deployment.ConfigFileName)
	viper.AddConfigPath(deployment.ConfigFilePath)

	if err := viper.ReadInConfig(); err != nil {
		return deployment, fmt.Errorf("[api] error reading config file: %w", err)
	}

	deployment.Viper = viper.GetViper()
	if err := validator.Struct(deployment); err != nil {
		return deployment, fmt.Errorf("[api] invalid deployment runner [%#v]: %v", deployment, err)
	}

	return deployment, nil
}

func (d *Deployment) ParseDBSecrets() error {
	dbSecrets := DBSecrets{}

	namespace, fullPath := d.GetDirectoryPair(DBNameFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the db name file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.DbName = dbName
		dbSecrets.DbNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBUserNameFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the username file [%s]: %v", fullPath, err)
	} else {
		dbSecrets.UserName = dbName
		dbSecrets.UserNameFile = namespace
	}

	namespace, fullPath = d.GetDirectoryPair(DBPasswordFileName)
	if dbName, err := pkg.GetFileContent(fullPath); err != nil {
		return fmt.Errorf("[api]  error parsing the password file [%s]: %v", fullPath, err)
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

func (d *Deployment) Run() error {
	projectRoot := d.Env.GetProjectRoot()
	fmt.Printf("\n ---> Run: Root directory: %#v\n", projectRoot)

	cmd := exec.Command("make", d.GetCommandArgs()...)

	// Pass the parent environment to the child process.
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error: 'make -C %s build:prod' command failed: %v\n", projectRoot, err)
	}

	return nil
}

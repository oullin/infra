package api

import (
	"fmt"
	"github.com/oullin/infra/pkg"
	"path/filepath"
	"strings"
)

func (d *Deployment) GetCommandArgs() []string {
	if d.Env.IsProduction() {
		return d.GetProdCommand()
	}

	return d.GetTestingCommand()
}

func (d *Deployment) GetProdCommand() []string {
	dbUsernameFile := filepath.Join("/", strings.TrimLeft(d.DBSecrets.UserNameFile, "/"))
	dbPasswordFile := filepath.Join("/", strings.TrimLeft(d.DBSecrets.PasswordFile, "/"))
	dbNameFile := filepath.Join("/", strings.TrimLeft(d.DBSecrets.DbNameFile, "/"))

	args := []string{
		"-C",
		d.Env.ApiProjectRoot,
		d.Command,
		fmt.Sprintf("DB_SECRET_USERNAME=%s", dbUsernameFile),
		fmt.Sprintf("DB_SECRET_PASSWORD=%s", dbPasswordFile),
		fmt.Sprintf("DB_SECRET_DBNAME=%s", dbNameFile),
	}

	PrintArgs(args)

	return args
}

func (d *Deployment) GetTestingCommand() []string {
	args := []string{
		"-C",
		d.Env.ProjectRoot,
		"build-test",
		fmt.Sprintf("DB_SECRET_USERNAME=%s", d.DBSecrets.UserNameFile),
		fmt.Sprintf("DB_SECRET_PASSWORD=%s", d.DBSecrets.PasswordFile),
		fmt.Sprintf("DB_SECRET_DBNAME=%s", d.DBSecrets.DbNameFile),
		fmt.Sprintf("OTHER_DB_USER_NAME=%s", d.DBSecrets.UserName),
		fmt.Sprintf("OTHER_DB_USER_PASSWORD=%s", d.DBSecrets.Password),
		fmt.Sprintf("OTHER_DB_DATABASE_NAME=%s", d.DBSecrets.DbName),
	}

	PrintArgs(args)

	return args
}

func PrintArgs(args []string) {
	fmt.Printf("\n--- Make Command: ")
	fmt.Printf(pkg.CyanColour+"\n%#v\n\n"+pkg.Reset, args)
}

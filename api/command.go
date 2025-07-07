package api

import "fmt"

func (d *Deployment) GetCommandArgs() []string {
	if d.Env.IsProduction() {
		return d.ResolveCommandFor(d.Env.ApiProjectRoot, d.Command)
	}

	return d.ResolveCommandFor(d.Env.ProjectRoot, "build-test")
}

func (d *Deployment) ResolveCommandFor(directory string, command string) []string {
	args := []string{
		"-C",
		directory,
		command,
		fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", d.DBSecrets.UserNameFile),
		fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", d.DBSecrets.PasswordFile),
		fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", d.DBSecrets.DbNameFile),
		fmt.Sprintf("ENV_DB_USER_NAME=%s", d.DBSecrets.UserName),
		fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", d.DBSecrets.Password),
		fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", d.DBSecrets.DbName),
	}

	fmt.Printf("\n ---> Command: %#v\n", args)

	return args
}

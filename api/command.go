package api

import "fmt"

func (d *Deployment) GetCommandArgs() []string {
    if d.Env.IsProduction() {
        return d.GetProdCommandArgs()
    }

    return d.GetTestingCommandArgs()
}

func (d *Deployment) GetProdCommandArgs() []string {
    args := []string{
        "-C",
        d.Env.GetProjectRoot(),
        d.Command,
        fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", d.DBSecrets.UserNameFile),
        fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", d.DBSecrets.PasswordFile),
        fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", d.DBSecrets.DbNameFile),
        fmt.Sprintf("ENV_DB_USER_NAME=%s", d.DBSecrets.UserName),
        fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", d.DBSecrets.Password),
        fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", d.DBSecrets.DbName),
    }

    fmt.Printf("\n ---> Prod Command: %#v\n", args)

    return args
}

func (d *Deployment) GetTestingCommandArgs() []string {
    args := []string{
        "-C",
        d.Env.GetProjectRoot(),
        "build-test",
        fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", d.DBSecrets.UserNameFile),
        fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", d.DBSecrets.PasswordFile),
        fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", d.DBSecrets.DbNameFile),
        fmt.Sprintf("ENV_DB_USER_NAME=%s", d.DBSecrets.UserName),
        fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", d.DBSecrets.Password),
        fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", d.DBSecrets.DbName),
    }

    fmt.Printf("\n ---> Test Command: %#v\n", args)

    return args
}

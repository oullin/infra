package api

import (
    "fmt"
    "github.com/go-playground/validator/v10"
    "github.com/oullin/infra/pkg"
    "os"
    "os/exec"
    "path/filepath"
)

func NewDeployment(validator *validator.Validate, request DeploymentRequest) (*Deployment, error) {
    if err := validator.Struct(request); err != nil {
        return nil, fmt.Errorf("invalid deployment request [%#v]: %v", request, err)
    }

    fmt.Printf("\n ---> Init: ProjectDir: %#v", request.ProjectDir)
    fmt.Printf("\n ---> Init: SecretsDir: %#v", request.SecretsDir)
    fmt.Printf("\n ---> Init: CaddyLogsPath: %#v", request.CaddyLogsDir)

    projectDir := pkg.Trim(request.ProjectDir)
    secretsDir := pkg.Trim(request.SecretsDir)

    dbSecretFiles := DbSecretFiles{
        dbName: pkg.Trim(filepath.Join(secretsDir, dbNameFileName)),
        dbUser: pkg.Trim(filepath.Join(secretsDir, dbUserNameFileName)),
        dbPass: pkg.Trim(filepath.Join(secretsDir, dbPasswordFileName)),
    }

    if err := assertSecretFiles(dbSecretFiles); err != nil {
        return nil, err
    }

    var dbSecrets DbSecrets
    if err := parseDbSecrets(dbSecretFiles, &dbSecrets); err != nil {
        return nil, err
    }

    deployment := Deployment{
        projectDir:   projectDir,
        secretsDir:   secretsDir,
        caddyLogsDir: pkg.Trim(request.CaddyLogsDir),
        dbSecrets:    dbSecrets,
        dbSecretFile: dbSecretFiles,
    }

    if err := validator.Struct(deployment); err != nil {
        return nil, fmt.Errorf("invalid deployment command [%#v]: %v", deployment, err)
    }

    return &deployment, nil
}

func (d Deployment) Run() error {
    projectRoot := pkg.Trim(d.projectDir)
    fmt.Printf("\n ---> Run: Root directory: %#v\n", projectRoot)

    makeArgs := []string{
        "-C",
        projectRoot,
        "build:prod",
        fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", d.dbSecretFile.dbUser),
        fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", d.dbSecretFile.dbPass),
        fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", d.dbSecretFile.dbName),
        fmt.Sprintf("ENV_DB_USER_NAME=%s", d.dbSecrets.dbUser),
        fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", d.dbSecrets.dbPass),
        fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", d.dbSecrets.dbName),
        fmt.Sprintf("CADDY_LOGS_PATH=%s", d.caddyLogsDir),
    }

    fmt.Printf("\n ---> Run: makeArgs: %#v\n", makeArgs)

    cmd := exec.Command("make", makeArgs...)

    // Pass the parent environment to the child process.
    cmd.Env = os.Environ()
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Error: 'make -C %s build:prod' command failed: %v\n", projectRoot, err)
    }

    return nil
}

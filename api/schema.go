package api

const dbNameFileName = "postgres_db"
const dbUserNameFileName = "postgres_user"
const dbPasswordFileName = "postgres_password"

type DeploymentScript struct {
	CredentialsFile string `validate:"required"`
	FileName        string `validate:"required"`
	Extension       string `validate:"required"`
}

//----- OLD

type DeploymentRequest struct {
	SecretsDir         string `validate:"required"`
	ProjectDir         string `validate:"required"`
	CaddyLogsDir       string `validate:"required"`
	ApiCredentialsFile string `validate:"required"`
}

type Deployment struct {
	secretsDir   string        `validate:"required"`
	projectDir   string        `validate:"required"`
	caddyLogsDir string        `validate:"required"`
	dbSecrets    DbSecrets     `validate:"required"`
	dbSecretFile DbSecretFiles `validate:"required"`
}

type DbSecrets struct {
	dbUser string
	dbName string
	dbPass string
}

type DbSecretFiles struct {
	dbUser string
	dbName string
	dbPass string
}

type Commands struct {
	signature string
	directory string
}

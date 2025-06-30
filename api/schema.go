package api

type DeploymentRequest struct {
	SecretsDir string
	ApiDir     string
}

type Deployment struct {
	secretsDir   string
	apiDir       string
	projectRoot  string
	apiDBSecrets DbSecrets
	commands     Commands
}

type DbSecrets struct {
	userSecretFile        string
	userSecretContent     string
	passwordSecretFile    string
	passwordSecretContent string
	dbSecretFile          string
	dbSecretContent       string
}

type Commands struct {
	Production string
}

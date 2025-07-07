package pkg

type Env struct {
	AppEnv            string `validate:"required,min=5"`
	ProjectRoot       string `validate:"required,min=5"`
	ApiConfigFilePath string `validate:"required,min=5"`
}

func (e Env) IsDev() bool {
	return !e.IsProduction()
}

func (e Env) IsProduction() bool {
	return e.AppEnv == "production"
}

func (e Env) GetProjectRoot() string {
	return e.ProjectRoot
}

func (e Env) GetApiConfigFilePath() string {
	if e.IsProduction() {
		return e.ApiConfigFilePath
	}

	// Testing file.
	return e.ProjectRoot + "/storage/api/"
}

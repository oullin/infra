package pkg

import "path/filepath"

type Env struct {
	AppEnv            string `validate:"required,min=5"`
	ProjectRoot       string `validate:"required,min=5"`
	ApiProjectRoot    string `validate:"required,min=3"`
	ApiConfigFilePath string `validate:"required,min=5"`
}

func (e Env) IsDev() bool {
	return !e.IsProduction()
}

func (e Env) IsProduction() bool {
	return e.AppEnv == "production"
}

func (e Env) GetApiConfigFilePath() string {
	if e.IsProduction() {
		return e.ApiConfigFilePath
	}

	// Testing file.
	return filepath.Join(e.ProjectRoot, "storage", "api") + "/"
}

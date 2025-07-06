package pkg

type Env struct {
	AppMachine string
}

func (e Env) IsDev() bool {
	return !e.IsProduction()
}

func (e Env) IsProduction() bool {
	return e.AppMachine == "production"
}

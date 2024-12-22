package application

import "github.com/festy23/GoCalcServer/internal/backend"

type Application struct {
}

func New() *Application {
	return &Application{}
}

func (a *Application) Run() {
	backend.StartServer()
}

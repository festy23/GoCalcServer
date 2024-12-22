package main

import "github.com/festy23/GoCalcServer/internal/application"

// Запуск приложения
func main() {
	app := application.New()
	app.Run()
}

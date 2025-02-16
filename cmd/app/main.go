package main

import (
	"avito-merch/internal/app"
	"avito-merch/internal/config"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Создаем и запускаем приложение
	app := app.NewApp(cfg)
	app.Run()
}

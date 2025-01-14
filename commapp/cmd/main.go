package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"APIGetaway/pkg/api"
	"APIGetaway/pkg/migrations"
	"APIGetaway/pkg/storage"
)

// конфигурация приложения
type config struct {
	DB storage.DBConfig `json:"db"`
}

func main() {
	// чтение и раскодирование файла конфигурации
	b, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config config
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal(err)
	}
	// инициализация зависимостей приложения
	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", config.DB.Host, config.DB.User, config.DB.Password, config.DB.DBName, config.DB.Port, config.DB.SSLMode)

	db, err := storage.New(dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	migrations.RunMigrations(dbInfo)
	api := api.New(db)

	// запуск веб-сервера с API и приложением
	err = http.ListenAndServe(":8082", api.Router())
	if err != nil {
		log.Fatal(err)
	}
}

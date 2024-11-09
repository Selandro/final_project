package main

import (
	"log"
	"net/http"

	"APIGetaway/pkg/api"
)

func main() {
	// Создаем новый API
	api := api.New()

	// Запуск HTTP сервера
	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", api.Router())
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}

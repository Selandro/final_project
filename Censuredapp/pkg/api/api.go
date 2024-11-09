package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Список запрещенных слов
var forbiddenWords = []string{"qwerty", "йцукен", "zxvbnm"}

// API приложения GoNews.
type API struct {
	r *chi.Mux
}

// Конструктор API.
func New() *API {
	// Инициализируем логгер
	initLogger()

	a := API{r: chi.NewRouter()}
	a.endpoints()
	return &a
}

// Router возвращает маршрутизатор для использования
// в качестве аргумента HTTP-сервера.
func (api *API) Router() *chi.Mux {
	return api.r
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {

	// Middleware для логирования запросов
	api.r.Use(RequestIDMiddleware) // Добавляем middleware для request_id
	api.r.Use(LoggingMiddleware)   // Добавляем middleware для логирования

	api.r.Post("/comments", api.Censored)
}

// Censored - обработчик POST запроса для добавления комментария.
func (api *API) Censored(w http.ResponseWriter, r *http.Request) {
	// Чтение тела запроса (комментарий)
	var requestData map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Проверка на наличие текста комментария
	text, exists := requestData["text"]
	if !exists || text == "" {
		http.Error(w, "Текст комментария не может быть пустым", http.StatusBadRequest)
		return
	}

	// Валидация: проверка на наличие запрещенных слов
	if isValidComment(text) {
		// Успешная валидация, возвращаем статус 200
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Комментарий добавлен"))
	} else {
		// Комментарий не прошел валидацию
		http.Error(w, "Комментарий содержит недопустимые слова", http.StatusBadRequest)
	}
}

// Функция для проверки комментария на наличие запрещенных слов
func isValidComment(text string) bool {
	for _, word := range forbiddenWords {
		if strings.Contains(strings.ToLower(text), word) {
			log.Printf("Комментарий содержит недопустимое слово: %s", word)
			return false
		}
	}
	return true
}

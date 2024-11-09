package api

import (
	"APIGetaway/pkg/models"
	"APIGetaway/pkg/storage"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// API структура.
type API struct {
	db storage.DBInterface
	r  *chi.Mux
}

// Конструктор API.
func New(db storage.DBInterface) *API {
	// Инициализируем логгер
	initLogger()
	a := API{db: db, r: chi.NewRouter()}
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
	api.r.Post("/comments", api.addCommentHandler)
	api.r.Get("/comments", api.getCommentsHandler)
}

// Обработчик для добавления комментария.
func (api *API) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Извлекаем request_id из контекста запроса
	requestID, ok := r.Context().Value(RequestIDKey{}).(string)
	if !ok {
		fmt.Println()
	}

	// Отправляем запрос к сервису цензуры для проверки текста комментария
	censorPassed, err := api.checkCommentWithCensorshipService(comment.Text, requestID)
	if err != nil {
		http.Error(w, "ошибка проверки цензуры", http.StatusInternalServerError)
		return
	}
	if !censorPassed {
		http.Error(w, "Комментарий содержит недопустимые слова", http.StatusBadRequest)
		return
	}

	// Устанавливаем статус и время создания по умолчанию
	comment.CreatedAt = time.Now()

	// Сохранение комментария в базе данных
	id, err := api.db.AddComment(context.Background(), comment)
	if err != nil {
		http.Error(w, "не удалось добавить комментарий", http.StatusInternalServerError)
		return
	}

	// Возвращаем ID созданного комментария
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// checkCommentWithCensorshipService отправляет запрос к сервису цензуры с request_id.
func (api *API) checkCommentWithCensorshipService(text, requestID string) (bool, error) {
	// Адрес сервиса цензуры
	censorshipServiceURL := fmt.Sprintf("http://localhost:8083/comments?request_id=%s", requestID)

	// Формируем JSON-запрос
	requestBody, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return false, err
	}

	// Создаем запрос к сервису цензуры
	req, err := http.NewRequest("POST", censorshipServiceURL, strings.NewReader(string(requestBody)))
	if err != nil {
		return false, err
	}

	// Устанавливаем заголовок Content-Type и request_id
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("request_id", requestID)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Обрабатываем ответ от сервиса цензуры
	if resp.StatusCode == http.StatusOK {
		return true, nil // Комментарий прошел цензуру
	} else if resp.StatusCode == http.StatusBadRequest {
		return false, nil // Комментарий не прошел цензуру
	}

	return false, fmt.Errorf("непредвиденный статус ответа от сервиса цензуры: %d", resp.StatusCode)
}

// Обработчик для получения комментариев по ID новости.
func (api *API) getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	newsIDParam := r.URL.Query().Get("news_id")
	if newsIDParam == "" {
		http.Error(w, "news_id является обязательным параметром", http.StatusBadRequest)
		return
	}

	// Преобразование ID новости из строки в число
	newsID, err := strconv.ParseInt(newsIDParam, 10, 64)
	if err != nil {
		http.Error(w, "неверный формат news_id", http.StatusBadRequest)
		return
	}

	// Получение комментариев из базы данных
	comments, err := api.db.GetCommentsByNewsID(context.Background(), newsID)
	if err != nil {
		http.Error(w, "не удалось получить комментарии", http.StatusInternalServerError)
		return
	}

	// Возвращаем список комментариев в формате JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

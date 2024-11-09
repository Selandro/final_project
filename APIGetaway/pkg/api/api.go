package api

import (
	"APIGetaway/pkg/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Конфигурация адресов сервисов
const (
	newsServiceURL     = "http://localhost:8081" // адрес сервиса новостей
	commentsServiceURL = "http://localhost:8082" // адрес сервиса комментариев
)

// API структура.
type API struct {
	r *chi.Mux
}

// Конструктор API.
func New() *API {
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
	api.r.Use(RequestIDMiddleware)
	api.r.Use(LoggingMiddleware)
	api.r.Use(middleware.Recoverer)

	api.r.Get("/news", api.getAllNews)
	api.r.Get("/news/filter", api.filterNews)
	api.r.Get("/news/{id}", api.getNewsByID)
	api.r.Post("/news/{id}/comment", api.addComment)
}

// Получить список всех новостей с пагинацией и фильтрацией.
func (api *API) getAllNews(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры для фильтрации и пагинации
	searchTerm := r.URL.Query().Get("s") // Параметр для поиска
	pageStr := r.URL.Query().Get("page") // Параметр для страницы
	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	// Формируем базовый URL для запроса
	baseURL := fmt.Sprintf("%s/news", newsServiceURL)

	// Добавляем параметры в строку запроса, если они заданы
	url := baseURL
	if searchTerm != "" {
		url += fmt.Sprintf("?s=%s", searchTerm)
	}
	if page > 0 {
		if searchTerm != "" {
			url += fmt.Sprintf("&page=%d", page)
		} else {
			url += fmt.Sprintf("?page=%d", page)
		}
	}

	// Создание запроса к новостному сервису
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "Не удалось создать запрос для получения списка новостей", http.StatusInternalServerError)
		return
	}

	// Инициализация HTTP-клиента
	client := &http.Client{}
	respChan := make(chan *http.Response, 1)
	errChan := make(chan error, 1)

	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		respChan <- resp
	}()

	select {
	case err := <-errChan:
		http.Error(w, fmt.Sprintf("Не удалось получить список новостей: %v", err), http.StatusInternalServerError)
		return
	case resp := <-respChan:
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

// Фильтрация новостей.
func (api *API) filterNews(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Encode()

	// Создание запроса к новостному сервису с параметрами фильтрации
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/news?%s", newsServiceURL, query), nil) // Изменение на /news
	if err != nil {
		http.Error(w, "Не удалось создать запрос для фильтрации новостей", http.StatusInternalServerError)
		return
	}

	// Добавляем request_id в заголовок запроса
	requestID, _ := r.Context().Value(RequestIDKey{}).(string)
	req.Header.Set("request_id", requestID)

	// Инициализация HTTP-клиента
	client := &http.Client{}
	respChan := make(chan *http.Response, 1)
	errChan := make(chan error, 1)

	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		respChan <- resp
	}()

	select {
	case err := <-errChan:
		http.Error(w, fmt.Sprintf("Не удалось отфильтровать новости: %v", err), http.StatusInternalServerError)
		return
	case resp := <-respChan:
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

// Получить детальную информацию о новости по ID.
func (api *API) getNewsByID(w http.ResponseWriter, r *http.Request) {
	newsID := chi.URLParam(r, "id")

	// Канал для получения результатов
	newsCh := make(chan models.NewsFullDetailed)
	commentsCh := make(chan []models.Comment)
	errCh := make(chan error, 2) // Канал для ошибок

	// Функция для получения новости
	go func() {
		reqNews, err := http.NewRequest("GET", fmt.Sprintf("%s/news/%s", newsServiceURL, newsID), nil)
		if err != nil {
			errCh <- fmt.Errorf("не удалось создать запрос для получения новости: %w", err)
			return
		}
		requestID, _ := r.Context().Value(RequestIDKey{}).(string)
		reqNews.Header.Set("request_id", requestID)

		client := &http.Client{}
		respNews, err := client.Do(reqNews)
		if err != nil || respNews.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("не удалось получить новость: %w", err)
			return
		}
		defer respNews.Body.Close()

		var news models.NewsFullDetailed
		if err := json.NewDecoder(respNews.Body).Decode(&news); err != nil {
			errCh <- fmt.Errorf("ошибка при декодировании новости: %w", err)
			return
		}
		newsCh <- news // Отправляем новость в канал
	}()

	// Функция для получения комментариев
	go func() {
		reqComments, err := http.NewRequest("GET", fmt.Sprintf("%s/comments?news_id=%s", commentsServiceURL, newsID), nil)
		if err != nil {
			errCh <- fmt.Errorf("не удалось создать запрос для получения комментариев: %w", err)
			return
		}

		client := &http.Client{}
		respComments, err := client.Do(reqComments)
		if err != nil || respComments.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("не удалось получить комментарии: %w", err)
			return
		}
		defer respComments.Body.Close()

		var comments []models.Comment
		if err := json.NewDecoder(respComments.Body).Decode(&comments); err != nil {
			errCh <- fmt.Errorf("ошибка при декодировании комментариев: %w", err)
			return
		}
		commentsCh <- comments // Отправляем комментарии в канал
	}()

	// Ожидаем результатов
	select {
	case news := <-newsCh:
		// Получили новость
		select {
		case comments := <-commentsCh:
			// Получили комментарии
			news.Comments = comments
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(news)
		case err := <-errCh:
			// Ошибка при получении комментариев
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case err := <-errCh:
		// Ошибка при получении новости
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Добавить комментарий к новости.
func (api *API) addComment(w http.ResponseWriter, r *http.Request) {
	newsID := chi.URLParam(r, "id")
	var comment models.Comment

	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(newsID) // Преобразование строки в int
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	comment.NewsID = int64(id) // Теперь можно присвоить `id`, который является int

	// Прокси-запрос на сервис комментариев
	// Мы повторно создаем тело запроса с нужным форматом
	commentBody, err := json.Marshal(comment)
	if err != nil {
		http.Error(w, "Ошибка при обработке комментария", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/comments", commentsServiceURL), strings.NewReader(string(commentBody)))
	if err != nil {
		http.Error(w, "Не удалось создать запрос для добавления комментария", http.StatusInternalServerError)
		return
	}

	// Добавляем request_id в заголовок запроса
	requestID, _ := r.Context().Value(RequestIDKey{}).(string)
	req.Header.Set("request_id", requestID)
	// Установка заголовков
	req.Header.Set("Content-Type", "application/json")

	// Инициализация HTTP-клиента
	client := &http.Client{}
	respChan := make(chan *http.Response, 1)
	errChan := make(chan error, 1)

	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		respChan <- resp
	}()

	select {
	case err := <-errChan:
		http.Error(w, fmt.Sprintf("Не удалось добавить комментарий: %v", err), http.StatusInternalServerError)
		return
	case resp := <-respChan:
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

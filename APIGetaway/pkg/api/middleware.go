package api

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// RequestIDKey — тип для ключа контекста
type RequestIDKey struct{}

// Создаем логгер для записи в файл и в stdout
var logFile *os.File
var logger *log.Logger

// Инициализация логгера
func initLogger() {
	var err error
	// Открытие или создание файла логов
	logFile, err = os.OpenFile("access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Создаем MultiWriter для записи в файл и в stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Создаем новый логгер с использованием MultiWriter
	logger = log.New(multiWriter, "", log.LstdFlags)
}

// Middleware для добавления request_id в контекст запроса и заголовок ответа
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем `request_id` из заголовка запроса или генерируем новый
		requestID := r.Header.Get("request_id")
		if requestID == "" {
			requestID = generateRandomString(6)
		}

		// Добавляем `request_id` в контекст запроса
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)

		// Устанавливаем `request_id` в заголовок ответа, чтобы передавать его между сервисами
		w.Header().Set("request_id", requestID)

		// Передаем запрос следующему обработчику с новым контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware для журналирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Запоминаем время начала обработки запроса
		start := time.Now()

		// Извлекаем request_id из контекста
		requestID, _ := r.Context().Value(RequestIDKey{}).(string)

		// Получаем IP-адрес клиента
		clientIP := r.RemoteAddr

		// Перехватываем стандартный http.ResponseWriter для записи кода ответа
		ww := &responseWriter{ResponseWriter: w}

		// Вызываем обработчик запроса
		next.ServeHTTP(ww, r)

		// Логируем информацию о запросе после его обработки
		logger.Printf("Request ID: %s, Method: %s, URL: %s, IP: %s, Status: %d, Duration: %v, Time: %s",
			requestID, r.Method, r.URL.Path, clientIP, ww.statusCode, time.Since(start), start.Format(time.RFC3339))
	})
}

// responseWriter - обертка для ResponseWriter для захвата кода статуса
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает код статуса HTTP-ответа
func (rw *responseWriter) WriteHeader(statusCode int) {
	// Если статусный код еще не установлен, записываем его
	if rw.statusCode == 0 {
		rw.statusCode = statusCode
	}
	// Вызываем WriteHeader у оригинального ResponseWriter
	rw.ResponseWriter.WriteHeader(statusCode)
}

// В случае, если Write не вызывается в вашем обработчике, можно использовать
// Write для корректной записи статуса.
func (rw *responseWriter) Write(b []byte) (int, error) {
	// Если код статуса еще не был установлен, можно установить его в 200
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	// Вызываем Write у оригинального ResponseWriter
	return rw.ResponseWriter.Write(b)
}

// Генерация случайного request_id
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

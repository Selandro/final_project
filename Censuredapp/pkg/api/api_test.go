// // API приложения GoNews.

package api

// import (
// 	"GoNews/pkg/storage"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// // MockDB представляет собой моки для базы данных.
// type MockDB struct{}

// // News возвращает фиксированный набор новостей для тестирования.
// func (mdb *MockDB) News(n int) ([]storage.Post, error) {
// 	if n <= 0 {
// 		return []storage.Post{}, nil // Возвращаем пустой срез для невалидного n
// 	}
// 	return []storage.Post{
// 		{Title: "Test News 1", Content: "Content for test news 1"},
// 		{Title: "Test News 2", Content: "Content for test news 2"},
// 	}, nil
// }

// // StoreNews реализует метод интерфейса, но не делает ничего для моков.
// func (mdb *MockDB) StoreNews(news []storage.Post) error {
// 	// Можно ничего не делать или возвращать nil, так как мы тестируем только чтение
// 	return nil
// }
// func TestAPI_posts(t *testing.T) {
// 	// Инициализируем API с моками базы данных
// 	mockDB := &MockDB{}
// 	api := New(mockDB)

// 	tests := []struct {
// 		name              string
// 		url               string
// 		expectedStatus    int
// 		expectedNewsCount int
// 	}{
// 		{
// 			name:              "Valid request with n=2",
// 			url:               "/news/2",
// 			expectedStatus:    http.StatusOK,
// 			expectedNewsCount: 2,
// 		},
// 		{
// 			name:              "Valid request with n=0",
// 			url:               "/news/0",
// 			expectedStatus:    http.StatusOK,
// 			expectedNewsCount: 0, // Ожидаем пустой массив
// 		},
// 		{
// 			name:              "Invalid request with n=-1",
// 			url:               "/news/-1",
// 			expectedStatus:    http.StatusOK,
// 			expectedNewsCount: 0, // Ожидаем пустой массив
// 		},
// 		{
// 			name:              "Invalid request with n=not-a-number",
// 			url:               "/news/not-a-number",
// 			expectedStatus:    http.StatusBadRequest,
// 			expectedNewsCount: 0,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Создаем новый HTTP-запрос
// 			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
// 			rec := httptest.NewRecorder()

// 			// Вызываем обработчик
// 			api.Router().ServeHTTP(rec, req)

// 			// Проверяем статус-код ответа
// 			if status := rec.Code; status != tt.expectedStatus {
// 				t.Errorf("Неверный статус-код: ожидается %v, получен %v", tt.expectedStatus, status)
// 			}

// 			// Проверяем содержимое ответа, только если ожидается статус 200
// 			if tt.expectedStatus == http.StatusOK {
// 				var news []storage.Post
// 				if err := json.Unmarshal(rec.Body.Bytes(), &news); err != nil {
// 					t.Fatalf("Ошибка при распаковке JSON: %v", err)
// 				}
// 				if len(news) != tt.expectedNewsCount {
// 					t.Errorf("Ожидалось %v новостей, получено %v", tt.expectedNewsCount, len(news))
// 				}
// 			}
// 		})
// 	}
// }

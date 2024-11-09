package models

import "time"

type NewsShortDetailed struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	Content string // содержание публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
}

// Полная детализированная новость с комментариями
type NewsFullDetailed struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Author   string    `json:"author"`
	Comments []Comment `json:"comments"` // Добавляем поле для комментариев
}

type Comment struct {
	ID        int64     `json:"id"`
	NewsID    int64     `json:"news_id"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

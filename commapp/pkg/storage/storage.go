package storage

import (
	"APIGetaway/pkg/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Интерфейс для работы с базой данных
type DBInterface interface {
	AddComment(ctx context.Context, comment models.Comment) (int64, error)
	GetCommentsByNewsID(ctx context.Context, newsID int64) ([]models.Comment, error)
	Close()
}

// Конфигурация БД
type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     int    `json:"port"`
	SSLMode  string `json:"sslmode"`
}

// База данных
type DB struct {
	pool *pgxpool.Pool
}

// Конструктор для инициализации соединения с БД
func New(connstr string) (*DB, error) {
	if connstr == "" {
		return nil, errors.New("не указано подключение к БД")
	}
	pool, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	db := DB{
		pool: pool,
	}

	return &db, nil
}

// Реализация метода для добавления комментария
func (db *DB) AddComment(ctx context.Context, comment models.Comment) (int64, error) {
	var id int64
	query := `INSERT INTO comments (news_id, parent_id, text, created_at)
			  VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.pool.QueryRow(ctx, query, comment.NewsID, comment.ParentID, comment.Text, comment.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления комментария: %w", err)
	}
	return id, nil
}

// Реализация метода для получения комментариев по ID новости
func (db *DB) GetCommentsByNewsID(ctx context.Context, newsID int64) ([]models.Comment, error) {
	query := `SELECT id, news_id, parent_id, text, created_at FROM comments WHERE news_id = $1`
	rows, err := db.pool.Query(ctx, query, newsID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения комментариев: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var parentID *int64
		err := rows.Scan(&comment.ID, &comment.NewsID, &parentID, &comment.Text, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка обработки комментария: %w", err)
		}
		comment.ParentID = parentID
		comments = append(comments, comment)
	}
	return comments, nil
}

// Закрытие соединения с БД
func (db *DB) Close() {
	db.pool.Close()
}

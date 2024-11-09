-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		news_id INTEGER NOT NULL,
		parent_id INTEGER,
		text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd

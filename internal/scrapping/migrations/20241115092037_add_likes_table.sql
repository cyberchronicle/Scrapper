-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scrapping.likes
(
    user_id    INTEGER NOT NULL,
    article_id INTEGER NOT NULL,
    liked_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, article_id),
    FOREIGN KEY (article_id) REFERENCES scrapping.articles (id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS scrapping.articles
(
    id           INTEGER PRIMARY KEY,
    name         TEXT,
    text         TEXT,
    complexity   VARCHAR,
    reading_time INTEGER,
    tags         jsonb
);

comment on table scrapping.articles is 'Статьи';

comment on column scrapping.articles.id is 'Идентификатор статьи';

comment on column scrapping.articles.name is 'Название статьи';

comment on column scrapping.articles.text is 'Текст статьи';

comment on column scrapping.articles.complexity is 'Сложность статьи';

comment on column scrapping.articles.reading_time is 'Время чтения в минутах';

comment on column scrapping.articles.tags is 'Теги статьи';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

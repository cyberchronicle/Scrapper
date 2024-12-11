-- +goose Up
-- +goose StatementBegin
-- Готовим полнотекстовый поиск

-- Создаем таблицу для хранения векторов
CREATE TABLE scrapping.article_vectors
(
    article_id    INTEGER PRIMARY KEY REFERENCES scrapping.articles(id),
    text_tsvector TSVECTOR
);

-- Заполняем таблицу векторами
INSERT INTO scrapping.article_vectors (article_id, text_tsvector)
                SELECT
                    id,
                    to_tsvector('simple', text)
                FROM scrapping.articles;


-- Создаем индекс
CREATE INDEX idx_gin_articles ON scrapping.article_vectors USING gin (text_tsvector);

-- Создаем функцию для триггера
CREATE FUNCTION insert_article_vectors() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO scrapping.article_vectors (article_id, text_tsvector) VALUES
        (new.id, to_tsvector('simple', new.text));

    return new;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер
CREATE TRIGGER insert_article_vectors_trigger
    AFTER INSERT ON scrapping.articles
    FOR EACH ROW EXECUTE FUNCTION insert_article_vectors();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

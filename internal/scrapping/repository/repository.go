package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

var errNotFound = errors.New("obj not found")

func IsNotFoundError(err error) bool {
	return errors.Is(err, errNotFound)
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetLastArticle(ctx context.Context) (int64, error) {
	var id sql.NullInt64
	err := r.db.GetContext(ctx, &id, "SELECT MAX(id) from scrapping.articles")
	if err != nil {
		return 0, fmt.Errorf("error in db: %v", err)
	}
	if !id.Valid {
		return 0, errNotFound
	}
	return id.Int64, nil
}

func (r *Repository) GetFirstArticle(ctx context.Context) (int64, error) {
	var id sql.NullInt64
	err := r.db.GetContext(ctx, &id, "SELECT MIN(id) from scrapping.articles")
	if err != nil {
		return 0, fmt.Errorf("error in db: %v", err)
	}
	if !id.Valid {
		return 0, errNotFound
	}
	return id.Int64, nil
}

func (r *Repository) AddArticle(ctx context.Context, article *Article) error {
	stmt, err := r.db.PrepareNamedContext(ctx, `INSERT INTO scrapping.articles 
    (id, name, text, complexity, reading_time, tags)
	VALUES (:id, :name, :text, :complexity, :reading_time, :tags)
	RETURNING id`)

	if err != nil {
		return fmt.Errorf("add article error: %v", err)
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &article.Id, article)
	if err != nil {
		return fmt.Errorf("add article, error in GetContext: %v", err)
	}
	return nil
}

func (r *Repository) GetArticleById(ctx context.Context, id int) (*Article, error) {
	article := &Article{}
	err := r.db.GetContext(ctx, article, "SELECT * FROM scrapping.articles WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("GetArticleById db error: %v", err)
	}
	return article, nil
}

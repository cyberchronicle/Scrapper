package external

import "context"

type Scrapping interface {
	GetArticles(ctx context.Context, userId int, page, pageSize int) ([]*ArticleInfo, *PaginationInfo, error)
	GetArticleInfoById(ctx context.Context, userId, id int) (*ArticleInfo, error)
	GetArticlesByIds(ctx context.Context, userId int, ids []int) ([]*ArticleInfo, error)
	Like(ctx context.Context, userId, articleId int) error
	Unlike(ctx context.Context, userId, articleId int) error
	Search(ctx context.Context, userId int, query string, pageSize int) ([]*ArticleInfo, error)
}

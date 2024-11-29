package external

import "context"

type Scrapping interface {
	GetArticles(ctx context.Context, userId int, page, pageSize int) ([]*ArticleInfo, *PaginationInfo, error)
	GetArticleInfoById(ctx context.Context, userId, id int) (*ArticleInfo, error)
	Like(ctx context.Context, userId, articleId int) error
	Unlike(ctx context.Context, userId, articleId int) error
}

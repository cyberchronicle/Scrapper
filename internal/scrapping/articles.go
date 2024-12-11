package scrapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"scrapping_service/internal/scrapping/external"
	"scrapping_service/internal/scrapping/repository"
	"scrapping_service/pkg/middlewares"
	"strconv"
)

func (s *Service) GetArticles(ctx context.Context, userId int, page, pageSize int) ([]*external.ArticleInfo, *external.PaginationInfo, error) {
	articlesRepo, paginationInfo, err := s.repo.GetArticlesInfo(ctx, userId, &repository.Cursor{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error in repo: %v", err)
	}

	result := make([]*external.ArticleInfo, 0, len(articlesRepo))
	for _, article := range articlesRepo {
		var tags []string
		err := json.Unmarshal(article.Tags, &tags)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Unmarshal error: %v", err)
		}
		result = append(result, &external.ArticleInfo{
			ID:          article.ID,
			Name:        article.Name,
			Text:        article.Text,
			Complexity:  article.Complexity.String,
			ReadingTime: article.ReadingTime,
			Tags:        tags,
			Likes:       article.LikeCount,
			LikedByUser: article.LikedByUser,
		})
	}

	return result, &external.PaginationInfo{
		Page:            page,
		PageSize:        pageSize,
		HasNextPage:     paginationInfo.HasNextPage,
		HasPreviousPage: paginationInfo.HasPreviousPage,
	}, nil
}

func (s *Service) GetArticleInfoById(ctx context.Context, userId, articleId int) (*external.ArticleInfo, error) {
	articleInfo, err := s.repo.GetArticleInfoById(ctx, userId, articleId)
	if err != nil {
		return nil, err
	}

	var tags []string
	err = json.Unmarshal(articleInfo.Tags, &tags)
	if err != nil {
		return nil, fmt.Errorf("unmarshal tags error: %v", err)
	}

	return &external.ArticleInfo{
		ID:          articleInfo.ID,
		Name:        articleInfo.Name,
		Text:        articleInfo.Text,
		Complexity:  articleInfo.Complexity.String,
		ReadingTime: articleInfo.ReadingTime,
		Tags:        tags,
		Likes:       articleInfo.LikeCount,
		LikedByUser: articleInfo.LikedByUser,
	}, nil
}

func (s *Service) Like(ctx context.Context, userId, articleId int) error {
	err := s.repo.Like(ctx, userId, articleId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Unlike(ctx context.Context, userId, articleId int) error {
	err := s.repo.Unlike(ctx, userId, articleId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Articles(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Header.Get(middlewares.UserId)
	if userIdStr == "" {
		http.Error(w, "User-Id required", http.StatusBadRequest)
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var ids []int

	err = json.NewEncoder(w).Encode(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	articlesInfo, err := s.GetArticlesByIds(r.Context(), userId, ids)
	if err != nil {
		http.Error(w, fmt.Sprintf("GetArticlesByIds error: %v", err), http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(articlesInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal response error: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (s *Service) GetArticlesByIds(ctx context.Context, userId int, ids []int) ([]*external.ArticleInfo, error) {
	repoArticles, err := s.repo.GetArticlesByIds(ctx, userId, ids)
	if err != nil {
		return nil, fmt.Errorf("error in repo: %v", err)
	}
	articlesInfo := make([]*external.ArticleInfo, 0, len(repoArticles))
	for _, article := range repoArticles {
		var tags []string
		err = json.Unmarshal(article.Tags, &tags)
		if err != nil {
			return nil, fmt.Errorf("error unmarshall tags: %v", err)
		}
		articlesInfo = append(articlesInfo, &external.ArticleInfo{
			ID:          article.ID,
			Name:        article.Name,
			Text:        article.Text,
			Complexity:  article.Complexity.String,
			ReadingTime: article.ReadingTime,
			Tags:        tags,
			Likes:       article.LikeCount,
			LikedByUser: article.LikedByUser,
		})
	}

	return articlesInfo, nil
}

func (s *Service) Search(ctx context.Context, userId int, query string, pageSize int) ([]*external.ArticleInfo, error) {
	repoArticles, err := s.repo.Search(ctx, userId, query, pageSize)
	if err != nil {
		return nil, fmt.Errorf("error in repo: %v", err)
	}

	articlesInfo := make([]*external.ArticleInfo, 0, len(repoArticles))
	for _, article := range repoArticles {
		var tags []string
		err = json.Unmarshal(article.Tags, &tags)
		if err != nil {
			return nil, fmt.Errorf("error unmarshall tags: %v", err)
		}
		articlesInfo = append(articlesInfo, &external.ArticleInfo{
			ID:          article.ID,
			Name:        article.Name,
			Text:        article.Text,
			Complexity:  article.Complexity.String,
			ReadingTime: article.ReadingTime,
			Tags:        tags,
			Likes:       article.LikeCount,
			LikedByUser: article.LikedByUser,
		})
	}

	return articlesInfo, nil

}

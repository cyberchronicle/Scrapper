package external

type ArticleInfo struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Text        string   `json:"text"`
	Complexity  string   `json:"complexity"`
	ReadingTime int      `json:"readingTime"`
	Tags        []string `json:"tags"`
	Likes       int      `json:"likes"`
	LikedByUser bool     `json:"likedByUser"`
}

type PaginationInfo struct {
	Page            int  `json:"page"`
	PageSize        int  `json:"pageSize"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

type Article struct {
	Id          int64    `json:"id"`
	Name        string   `json:"name"`
	Text        string   `json:"text"`
	Complexity  string   `json:"complexity"`
	ReadingTime int64    `json:"reading_time"`
	Tags        []string `json:"tags"`
}

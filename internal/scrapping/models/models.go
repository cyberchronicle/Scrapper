package models

type Article struct {
	Id          int64    `json:"id"`
	Name        string   `json:"name"`
	Text        string   `json:"text"`
	Complexity  string   `json:"complexity"`
	ReadingTime int64    `json:"reading_time"`
	Tags        []string `json:"tags"`
}

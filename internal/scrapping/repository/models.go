package repository

import "encoding/json"

type Article struct {
	Id          int64           `db:"id"`
	Name        string          `db:"name"`
	Text        string          `db:"text"`
	Complexity  string          `db:"complexity"`
	ReadingTime int64           `db:"reading_time"`
	Tags        json.RawMessage `db:"tags"`
}

package repository

import (
	"database/sql"
	"encoding/json"
)

type Article struct {
	Id          int64           `db:"id"`
	Name        string          `db:"name"`
	Text        string          `db:"text"`
	Complexity  sql.NullString  `db:"complexity"`
	ReadingTime int64           `db:"reading_time"`
	Tags        json.RawMessage `db:"tags"`
}

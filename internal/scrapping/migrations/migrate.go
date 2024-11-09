package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

//go:embed *.sql
var embedMigrations embed.FS

func MigrateUp(db *sql.DB) {
	log.Info().Str("module", "database").Msgf("scrapping migrate begin")

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("scrapping migrate db error in goose.SetDialect: %v", err))
	}

	goose.SetTableName("scrapping.goose_db_version")

	if err := goose.Up(db, "."); err != nil {
		panic(fmt.Errorf("scrapping migrate db error in goose.Up: %v", err))
	}
	log.Info().Str("module", "database").Msgf("scrapping migrate end")
}

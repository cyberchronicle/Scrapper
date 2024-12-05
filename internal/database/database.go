package database

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"scrapping_service/pkg/utils"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Conf struct {
	Dialect         string `yaml:"dialect"`
	Dsn             string `yaml:"dsn"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	ConnMaxLifeTime int    `yaml:"connMaxLifeTime"`
}

type Database struct {
	utils.Conv

	DBX   *sqlx.DB
	ctx   context.Context
	xConf sync.RWMutex
	conf  *Conf
}

func (d *Database) setConf(conf *Conf) {
	d.xConf.Lock()
	d.conf = conf
	d.xConf.Unlock()
}

func (d *Database) getConf() *Conf {
	d.xConf.RLock()
	defer d.xConf.RUnlock()
	return d.conf
}

func NewDatabase(ctx context.Context, name, namespace string) *Database {
	return &Database{ctx: ctx, Conv: utils.NewConv(name, namespace)}
}

func (d *Database) Configure(conf *Conf) {
	log.Info().Str("module", d.Name).Msgf("database %s conf: configure begin", d.Name)

	d.setConf(conf)

	d.Load.Do(func() {
		log.Info().Str("module", d.Name).Msg("database connecting...")

		db, err := sqlx.Connect(conf.Dialect, conf.Dsn)
		if err != nil {
			err = fmt.Errorf("error in sqlx.Connect: %v", err)
			log.Panic().Str("module", d.Name).Err(err)
			panic(err)
		}
		if err = db.Ping(); err != nil {
			err = fmt.Errorf("error in db.Ping: %v", err)
			log.Panic().Str("module", d.Name).Msgf("%v", err)
			panic(err)
		}

		db.SetMaxOpenConns(conf.MaxOpenConns)
		db.SetMaxIdleConns(conf.MaxIdleConns)
		db.SetConnMaxLifetime(time.Second * time.Duration(conf.ConnMaxLifeTime))

		d.DBX = db

		d.RunWorker(d.ping, "ping", 1)
	})
}

func (d *Database) ping() {
	defer func() {
		_ = d.DBX.Close()
	}()
	for {
		if err := d.DBX.Ping(); err != nil {
			log.Error().Str("module", d.Name).Msgf("database error in ping: %v", err)
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 30):
		}
	}
}

func (d *Database) WaitTerminate() {
	log.Info().Str("module", d.Name).Msg("database term: begin")

	_ = d.DBX.Close()

	log.Info().Str("module", d.Name).Msg("database term: end")
}

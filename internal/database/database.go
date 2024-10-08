package database

import (
	"context"
	"fmt"
	"scrapping_service/pkg/utils"
	"sync"
	"time"

	zlog "github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Conf struct {
	Dialect string `yaml:"dialect"`
	Dsn     string `yaml:"dsn"`
}

type Database struct {
	utils.Conv

	DB    *mongo.Client
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
	zlog.Info().Msgf("database %s conf: configure begin", d.Name)

	d.setConf(conf)

	d.Load.Do(func() {
		zlog.Info().Msg("database connecting...")

		db, err := mongo.Connect(d.ctx, options.Client().ApplyURI(d.getConf().Dsn))
		if err != nil {
			err = fmt.Errorf("error in mongo.Connect: %v", err)
			zlog.Panic().Err(err)
			panic(err)
		}
		d.DB = db

		d.RunWorker(d.ping, "ping", 1)
	})
}

func (d *Database) ping() {
	defer func() {
		_ = d.DB.Disconnect(d.ctx)
	}()
	for {
		if err := d.DB.Ping(d.ctx, readpref.Primary()); err != nil {
			zlog.Error().Msgf("database error in ping: %v", err)
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 30):
		}
	}
}

func (d *Database) WaitTerminate() {
	zlog.Info().Msg("database term: begin")

	_ = d.DB.Disconnect(d.ctx)

	zlog.Info().Msg("database term: end")
}

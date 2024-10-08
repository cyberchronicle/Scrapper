package main

import (
	"scrapping_service/internal/database"

	zlog "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"os"
	"scrapping_service/internal/scraping_server"
	"scrapping_service/pkg/signal"
	"sync"
)

var (
	configPath     = "config/config.yaml"
	initMain       sync.Once
	scrapingServer *scraping_server.Service
)

func main() {
	zlog.Info().Msg("service starting...")

	err := Configure()
	if err != nil {
		zlog.Err(err)
		return
	}

	go WaitTerminate()

	signal.Wait()
}

type Conf struct {
	scraping *scraping_server.Conf
	database *database.Conf
}

func Configure() error {

	file, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var conf Conf
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return err
	}

	initMain.Do(func() {
		scrapingServer = scraping_server.NewServer(signal.Context, "scrapping_server", "scrapper")
	})

	scrapingServer.Configure(conf.scraping, conf.database)
	return nil
}

func WaitTerminate() {
	signal.WaitGroup.Add(1)
	defer signal.WaitGroup.Done()

	<-signal.Context.Done()

	zlog.Info().Msg("term: begin")

	scrapingServer.WaitTerminate()

	zlog.Info().Msg("term: end")
}

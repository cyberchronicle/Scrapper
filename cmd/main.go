package main

import (
	"scrapping_service/internal/database"
	"scrapping_service/internal/scrapping"

	log "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"os"
	"scrapping_service/pkg/signal"
	"sync"
)

var (
	configPath      = "config/config.yaml"
	initMain        sync.Once
	scrapingService *scrapping.Service
)

func main() {
	log.Info().Msg("service starting...")

	err := Configure()
	if err != nil {
		log.Err(err)
		return
	}

	go WaitTerminate()

	signal.Wait()
}

type Conf struct {
	Scraping *scrapping.Conf `yaml:"scraping"`
	Database *database.Conf  `yaml:"database"`
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
		scrapingService = scrapping.NewService(signal.Context, "scrapping_server", "scrapper")
	})

	scrapingService.Configure(conf.Scraping, conf.Database)
	return nil
}

func WaitTerminate() {
	signal.WaitGroup.Add(1)
	defer signal.WaitGroup.Done()

	<-signal.Context.Done()

	log.Info().Msg("term: begin")

	scrapingService.WaitTerminate()

	log.Info().Msg("term: end")
}

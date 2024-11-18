package scraping_server

import (
	"context"
	"errors"
	"net/http"
	"scrapping_service/internal/database"
	"scrapping_service/internal/scraping_server/repository"
	"scrapping_service/pkg/middlewares"
	"scrapping_service/pkg/utils"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	zlog "github.com/rs/zerolog/log"
)

type Conf struct {
	Host string `yaml:"host"`
}

type Service struct {
	utils.Conv

	ctx   context.Context
	xConf sync.RWMutex
	conf  *Conf

	server *http.Server

	repo *repository.Repository

	// поля для скрапинга
}

func NewServer(ctx context.Context, name, namespace string) *Service {
	return &Service{ctx: ctx, Conv: utils.NewConv(name, namespace)}
}

func (s *Service) Configure(conf *Conf, confDb *database.Conf) {
	zlog.Info().Msg("conf: configure begin")

	s.setConf(conf)

	s.Load.Do(func() {
		// подключаемся к БД
		db := database.NewDatabase(s.ctx, "database", "scrapping")
		db.Configure(confDb)

		s.repo = repository.NewRepository(db.DB)

		s.RunWorker(s.start, "start", 1)
	})

	zlog.Info().Msg("conf: configure end")
}

func (s *Service) start() {
	defer zlog.Info().Msg("start worker closed")

	r := chi.NewRouter()

	r.Use(middlewares.Logger(s.Name))

	r.Get("some", func(w http.ResponseWriter, r *http.Request) {
	})

	s.server = &http.Server{
		Addr:    s.getConf().Host,
		Handler: r,
	}

	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		zlog.Error().Msgf("scrapper server http error %v", err)
	} else {
		zlog.Error().Msgf("scrapper server http shudown %v", err)
	}
}

func (s *Service) setConf(conf *Conf) {
	s.xConf.Lock()
	s.conf = conf
	s.xConf.Unlock()
}

func (s *Service) getConf() *Conf {
	s.xConf.RLock()
	defer s.xConf.RUnlock()
	return s.conf
}

func (s *Service) WaitTerminate() {
	zlog.Info().Msg("scrapper server term: begin")

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)

	defer func() {
		cancel()
	}()

	if err := s.server.Shutdown(ctx); err != nil {
		zlog.Err(err)
	}

	s.WaitWorker("start")

	zlog.Info().Msg("scrapper server term: begin")
}

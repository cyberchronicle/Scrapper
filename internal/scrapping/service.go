package scrapping

import (
	"context"
	"errors"
	"net/http"
	"scrapping_service/internal/database"
	"scrapping_service/internal/scrapping/repository"
	"scrapping_service/pkg/middlewares"
	"scrapping_service/pkg/utils"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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

func NewService(ctx context.Context, name, namespace string) *Service {
	return &Service{ctx: ctx, Conv: utils.NewConv(name, namespace)}
}

func (s *Service) Configure(conf *Conf, confDb *database.Conf) {
	log.Info().Str("module", s.Name).Msg("conf: configure begin")

	s.setConf(conf)

	s.Load.Do(func() {
		// подключаемся к БД
		db := database.NewDatabase(s.ctx, "database", "scrapping")
		db.Configure(confDb)

		s.repo = repository.NewRepository(db.DBX)

		s.RunWorker(s.start, "start", 1)
	})

	log.Info().Str("module", s.Name).Msg("conf: configure end")
}

func (s *Service) start() {
	defer log.Info().Str("module", s.Name).Msg("start worker closed")

	r := chi.NewRouter()

	r.Use(middlewares.Logger(s.Name))

	r.Get("/api/v1/scrapper/health", checkHealth)

	for {

		conf := s.getConf()
		log.Info().Str("module", s.Name).Msgf("scrapper http server starting on %s", conf.Host)

		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// запускаем http сервер
		s.server = &http.Server{
			Addr:    s.getConf().Host,
			Handler: r,
		}

		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error().Str("module", s.Name).Msgf("scrapper server http error %v", err)
		} else {
			log.Error().Str("module", s.Name).Msgf("scrapper server http shudown %v", err)
			return
		}

		time.Sleep(time.Second)
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
	log.Info().Str("module", s.Name).Msg("scrapper server term: begin")

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)

	defer func() {
		cancel()
	}()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Err(err).Str("module", s.Name)
	}

	s.WaitWorker("start")

	log.Info().Str("module", s.Name).Msg("scrapper server term: begin")
}

func checkHealth(w http.ResponseWriter, r *http.Request) {
	// Простая проверка здоровья, отвечаем статусом 200
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

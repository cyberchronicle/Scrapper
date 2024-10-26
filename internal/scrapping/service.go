package scrapping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-co-op/gocron"
	"net/http"
	"scrapping_service/internal/database"
	"scrapping_service/internal/scrapping/migrations"
	"scrapping_service/internal/scrapping/models"
	"scrapping_service/internal/scrapping/repository"
	"scrapping_service/pkg/middlewares"
	"scrapping_service/pkg/utils"
	"strconv"
	"strings"
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
	client *http.Client

	repo *repository.Repository

	cron *gocron.Scheduler

	// поля для скрапинга
	converter *md.Converter
}

func NewService(ctx context.Context, name, namespace string) *Service {
	return &Service{
		ctx:       ctx,
		Conv:      utils.NewConv(name, namespace),
		cron:      gocron.NewScheduler(time.UTC),
		converter: md.NewConverter("", true, nil),
	}
}

func (s *Service) Configure(conf *Conf, confDb *database.Conf) {
	log.Info().Str("module", s.Name).Msg("conf: configure begin")

	s.setConf(conf)

	s.Load.Do(func() {
		// подключаемся к БД
		db := database.NewDatabase(s.ctx, "database", "scrapping")
		db.Configure(confDb)

		// запускаем миграции
		migrations.MigrateUp(db.DBX.DB)

		s.repo = repository.NewRepository(db.DBX)

		s.client = &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.Contains(req.URL.String(), "articles") {
				return nil
			}
			return errors.New("redirected")
		}}

		s.RunWorker(s.start, "start", 1)

		_, err := s.cron.Every(time.Minute * 30).Do(s.scrap)
		if err != nil {
			log.Error().Str("module", s.Name).Msgf("start cron for scrapping: %v", err)
		}

		s.cron.StartAsync()
	})

	log.Info().Str("module", s.Name).Msg("conf: configure end")
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

func (s *Service) scrap() {
	lastArticleSite, err := getLastArticle()
	if err != nil {
		log.Error().Str("module", s.Name).Msgf("GetLastArticle from site error: %v", err)
		return
	}

	lastArticle, err := s.repo.GetLastArticle(s.ctx)
	if err != nil {
		if repository.IsNotFoundError(err) {
			lastArticle = lastArticleSite - 30
		} else {
			log.Error().Str("module", s.Name).Msgf("GetLastArticle from repo error: %v", err)
			return
		}
	}
	if lastArticleSite < lastArticle {
		log.Warn().Str("module", s.Name).Msgf("strange article ids: in repo %v, in site %v", lastArticle, lastArticleSite)
		return
	}

	for i := lastArticle + 1; i <= lastArticleSite; i++ {
		// выходим из цикла
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		article, err := s.getArticle(i)
		if err != nil {
			log.Error().Str("module", s.Name).Msgf("getArticle error: %v", err)
			continue
		}
		repoArticle, err := mapArticle(article)
		if err != nil {
			log.Error().Str("module", s.Name).Msgf("mapArticle error: %v", err)
			continue
		}
		err = s.repo.AddArticle(s.ctx, repoArticle)
		if err != nil {
			log.Error().Str("module", s.Name).Msgf("error in repo: %v", err)
			continue
		}
	}
}

func getLastArticle() (int64, error) {
	doc, err := goquery.NewDocument("https://habr.com/ru/articles/")
	if err != nil {
		return 0, err
	}

	var articleID int64
	var parseErr error
	found := false

	doc.Find("article[class=tm-articles-list__item]").First().Each(func(i int, item *goquery.Selection) {
		val, ok := item.Attr("id")
		if !ok {
			parseErr = errors.New("attribute 'id' not found")
			return
		}

		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			parseErr = err
			return
		}

		articleID = id
		found = true
	})

	if parseErr != nil {
		return 0, parseErr
	}

	if !found {
		return 0, errors.New("no article found")
	}

	return articleID, nil
}

func (s *Service) getArticle(id int64) (*models.Article, error) {
	url := fmt.Sprintf("https://habr.com/ru/articles/%v/", id)
	get, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	if get.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("not received 200 status code: %v", get.StatusCode)
	}
	defer get.Body.Close()
	doc, err := goquery.NewDocumentFromReader(get.Body)
	if err != nil {
		return nil, err
	}
	var (
		name, text, complexity string
		readingTime            int64
		tags                   []string
	)
	doc.Find("div[xmlns='http://www.w3.org/1999/xhtml']").Each(func(i int, item *goquery.Selection) {
		text = s.converter.Convert(item)
	})

	doc.Find("h1[class='tm-title tm-title_h1']").Each(func(i int, item *goquery.Selection) {
		name = item.Text()
	})

	doc.Find("span[class=tm-article-complexity__label]").Each(func(i int, item *goquery.Selection) {
		complexity = item.Text()
	})

	doc.Find("span[class=tm-article-reading-time__label]").Each(func(i int, item *goquery.Selection) {
		arr := strings.Split(item.Text(), " ")
		readingTime, _ = strconv.ParseInt(arr[0], 10, 64)
	})

	doc.Find("a[class=tm-tags-list__link]").Each(func(i int, item *goquery.Selection) {
		tag := item.Text()
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	return &models.Article{
		Id:          id,
		Name:        name,
		Text:        text,
		Complexity:  complexity,
		ReadingTime: readingTime,
		Tags:        tags,
	}, nil
}

func mapArticle(article *models.Article) (*repository.Article, error) {
	tags, err := json.Marshal(article.Tags)
	if err != nil {
		return nil, err
	}
	return &repository.Article{
		Id:          article.Id,
		Name:        article.Name,
		Text:        article.Text,
		Complexity:  article.Complexity,
		ReadingTime: article.ReadingTime,
		Tags:        tags,
	}, nil
}

func (s *Service) WaitTerminate() {
	log.Info().Str("module", s.Name).Msg("scrapper server term: begin")

	s.cron.Stop()

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

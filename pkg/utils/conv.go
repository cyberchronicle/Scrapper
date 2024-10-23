package utils

import (
	"scrapping_service/pkg/signal"
	"sync"

	log "github.com/rs/zerolog/log"
)

// ---------------
// ConvGraceful - интерфейс компонентов для грейсфул завершения

type ConvGraceful interface {
	WaitTerminate()
}

// ---------------
// Conv - интерфейс для компонентов приложения, обрабатывающие события

type Conv struct {
	ConvGraceful

	Name      string
	Namespace string
	Load      sync.Once
	term      map[string]chan bool
	xTerm     sync.RWMutex
}

func NewConv(name string, namespace string) Conv {
	return Conv{
		Name:      name,
		Namespace: namespace,
		term:      make(map[string]chan bool),
	}
}

func (c *Conv) RunWorker(worker func(), name string, count int) {
	workersWaitGroup := &sync.WaitGroup{}
	c.xTerm.Lock()
	c.term[name] = make(chan bool, 1)
	c.xTerm.Unlock()

	for i := 0; i < count; i++ {
		// регистрируем в группе воркеров
		workersWaitGroup.Add(1)
		// запускаем обработку сообщений
		go func() {
			// регистрируем себя в Graceful Terminating модуле
			signal.WaitGroup.Add(1)
			defer signal.WaitGroup.Done()

			defer workersWaitGroup.Done()
			worker()
		}()
	}

	go func() {
		workersWaitGroup.Wait()
		c.xTerm.RLock()
		defer c.xTerm.RUnlock()
		c.term[name] <- true
	}()
}

func (c *Conv) WaitWorker(name string) {
	c.xTerm.RLock()
	m := c.term[name]
	c.xTerm.RUnlock()

	<-m
}

func (c *Conv) WaitTerminate() {
	log.Error().Msgf("%s: WaitTerminate not implemented!!!", c.Name)
}

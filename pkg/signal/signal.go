package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/rs/zerolog/log"
)

var (
	Context           context.Context
	WaitGroup         sync.WaitGroup
	gracefulTerminate chan os.Signal
	cancelFunc        context.CancelFunc
)

// Init Graceful Terminating
func init() {
	// создаём контекст для выполнения задач
	Context, cancelFunc = context.WithCancel(context.Background())

	gracefulTerminate = make(chan os.Signal)
	// подписываемся на сигналы для остановки программы
	signal.Notify(gracefulTerminate, syscall.SIGTERM)
	signal.Notify(gracefulTerminate, syscall.SIGINT)
}

// Ожидает выполнения Graceful Terminating. Используется в главном потоке
func Wait() {
	sig := <-gracefulTerminate

	log.Info().Msg(sig.String())

	// сообщаем всем воркерам что необходимо остановить работу
	cancelFunc()

	// ожидаем остановки воркеров n секунд
	if WaitTimeout(&WaitGroup, time.Second*60) {
		log.Error().Msg("signal: completion waiting timed out")

	} else {
		log.Info().Msg("signal: all services successfully completed")
	}

	// завершаем программу
	log.Info().Msg("signal: terminated")

	os.Exit(0)
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // корректное завершение
	case <-time.After(timeout):
		return true // превышен тайамут ожидания
	}
}

package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
	"scrapping_service/pkg/utils"
	"sync"
	"time"
)

type Kafka interface {
	SendAsyncMessage(message json.RawMessage)
}

type Conf struct {
	Topic       string   `yaml:"topic"`
	Brokers     []string `yaml:"brokers"`
	DialTimeout int      `yaml:"dialTimeout"`
}

type Service struct {
	utils.Conv

	ctx      context.Context
	producer sarama.AsyncProducer
	topic    string

	xConf sync.RWMutex
	conf  *Conf
}

func NewService(ctx context.Context, name, namespace string) *Service {
	return &Service{
		ctx:  ctx,
		Conv: utils.NewConv(name, namespace),
	}
}

func (s *Service) Configure(conf *Conf) {
	log.Info().Str("module", s.Name).Msg("conf: configure begin")

	s.setConf(conf)

	s.Load.Do(func() {
		log.Info().Str("module", s.Name).Msg("kafka starting")

		s.topic = conf.Topic

		producerConfig := sarama.NewConfig()
		producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner
		producerConfig.Producer.RequiredAcks = sarama.WaitForAll

		producerConfig.Producer.Return.Errors = true
		producerConfig.Producer.Return.Successes = true
		producerConfig.Net.DialTimeout = time.Duration(conf.DialTimeout) * time.Second

		producer, err := sarama.NewAsyncProducer(conf.Brokers, producerConfig)
		if err != nil {
			panic(err)
		}
		s.producer = producer

		s.RunWorker(s.kafkaErrors, "kafka_errors", 1)
		s.RunWorker(s.kafkaSuccess, "kafka_success", 1)

	})
	log.Info().Str("module", s.Name).Msg("conf: configure end")
}

func (s *Service) SendAsyncMessage(message json.RawMessage) {
	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(message),
	}
	s.producer.Input() <- msg
}

func (s *Service) kafkaErrors() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.producer.Errors():
			log.Error().Str("module", s.Name).Msgf("kafka error: %v", msg.Error())
		}
	}
}

func (s *Service) kafkaSuccess() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.producer.Successes():
			value, _ := msg.Value.Encode()
			log.Info().Str("module", s.Name).Msgf("kafka success: %v", string(value))
		}
	}
}

func (s *Service) getConf() *Conf {
	s.xConf.RLock()
	defer s.xConf.RUnlock()
	return s.conf
}

func (s *Service) setConf(conf *Conf) {
	s.xConf.Lock()
	defer s.xConf.Unlock()
	s.conf = conf
}

func (s *Service) WaitTerminate() {
	log.Info().Str("module", s.Name).Msg("term: begin")

	s.WaitWorker("kafka_errors")
	s.WaitWorker("kafka_success")

	log.Info().Str("module", s.Name).Msg("term: end ")
}

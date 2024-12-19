package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	events2 "basic-go/webook/interact/events"
	"basic-go/webook/internal/events"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

func NewConsumers(c0 *events2.InteractReadEventBatchConsumer) []events.Consumer {
	return []events.Consumer{c0}
}

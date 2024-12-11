package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"basic-go/webook/internal/events"
	"basic-go/webook/internal/events/article"
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

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
// func NewConsumers(c0 *article.InteractReadEventConsumer) []events.Consumer {
// 	return []events.Consumer{c0}
// }

func NewConsumers(c0 *article.InteractReadEventBatchConsumer) []events.Consumer {
	return []events.Consumer{c0}
}

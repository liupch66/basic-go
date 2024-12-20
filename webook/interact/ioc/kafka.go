package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"github.com/liupch66/basic-go/webook/interact/events"
	"github.com/liupch66/basic-go/webook/pkg/saramax"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("kafka", &cfg); err != nil {
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

func NewConsumers(c0 *events.InteractReadEventConsumer) []saramax.Consumer {
	return []saramax.Consumer{c0}
}

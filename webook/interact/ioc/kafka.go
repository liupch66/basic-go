package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"github.com/liupch66/basic-go/webook/interact/events"
	"github.com/liupch66/basic-go/webook/interact/repository/dao"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events/fixer"
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

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c0 *events.InteractReadEventConsumer,
	c1 *fixer.Consumer[dao.Interact]) []saramax.Consumer {
	return []saramax.Consumer{c0, c1}
}

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

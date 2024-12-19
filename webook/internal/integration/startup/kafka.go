package startup

import (
	"github.com/IBM/sarama"

	"basic-go/webook/internal/events"
	"basic-go/webook/internal/events/article"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9092"}, saramaCfg)
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

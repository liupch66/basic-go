package article

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Uid int64
	Aid int64
}

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}

// ProduceReadEvent 可以考虑重试，简单直接实现，复杂装饰器
func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicReadEvent,
		Value: sarama.ByteEncoder(data),
	})
	return err
}

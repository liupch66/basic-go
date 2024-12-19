package saramax

import (
	"encoding/json"

	"github.com/IBM/sarama"

	"github.com/liupch66/basic-go/webook/pkg/logger"
)

type Handler[T any] struct {
	l  logger.LoggerV1
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger.LoggerV1, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{l: l, fn: fn}
}

func (h *Handler[T]) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(sess sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 可以考虑在这个封装里面提供统一的重试机制，复杂的重试可以用装饰器实现
func (h *Handler[T]) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化消息失败",
				logger.String("topic: ", claim.Topic()),
				logger.Int32("partition: ", msg.Partition),
				logger.Int64("offset: ", msg.Offset),
				logger.Error(err),
			)
			// 不中断，继续消费下一个消息
			continue
		}
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			if err == nil {
				break
			}
			h.l.Error("处理消息失败",
				logger.String("topic: ", claim.Topic()),
				logger.Int32("partition: ", msg.Partition),
				logger.Int64("offset: ", msg.Offset),
				logger.Error(err),
			)
		}
		if err != nil {
			h.l.Error("处理消息失败-重试上限",
				logger.String("topic: ", claim.Topic()),
				logger.Int32("partition: ", msg.Partition),
				logger.Int64("offset: ", msg.Offset),
				logger.Error(err),
			)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"github.com/liupch66/basic-go/webook/pkg/logger"
)

type BatchHandler[T any] struct {
	l  logger.LoggerV1
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 用 option 模式来设置这个 batchSize 和 batchDuration
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(msg []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{l: l, fn: fn, batchSize: 10, batchDuration: time.Second}
}

func (b *BatchHandler[T]) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batchSize := b.batchSize
	for {
		var (
			done bool
			msgs = make([]*sarama.ConsumerMessage, 0, batchSize)
			ts   = make([]T, 0, batchSize)
		)
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-claim.Messages():
				if !ok {
					cancel()
					return nil
				}
				var t T
				if err := json.Unmarshal(msg.Value, &t); err != nil {
					b.l.Error("反序列化消息失败",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err),
					)
					// 不中断，继续下一个
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error("调用业务批量消费消息失败", logger.Error(err))
		} else {
			for _, msg := range msgs {
				sess.MarkMessage(msg, "")
			}
		}
		cancel()
	}
}

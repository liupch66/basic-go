package fixer

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"gorm.io/gorm"

	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/migrator"
	"github.com/liupch66/basic-go/webook/pkg/migrator/events"
	"github.com/liupch66/basic-go/webook/pkg/migrator/fixer"
	"github.com/liupch66/basic-go/webook/pkg/saramax"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](client sarama.Client, l logger.LoggerV1,
	src, dst *gorm.DB, topic string) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

// Start 这边就是自己启动 goroutine 了
func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{c.topic}, saramax.NewHandler(c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, t.Id)
	case "DST":
		return c.dstFirst.Fix(ctx, t.Id)
	}
	return errors.New("未知的修复方向")
}

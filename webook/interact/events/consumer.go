package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/liupch66/basic-go/webook/interact/repository"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/saramax"
)

type InteractReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractRepository
	l      logger.LoggerV1
}

func NewInteractReadEventConsumer(client sarama.Client, repo repository.InteractRepository, l logger.LoggerV1) *InteractReadEventConsumer {
	return &InteractReadEventConsumer{client: client, repo: repo, l: l}
}

func (i *InteractReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", t.Aid)
}

func (i *InteractReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interact", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{"article_read_event"},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

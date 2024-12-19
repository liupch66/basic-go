package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/liupch66/basic-go/webook/interact/repository"
	"github.com/liupch66/basic-go/webook/pkg/logger"
	"github.com/liupch66/basic-go/webook/pkg/saramax"
)

type InteractReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractRepository
	l      logger.LoggerV1
}

func NewInteractReadEventBatchConsumer(client sarama.Client, repo repository.InteractRepository, l logger.LoggerV1) *InteractReadEventBatchConsumer {
	return &InteractReadEventBatchConsumer{client: client, repo: repo, l: l}
}

func (i *InteractReadEventBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	bizIds := make([]int64, 0, len(ts))
	for _, t := range ts {
		bizIds = append(bizIds, t.Aid)
	}
	return i.repo.BatchIncrReadCnt(ctx, "article", bizIds)
}

func (i *InteractReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interact", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{"article_read_event"}, saramax.NewBatchHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出了批量消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

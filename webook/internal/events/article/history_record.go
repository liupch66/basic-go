package article

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/events"
	"basic-go/webook/internal/repository"
	"basic-go/webook/pkg/logger"
	"basic-go/webook/pkg/saramax"
)

var _ events.Consumer = (*HistoryRecordConsumer)(nil)

type HistoryRecordConsumer struct {
	client sarama.Client
	repo   repository.HistoryRecordRepository
	l      logger.LoggerV1
}

func NewHistoryRecordConsumer(client sarama.Client, repo repository.HistoryRecordRepository, l logger.LoggerV1) *HistoryRecordConsumer {
	return &HistoryRecordConsumer{client: client, repo: repo, l: l}
}

func (i *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.AddRecord(ctx, domain.HistoryRecord{
		Biz:   "article",
		BizId: t.Aid,
		Uid:   t.Uid,
	})
}

func (i *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("history_record", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{topicReadEvent}, saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

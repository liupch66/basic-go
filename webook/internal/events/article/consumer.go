package article

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"basic-go/webook/internal/events"
	"basic-go/webook/internal/repository"
	"basic-go/webook/pkg/logger"
	"basic-go/webook/pkg/saramax"
)

var _ events.Consumer = (*InteractReadEventConsumer)(nil)

type InteractReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractRepository
	l      logger.LoggerV1
}

func NewInteractReadEventConsumer(client sarama.Client, repo repository.InteractRepository, l logger.LoggerV1) *InteractReadEventConsumer {
	return &InteractReadEventConsumer{client: client, repo: repo, l: l}
}

// Consume 这个不是幂等的
/*
幂等（Idempotence）是指一个操作无论执行多少次，产生的效果都是一样的。换句话说，幂等操作在重复执行时不会改变系统的状态。
在编程中，实现幂等性的方法有很多，具体取决于操作的类型和上下文。以下是一些常见的方法：
	1. 使用唯一标识符：确保每个请求都有一个唯一的标识符，服务器可以使用这个标识符来检查请求是否已经处理过。如果处理过，则直接返回之前的结果。
	2. 乐观锁：在更新数据时，使用版本号或时间戳来确保数据的一致性。如果版本号或时间戳不匹配，则拒绝更新。
	3. 幂等操作：设计操作本身就是幂等的。例如，设置某个值而不是增加或减少某个值。
	4. 去重机制：在处理请求时，记录已经处理过的请求，避免重复处理。
*/
func (i *InteractReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", t.Aid)
}

// func (i *InteractReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
//
// 	// 使用唯一标识符检查请求是否已经处理过
// 	processed, err := i.repo.HasProcessed(ctx, msg.Key)
// 	if err != nil {
// 		return err
// 	}
// 	if processed {
// 		return nil
// 	}
//
// 	// 处理请求
// 	err = i.repo.IncrReadCnt(ctx, "article", t.Aid)
// 	if err != nil {
// 		return err
// 	}
//
// 	// 标记请求为已处理
// 	return i.repo.MarkProcessed(ctx, msg.Key)
// }

func (i *InteractReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interact", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{topicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

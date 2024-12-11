package sarama

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

type TestConsumerGroupHandler struct {
}

func (h *TestConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	// sess.Claims() returns information about the claimed partitions by topic.
	partitions := sess.Claims()["test_topic"]
	for _, part := range partitions {
		// 在部分场景下，我们会希望消费历史消息，或者从某个消息开始消费，那么可以考虑在 Setup 里面设置偏移量。
		// 关键调用是 ResetOffset 。不过一般我都是建议走离线渠道，操作 Kafka 集群去重置对应的偏移量。
		// 核心在于，你并不是每次重新部署，重新启动都是要重置这个偏移量的
		sess.ResetOffset("test_topic", part, sarama.OffsetOldest, "")
		// sess.ResetOffset("test_topic", part, sarama.OffsetNewest, "")
		// sess.ResetOffset("test_topic", part, 0, "")
	}
	log.Println("Setup")
	return nil
}

func (h *TestConsumerGroupHandler) ConsumeClaimV1(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Println("consumed message: ", string(msg.Value))
		// 提交消费偏移量
		sess.MarkMessage(msg, "")
	}
	return nil
}

// 异步批量消费
func (h *TestConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	const batchSize = 10
	for {
		var eg errgroup.Group
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		for i := 0; i < batchSize; i++ {
			select {
			case <-ctx.Done():
				// 这一批次超时了，不再尝试凑够一批了
				goto outerLoop
			case msg, ok := <-claim.Messages():
				if !ok {
					// channel 关闭了
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					// 这里消费消息
					time.Sleep(time.Second)
					log.Println("consumed message: ", string(msg.Value))
					return nil
				})
			}
		}
	outerLoop:
		err := eg.Wait()
		if err != nil {
			// 部分失败，可以在这里重试，记录日志，也可以在业务逻辑，也就是 eg.Go 里重试
			log.Println("consume batch error: ", err)
			continue
		} else {
			// 提交消费偏移量
			for _, msg := range batch {
				sess.MarkMessage(msg, "")
			}
		}
		cancel()
	}
}

func (h *TestConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addrs, "test_group", cfg)
	assert.NoError(t, err)
	// 通过 ConsumerGroupHandler 启动一个阻塞的 ConsumerGroupSession
	// Setup --> ConsumeClaim --> Cleanup
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	err = consumer.Consume(ctx, []string{"test_topic"}, &TestConsumerGroupHandler{})
	assert.NoError(t, err)
	t.Log(time.Since(start).String())
}

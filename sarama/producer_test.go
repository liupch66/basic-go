package sarama

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var addrs = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 第一种创建方法
	// client, err := sarama.NewClient(addrs, cfg)
	// assert.NoError(t, err)
	// producer, err  := sarama.NewSyncProducerFromClient(client)
	// assert.NoError(t, err)

	// 第二种创建方法
	cfg.Producer.Return.Successes = true
	// 0：客户端发一次，不需要服务端的确认。
	cfg.Producer.RequiredAcks = sarama.NoResponse
	// 1：客户端发送，并且需要服务端写入到主分区。
	// cfg.Producer.RequiredAcks = sarama.WaitForLocal
	// -1：客户端发送，并且需要服务端同步到所有的 ISR（In-Sync Replicas，同步副本） 上。
	// cfg.Producer.RequiredAcks = sarama.WaitForAll
	// // Hash：根据 key 的哈希值来筛选一个，默认值
	// cfg.Producer.Partitioner = sarama.NewHashPartitioner
	// // Random：随机
	// cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	// // RoundRobin：轮询
	// cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	// // ConsistentCRC：一致性哈希，用的是 CRC16 算法，通常用于确保相同的 Key 总是被发送到同一个分区
	// cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	// // Manual：根据 Message 中的 partition 字段来选择
	// cfg.Producer.Partitioner = sarama.NewManualPartitioner
	// // Custom：实际上不 Custom，而是自定义一部分 Hash 的参数，本质上是一个 Hash 的实现
	// // 你可以通过实现自己的哈希算法或分区逻辑来控制消息发送到哪个分区。
	// cfg.Producer.Partitioner = sarama.NewCustomPartitioner()
	// // 比默认的 NewHashPartitioner 提供更多的灵活性，因为你可以完全控制如何计算消息的哈希值
	// cfg.Producer.Partitioner = sarama.NewCustomHashPartitioner(func() hash.Hash32 {
	//
	// })
	producer, err := sarama.NewSyncProducer(addrs, cfg)
	assert.NoError(t, err)

	// 在测试结束时关闭生产者
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatalf("Failed to close producer: %v", err)
		}
	}()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test_topic",
		Key:   sarama.StringEncoder("oid-123"),
		Value: sarama.StringEncoder("Hello, 第一条消息"),
		// 可以存放与消息内容相关的附加信息，通常用于跟踪、标识或其他业务场景。
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("trace_id"),
				Value: []byte("123456"),
			},
			{
				Key:   []byte("user_id"),
				Value: []byte("user-789"),
			},
			{
				Key:   []byte("version"),
				Value: []byte("v1"),
			},
		},
		Metadata: "这是 metadata",
	})
	assert.NoError(t, err)
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 表示当消息发送成功时，可以通过 Successes() 获取成功信息。
	cfg.Producer.Return.Successes = true
	// 默认启用。表示当发送消息失败时，可以通过 Errors() 获取错误信息。
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addrs, cfg)
	require.NoError(t, err)
	go func() {
		producer.Input() <- &sarama.ProducerMessage{
			Topic: "test_topic",
			Key:   sarama.StringEncoder("oid-123"),
			Value: sarama.StringEncoder("hello,异步发送"),
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("trace_id"),
					Value: []byte("123456"),
				},
			},
			Metadata: "这是 metadata",
		}
	}()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case res := <-producer.Successes():
			t.Logf("send success, %+v\n", res)
		case <-producer.Errors():
			t.Log("some error")
		case <-timeout:
			t.Log("timeout")
			return
		}
	}
}

func TestReadEvent(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addrs, cfg)
	require.NoError(t, err)
	for i := 0; i < 10000; i++ {
		part, offset, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: "article_read_event",
			Value: sarama.StringEncoder(`{"uid":6, "aid": 6}`),
		})
		t.Logf("part: %d, offset: %d, err: %v", part, offset, err)
	}
}

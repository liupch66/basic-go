package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

// CRUD Create（创建）、Read（读取）、Update（更新）和 Delete（删除）

func TestMongodb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		// 这个回调函数在 MongoDB 执行一个命令时被触发。
		// 它接收一个 CommandStartedEvent，该事件包含了命令的相关信息，如命令名称、执行时的时间戳、命令参数等。
		// 你可以使用这个事件来记录日志或执行其他操作，以便追踪每个数据库命令的执行情况。
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			// 坑:这里 startedEvent.Command 虽然是 []byte,但是转 string 会乱码,直接打印就好了
			// fmt.Println(string(startedEvent.Command))
			// fmt.Println(startedEvent.Command)
			t.Log(startedEvent.Command)
			t.Logf("Started command: %s, with request id: %d\n", startedEvent.CommandName, startedEvent.RequestID)
		},
		// 这个回调函数在 MongoDB 命令执行成功并且返回结果时被触发。
		// 它接收一个 CommandSucceededEvent，该事件包含命令执行成功时的详细信息，如执行时间、返回的结果等。
		// 你可以在这个函数里记录命令成功执行的日志，或者统计操作的耗时。
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
			t.Logf("Command %s succeeded, took %d ms\n", succeededEvent.CommandName, succeededEvent.Duration)
		},
		// 这个回调函数在 MongoDB 命令执行失败时被触发。
		// 它接收一个 CommandFailedEvent，该事件包含命令失败时的详细信息，如失败原因、错误代码、错误消息等。
		// 你可以使用这个事件来记录失败的命令，帮助你分析错误或进行异常处理。
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
			t.Logf("Command %s failed, error: %s\n", failedEvent.CommandName, failedEvent.Failure)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017").SetMonitor(monitor)
	client, err := mongo.Connect(opts)
	assert.NoError(t, err)
	// Make sure to defer a call to Disconnect after instantiating your client
	defer func() {
		fmt.Println("========================================================================================")
		err = client.Disconnect(ctx)
		assert.NoError(t, err)
	}()

	db := client.Database("webook")
	coll := db.Collection("articles")
	defer func() {
		fmt.Println("========================================================================================")
		// An empty document (e.g. bson.D{}) should be used to delete all documents in the collection
		deleteRes, err := coll.DeleteMany(ctx, bson.D{})
		assert.NoError(t, err)
		t.Logf("deleted documents count: %d\n", deleteRes.DeletedCount)
	}()

	fmt.Println("========================================================================================")
	res, err := coll.InsertOne(ctx, Article{
		Id:       123,
		Title:    "新建标题",
		Content:  "新建内容",
		AuthorId: 6,
		Status:   1,
		Ctime:    time.Now().UnixMilli(),
		Utime:    time.Now().UnixMilli(),
	})
	assert.NoError(t, err)
	t.Logf("inserted document with ID %v\n", res.InsertedID)

	/*
		bson.D 是一个 有序的 BSON 数据结构，通常用于需要保持字段顺序的情况。它实际上是一个结构体类型，底层是一个切片（[]bson.E），其中每个元素表示一个键值对。
			bson.D 中的每个元素是一个 bson.E 结构体，包含了 Key 和 Value 字段。
			bson.D 保留了插入字段的顺序，这在某些操作（如排序）中非常重要。
		使用场景：
			插入文档时，如果字段的顺序很重要，可以使用 bson.D。
			排序和聚合操作：在一些 MongoDB 操作中，顺序是重要的，比如 $push 操作中的数组顺序。

		bson.M 是一个 无序的 BSON 数据结构，底层实现是 map[string]interface{}。它的使用场景通常是数据的查询或更新操作，且无需保留字段顺序。
			bson.M 是一个映射（map）类型，它不保证字段的顺序。
			在大多数查询和更新场景中，顺序并不重要，因此 bson.M 更为常用。
		使用场景：
			查询和更新操作：对于大多数查询或更新操作，如果不需要关心字段的顺序，可以使用 bson.M。
			简化代码：在不需要严格控制顺序的情况下，bson.M 更加简洁。
	*/
	fmt.Println("========================================================================================")
	// 通过 bsonDE 查询
	var art Article
	// err = coll.FindOne(ctx, bson.D{bson.E{Key: "title", Value: "新建标题"}}).Decode(&art)
	err = coll.FindOne(ctx, bson.D{{"title", "新建标题"}}).Decode(&art)
	assert.NoError(t, err)
	t.Logf("got document by bsonDE: %#v\n", art)
	fmt.Println("========================================================================================")
	// 通过 bsonM 查询
	art = Article{}
	err = coll.FindOne(ctx, bson.M{"title": "新建标题"}).Decode(&art)
	assert.NoError(t, err)
	t.Logf("got document by bsonM: %#v\n", art)
	fmt.Println("========================================================================================")
	// 通过结构体查询,结构体一定要打标签 omitempty,不然会把零值也作为查询条件传入,返回错误:mongo.ErrNoDocuments
	art = Article{}
	err = coll.FindOne(ctx, Article{Title: "新建标题"}).Decode(&art)
	assert.NoError(t, err)
	t.Logf("got document by struct: %#v\n", art)
	fmt.Println("========================================================================================")
	var arts []Article
	cursor, err := coll.Find(ctx, bson.D{{"title", "新建标题"}})
	assert.NoError(t, err)
	err = cursor.All(ctx, &arts)
	assert.NoError(t, err)
	t.Logf("got documents by bsonDE: %#v\n", arts)

	fmt.Println("========================================================================================")
	updateRes, err := coll.UpdateMany(ctx, bson.M{"title": "新建标题"}, bson.M{"$set": bson.M{"title": "更新标题", "content": "更新内容"}})
	assert.NoError(t, err)
	t.Logf("updated documents count: %d\n", updateRes.ModifiedCount)

	fmt.Println("========================================================================================")
	// 类似的还有 $eq, $ne(not equal), $gt, $gte, $lt, $lte, $in, $nin(not in),  $exists, $not({ age: { $not: { $eq: 30 } } })
	arts = []Article{}
	inRes, err := coll.Find(ctx, bson.M{"id": bson.M{"$in": []int64{123, 234}}})
	assert.NoError(t, err)
	err = inRes.All(ctx, &arts)
	assert.NoError(t, err)
	t.Logf("got documents by $in: %#v\n", arts)
	fmt.Println("========================================================================================")
	// $and, $or, 这个和上面不同,是放在最前面的
	arts = []Article{}
	andRes, err := coll.Find(ctx, bson.M{"$and": []bson.M{{"title": "更新标题"}, {"content": "更新内容"}}})
	assert.NoError(t, err)
	err = andRes.All(ctx, &arts)
	assert.NoError(t, err)
	t.Logf("got documents by $and: %#v\n", arts)
	fmt.Println("========================================================================================")
	// 这里定义了索引的字段，即在 MongoDB 中创建索引的字段是 id。
	// bson.M{"id": 1} 表示按 id 字段升序 (1 表示升序，-1 表示降序)，可以理解为索引将会按 id 字段的升序排列数据。
	idxRes, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"id": 1},
		Options: options.Index().SetUnique(true),
	})
	assert.NoError(t, err)
	t.Logf("idxRes: %s\n", idxRes)
}

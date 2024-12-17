package article

import (
	"context"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBDAO struct {
	// client *mongo.Client
	// 代表 webook
	// database *mongo.Database
	// 制作表
	coll *mongo.Collection
	// 线上表
	liveColl *mongo.Collection
	node     *snowflake.Node
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBDAO{
		coll:     db.Collection("articles"),
		liveColl: db.Collection("published_articles"),
		node:     node,
	}
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	idx := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			// 这里不能用 bson.M{"author_id": 1, "ctime", 1}
			Keys:    bson.D{{"author_id", 1}, {"ctime", 1}},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().CreateMany(ctx, idx)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().CreateMany(ctx, idx)
	return err
}

func (m *MongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.coll.InsertOne(ctx, art)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *MongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	res, err := m.coll.UpdateOne(ctx, bson.M{"id": art.Id, "author_id": art.AuthorId}, bson.M{"$set": bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   time.Now().UnixMilli(),
	}})
	if err != nil {
		return err
	}
	// 这边就是校验 author_id 是不是合法
	if res.ModifiedCount == 0 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func (m *MongoDBDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	// TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = m.Insert(ctx, art)
	} else {
		// 这是雪花算法生成的 id
		err = m.UpdateById(ctx, art)
	}
	// art.Id = id 好像没啥用?
	// 这里有个 bug, id 永远是 0,因为没有生成 id
	now := time.Now().UnixMilli()
	art.Utime = now
	_, err = m.liveColl.UpdateOne(ctx, bson.M{"id": art.Id}, bson.M{
		// $set：更新已有文档的字段;如果文档不存在，则插入时设置 ctime 字段
		"$set":         PublishedArticle(art),
		"$setOnInsert": bson.M{"ctime": now},
	}, options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	// TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	// TODO implement me
	panic("implement me")
}

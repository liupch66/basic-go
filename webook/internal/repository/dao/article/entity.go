package article

// Article 制作表
type Article struct {
	Id    int64  `gorm:"primaryKey,Increment" bson:"id,omitempty"`
	Title string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	// 在 GORM 中，BLOB 是用于存储二进制数据的字段类型，通常用于存储图像、文件、加密数据等不适合存储为普通字符串的内容。
	// 在数据库中，BLOB（Binary Large Object）是一个数据类型，用来存储大块的二进制数据。
	Content  string `gorm:"type:blob" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

// PublishedArticle 衍生类型,自定义类型;还可以考虑组合和重新定义表结构
type PublishedArticle Article

package web

import (
	"github.com/liupch66/basic-go/webook/internal/domain"
)

// VO: view object, 就是对标前端的

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract string `json:"abstract"`
	// 内容
	Content    string `json:"content"`
	Status     uint8  `json:"status"`
	Author     string `json:"author"`
	Liked      bool   `json:"liked"`
	Collected  bool   `json:"collected"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uid},
	}
}

package job

import (
	"context"
	"time"

	"basic-go/webook/internal/service"
)

type RankJob struct {
	svc     service.RankService
	timeout time.Duration
}

func NewRankJob(svc service.RankService, timeout time.Duration) *RankJob {
	return &RankJob{svc: svc, timeout: timeout}
}

func (r *RankJob) Name() string {
	return "rank"
}

func (r *RankJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

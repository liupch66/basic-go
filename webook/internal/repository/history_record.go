package repository

import (
	"context"

	"github.com/liupch66/basic-go/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, r domain.HistoryRecord) error
}

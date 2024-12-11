package repository

import (
	"context"

	"basic-go/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, r domain.HistoryRecord) error
}

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"basic-go/webook/internal/domain"
	svcmocks "basic-go/webook/internal/service/mocks"
)

func TestBatchRankService_TopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name         string
		mock         func(ctrl *gomock.Controller) (ArticleService, InteractService)
		expectedErr  error
		expectedArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractService) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				interSvc := svcmocks.NewMockInteractService(ctrl)
				// 取第一批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Ctime: now, Utime: now},
						{Id: 2, Ctime: now, Utime: now},
						{Id: 3, Ctime: now, Utime: now},
					}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).
					Return(map[int64]domain.Interact{
						1: domain.Interact{BizId: 1, LikeCnt: 1},
						2: domain.Interact{BizId: 2, LikeCnt: 2},
						3: domain.Interact{BizId: 3, LikeCnt: 3},
					}, nil)

				// 取第二批，全被第一批取完了
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).
					Return([]domain.Article{}, nil)
				return artSvc, interSvc
			},
			expectedErr: nil,
			expectedArts: []domain.Article{
				{Id: 3, Ctime: now, Utime: now},
				{Id: 2, Ctime: now, Utime: now},
				{Id: 1, Ctime: now, Utime: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewBatchRankService(tc.mock(ctrl))
			// 方便测试
			svc.batchSize = 3
			svc.n = 3
			svc.scoreFunc = func(utime time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}

			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedArts, arts)
		})
	}
}

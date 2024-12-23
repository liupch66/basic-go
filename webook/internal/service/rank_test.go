package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	mockinteract "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1/mocks"
	"github.com/liupch66/basic-go/webook/internal/domain"
	svcmocks "github.com/liupch66/basic-go/webook/internal/service/mocks"
)

func TestBatchRankService_TopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name         string
		mock         func(ctrl *gomock.Controller) (ArticleService, interactv1.InteractServiceClient)
		expectedErr  error
		expectedArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, interactv1.InteractServiceClient) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				interSvc := mockinteract.NewMockInteractServiceClient(ctrl)
				// 取第一批
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Ctime: now, Utime: now},
						{Id: 2, Ctime: now, Utime: now},
						{Id: 3, Ctime: now, Utime: now},
					}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), gomock.Any()).
					Return(&interactv1.GetByIdsResponse{
						Interacts: map[int64]*interactv1.Interact{
							1: {BizId: 1, LikeCnt: 1},
							2: {BizId: 2, LikeCnt: 2},
							3: {BizId: 3, LikeCnt: 3},
						},
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

			artSvc, interSvc := tc.mock(ctrl)
			svc := &BatchRankService{
				artSvc:    artSvc,
				interSvc:  interSvc,
				repo:      nil,
				batchSize: 3,
				n:         3,
				scoreFunc: func(utime time.Time, likeCnt int64) float64 {
					return float64(likeCnt)
				},
			}

			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedArts, arts)
		})
	}
}

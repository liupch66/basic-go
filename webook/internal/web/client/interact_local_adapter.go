package client

import (
	"context"

	"google.golang.org/grpc"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	"github.com/liupch66/basic-go/webook/interact/domain"
	"github.com/liupch66/basic-go/webook/interact/service"
)

// InteractLocalAdapter 将本地实现伪装成 gRPC 客户端，因为 ArticleHandler 用的就是 InteractServiceClient
// 为了使得本地调用和 gRPC 调用可以并行，可以做流量分发，可以回滚，保证迁移顺利
type InteractLocalAdapter struct {
	svc service.InteractService
}

func NewInteractLocalAdapter(svc service.InteractService) *InteractLocalAdapter {
	return &InteractLocalAdapter{svc: svc}
}

func (i *InteractLocalAdapter) IncrReadCnt(ctx context.Context, in *interactv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interactv1.IncrReadCntResponse{}, err
}

func (i *InteractLocalAdapter) Like(ctx context.Context, in *interactv1.LikeRequest, opts ...grpc.CallOption) (*interactv1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interactv1.LikeResponse{}, err
}

func (i *InteractLocalAdapter) CancelLike(ctx context.Context, in *interactv1.CancelLikeRequest, opts ...grpc.CallOption) (*interactv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interactv1.CancelLikeResponse{}, err
}

func (i *InteractLocalAdapter) Collect(ctx context.Context, in *interactv1.CollectRequest, opts ...grpc.CallOption) (*interactv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &interactv1.CollectResponse{}, err
}

func (i *InteractLocalAdapter) Get(ctx context.Context, in *interactv1.GetRequest, opts ...grpc.CallOption) (*interactv1.GetResponse, error) {
	res, err := i.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactv1.GetResponse{Interact: i.toDTO(res)}, nil
}

func (i *InteractLocalAdapter) GetByIds(ctx context.Context, in *interactv1.GetByIdsRequest, opts ...grpc.CallOption) (*interactv1.GetByIdsResponse, error) {
	data, err := i.svc.GetByIds(ctx, in.GetBiz(), in.GetBizIds())
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*interactv1.Interact, len(data))
	for k, v := range data {
		res[k] = i.toDTO(v)
	}
	return &interactv1.GetByIdsResponse{Interacts: res}, nil
}

// DTO: Data Transfer Object
func (i *InteractLocalAdapter) toDTO(inter domain.Interact) *interactv1.Interact {
	return &interactv1.Interact{
		Biz:        inter.Biz,
		BizId:      inter.BizId,
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
		Liked:      inter.Liked,
		Collected:  inter.Collected,
	}
}

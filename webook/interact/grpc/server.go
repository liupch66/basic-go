package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
	"github.com/liupch66/basic-go/webook/interact/domain"
	"github.com/liupch66/basic-go/webook/interact/service"
)

// InteractServiceServer 这里只是把 service 包装成一个 grpc 而已，核心逻辑一定是在 service 里面，不关心上面是用 grpc, http 什么的
type InteractServiceServer struct {
	interactv1.UnimplementedInteractServiceServer
	svc service.InteractService
}

func NewInteractServiceServer(svc service.InteractService) *InteractServiceServer {
	return &InteractServiceServer{svc: svc}
}

func (i *InteractServiceServer) Register(server *grpc.Server) {
	interactv1.RegisterInteractServiceServer(server, i)
}

func (i *InteractServiceServer) IncrReadCnt(ctx context.Context, request *interactv1.IncrReadCntRequest) (*interactv1.IncrReadCntResponse, error) {
	// request.Biz 大部分时候也没有问题，为了万无一失最好还是 request.GetBiz()，这个有判断 request 是否是 nil
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBidId())
	// 标准写法
	// if err != nil {
	// 	return nil, err
	// }
	// return &interactv1.IncrReadCntResponse{}, nil
	// 偷懒写法
	return &interactv1.IncrReadCntResponse{}, err
}

func (i *InteractServiceServer) Like(ctx context.Context, request *interactv1.LikeRequest) (*interactv1.LikeResponse, error) {
	// 参数校验，也可以考虑在 buf.gen.yaml 中增加 validate 插件
	if request.Uid <= 0 {
		// return nil, errors.New("uid 非法")
		return nil, status.Error(codes.InvalidArgument, "uid 非法")
	}
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBidId(), request.GetUid())
	return &interactv1.LikeResponse{}, err
}

func (i *InteractServiceServer) CancelLike(ctx context.Context, request *interactv1.CancelLikeRequest) (*interactv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBidId(), request.GetUid())
	return &interactv1.CancelLikeResponse{}, err
}

func (i *InteractServiceServer) Collect(ctx context.Context, request *interactv1.CollectRequest) (*interactv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBidId(), request.GetUid(), request.GetCid())
	return &interactv1.CollectResponse{}, err
}

func (i *InteractServiceServer) Get(ctx context.Context, request *interactv1.GetRequest) (*interactv1.GetResponse, error) {
	res, err := i.svc.Get(ctx, request.GetBiz(), request.GetBidId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactv1.GetResponse{Interact: i.toDTO(res)}, nil
}

func (i *InteractServiceServer) GetByIds(ctx context.Context, request *interactv1.GetByIdsRequest) (*interactv1.GetByIdsResponse, error) {
	data, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
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
func (i *InteractServiceServer) toDTO(inter domain.Interact) *interactv1.Interact {
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

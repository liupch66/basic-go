package client

import (
	"context"
	"math/rand"

	"go.uber.org/atomic"
	"google.golang.org/grpc"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
)

type InteractGrayscaleRelease struct {
	remote interactv1.InteractServiceClient
	local  *InteractLocalAdapter
	// 随机数+阈值，用来做流量控制，取值是[0, 100)之间
	threshold *atomic.Int32
}

func NewInteractGrayscaleRelease(remote interactv1.InteractServiceClient, local *InteractLocalAdapter, threshold int32) *InteractGrayscaleRelease {
	return &InteractGrayscaleRelease{remote: remote, local: local, threshold: atomic.NewInt32(threshold)}
}

func (i *InteractGrayscaleRelease) IncrReadCnt(ctx context.Context, in *interactv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) Like(ctx context.Context, in *interactv1.LikeRequest, opts ...grpc.CallOption) (*interactv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) CancelLike(ctx context.Context, in *interactv1.CancelLikeRequest, opts ...grpc.CallOption) (*interactv1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) Collect(ctx context.Context, in *interactv1.CollectRequest, opts ...grpc.CallOption) (*interactv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) Get(ctx context.Context, in *interactv1.GetRequest, opts ...grpc.CallOption) (*interactv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) GetByIds(ctx context.Context, in *interactv1.GetByIdsRequest, opts ...grpc.CallOption) (*interactv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *InteractGrayscaleRelease) UpdateThreshold(newThreshold int32) {
	i.threshold.Store(newThreshold)
}

func (i *InteractGrayscaleRelease) selectClient() interactv1.InteractServiceClient {
	num := rand.Int31n(100)
	// 比如 threshold=10,那就是 [0,10),也就是 10% 的流量
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

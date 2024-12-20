package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
)

// 测试 grpc 的 server 端是否启动
func TestGrpcClient(t *testing.T) {
	cc, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactv1.NewInteractServiceClient(cc)
	resp, err := client.Get(context.Background(), &interactv1.GetRequest{
		Biz:   "test",
		BidId: 4,
		Uid:   123,
	})
	require.NoError(t, err)
	t.Log(resp.Interact)
}

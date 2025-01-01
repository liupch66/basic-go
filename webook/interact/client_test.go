package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	interactv1 "github.com/liupch66/basic-go/webook/api/proto/gen/interact/v1"
)

// 测试 gRPC 的 server 端是否启动
func TestGRPCClient(t *testing.T) {
	cc, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactv1.NewInteractServiceClient(cc)
	resp, err := client.Get(context.Background(), &interactv1.GetRequest{
		Biz:   "test",
		BizId: 1,
		Uid:   123,
	})
	require.NoError(t, err)
	t.Log(resp.Interact)
}

// 测试双写
// 分别 "post"
// localhost:8083/migrator/src_only
// localhost:8083/migrator/src_fist
// localhost:8083/migrator/dst_firs
// localhost:8083/migrator/dst_only
// 即可测试
// 增量校验和全量校验附在 validate_example.sql
func TestGRPCDoubleWrite(t *testing.T) {
	cc, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactv1.NewInteractServiceClient(cc)
	// 每次运行测试记得修改 bizId，每次加 1
	// 可以写个 for 循环
	_, err = client.IncrReadCnt(context.Background(), &interactv1.IncrReadCntRequest{
		Biz:   "test",
		BizId: 55,
	})
	require.NoError(t, err)
}

func TestEtcdGRPCClient(t *testing.T) {
	cli, err := clientv3.NewFromURL("localhost:22379")
	require.NoError(t, err)
	bd, err := resolver.NewBuilder(cli)
	require.NoError(t, err)
	cc, err := grpc.NewClient("etcd:///service/interact",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactv1.NewInteractServiceClient(cc)
	resp, err := client.GetByIds(context.Background(), &interactv1.GetByIdsRequest{
		Biz:    "test",
		BizIds: []int64{1, 2, 3, 4, 5, 6, 7, 8},
	})
	require.NoError(t, err)
	for k, v := range resp.Interacts {
		t.Logf("%d, %v\n", k, v)

	}
}

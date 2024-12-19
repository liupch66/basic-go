package userpb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientTokenAuth struct{}

func (c ClientTokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appId":  "liupch6",
		"appKey": "hello",
	}, nil
}

func (c ClientTokenAuth) RequireTransportSecurity() bool {
	return false
}

func TestClient(t *testing.T) {
	cc, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// cc, err := grpc.NewClient(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(ClientTokenAuth{}))
	require.NoError(t, err)
	defer cc.Close()
	client := NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &GetByIdReq{Id: 456})
	require.NoError(t, err)
	t.Log(resp.User)
}

package userpb

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	s := grpc.NewServer()
	defer func() {
		s.GracefulStop()
	}()
	RegisterUserServiceServer(s, &Server{})
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	if err = s.Serve(lis); err != nil {
		t.Log("启动失败：", err)
	}
}

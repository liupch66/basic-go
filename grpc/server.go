package userpb

import (
	"context"
	"fmt"
)

var _ UserServiceServer = (*Server)(nil)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, errors.New("未传输 token")
	// }
	// var appId, appKey string
	// if v, ok := md["appid"]; ok {
	// 	appId = v[0]
	// }
	// if v, ok := md["appkey"]; ok {
	// 	appKey = v[0]
	// }
	// if appId != "liupch6" || appKey != "hello" {
	// 	return nil, errors.New("错误 token")
	// }

	fmt.Println("Hello ", req.Id)
	return &GetByIdResp{
		User: &User{
			Id:   req.Id,
			Name: "测试用户 from" + s.Name,
		},
	}, nil
}

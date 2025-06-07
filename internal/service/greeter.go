package service

import (
	"context"
	pb "demo/api/v1"

	"demo/internal/biz"
)

// GreeterService is a greeter service.
type GreeterService struct {
	pb.UnimplementedGreeterSvcServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

func (g *GreeterService) ListGreeter(ctx context.Context, req *pb.ListGreeterReq) (*pb.ListGreeterRsp, error) {
	greeters, err := g.uc.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	rsp := &pb.ListGreeterRsp{
		Body: &pb.ListGreeterRsp_Body{
			Greeters: convertGreeterDoToPbs(greeters),
		},
	}
	return rsp, nil
}

func convertGreeterDoToPbs(gs []*biz.Greeter) []*pb.Greeter {
	ret := make([]*pb.Greeter, 0)
	for _, g := range gs {
		ret = append(ret, convertGreeterDoToPb(g))
	}
	return ret
}

func convertGreeterDoToPb(g *biz.Greeter) *pb.Greeter {
	return &pb.Greeter{
		Name: g.Name,
		Age:  g.Age,
	}
}

package service

import (
	"context"
	v2 "demo/api/v1"

	"demo/internal/biz"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v2.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v2.HelloRequest) (*v2.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &v2.HelloReply{Message: "Hello " + g.Hello}, nil
}

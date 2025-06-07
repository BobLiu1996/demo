package service

import (
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

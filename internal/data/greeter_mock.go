package data

import (
	"context"
	"demo/internal/biz"
)

type mockGreeterRepo struct{}

func NewMockGreeterRepo() biz.GreeterRepo {
	return &mockGreeterRepo{}
}

func (m mockGreeterRepo) Save(ctx context.Context, greeter *biz.Greeter) (*biz.Greeter, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockGreeterRepo) Update(ctx context.Context, greeter *biz.Greeter) (*biz.Greeter, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockGreeterRepo) FindByID(ctx context.Context, i int64) (*biz.Greeter, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockGreeterRepo) ListByHello(ctx context.Context, s string) ([]*biz.Greeter, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockGreeterRepo) ListAll(ctx context.Context) ([]*biz.Greeter, error) {
	//TODO implement me
	panic("implement me")
}

package data

import (
	"context"
	"demo/internal/biz"
)

type greeterRepo struct {
	data *Data
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data) biz.GreeterRepo {
	return &greeterRepo{
		data: data,
	}
}

func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) Update(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) FindByID(context.Context, int64) (*biz.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListByHello(context.Context, string) ([]*biz.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListAll(ctx context.Context) ([]*biz.Greeter, error) {
	tGreeter := r.data.Query().TGreeter
	res := make([]*biz.Greeter, 0)
	if err := tGreeter.WithContext(ctx).Select(tGreeter.Name, tGreeter.Age).Scan(&res); err != nil {
		return nil, err
	}
	return res, nil
}

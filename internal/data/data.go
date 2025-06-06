package data

import (
	"demo/internal/biz"
	"demo/internal/conf"
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, ProvideGreeterRepo)

func ProvideGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
	switch data.dataCfg.RepoSelector {
	case "mysql":
		return NewGreeterRepo(data, logger)
	case "mock":
		return NewMockGreeterRepo()
	default:
		panic("unknown user repo type")
	}
}

// Data .
type Data struct {
	dataCfg *conf.Data
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		dataCfg: c,
	}, cleanup, nil
}

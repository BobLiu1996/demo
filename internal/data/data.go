package data

import (
	"context"
	"demo/internal/biz"
	"demo/internal/conf"
	"demo/internal/data/dao"
	"demo/pkg/client/db"
	"github.com/google/wire"
	"gorm.io/gorm"

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

type ContextTxKey struct{}

// Data .
type Data struct {
	dataCfg  *conf.Data
	mysqlCli *gorm.DB
	query    *dao.Query
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	d := &Data{
		dataCfg: c,
	}
	if err := d.initMysql(); err != nil {
		return nil, nil, err
	}
	return d, cleanup, nil
}

func (d *Data) initMysql() error {
	if mysqlCfg := d.dataCfg.GetMysql(); mysqlCfg != nil {
		if db, err := db.NewMysqlClient(
			db.WithSource(mysqlCfg.GetSource()),
			db.WithMaxConn(int(mysqlCfg.GetMaxConn())),
			db.WithMaxIdleConn(int(mysqlCfg.GetMaxIdleConn())),
			db.WithMaxLifeTime(mysqlCfg.GetMaxLifetime().AsDuration()),
		); err != nil {
			return err
		} else {
			d.mysqlCli = db
			if d.dataCfg.GetDebug() {
				db = db.Debug()
			}
			d.query = dao.Use(db)
		}
	}
	return nil
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := d.mysqlCli.WithContext(ctx)
	if d.dataCfg.GetDebug() {
		db = db.Debug()
	}
	return db.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, ContextTxKey{}, tx)
		return fn(ctx)
	})
}

func (d *Data) Mysql(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(ContextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	db := d.mysqlCli.WithContext(ctx)
	if d.dataCfg.GetDebug() {
		db = db.Debug()
	}
	return db
}

func (d *Data) Query() *dao.Query {
	return d.query
}

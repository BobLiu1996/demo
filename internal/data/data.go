package data

import (
	"context"
	"demo/internal/biz"
	"demo/internal/conf"
	"demo/internal/data/dao"
	"demo/pkg/client/cache"
	"demo/pkg/client/db"
	plog "demo/pkg/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, ProvideGreeterRepo)

func ProvideGreeterRepo(data *Data) biz.GreeterRepo {
	switch data.dataCfg.RepoSelector {
	case "mysql":
		return NewGreeterRepo(data)
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
	redisCli *cache.RedisResource
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	d := &Data{
		dataCfg: c,
	}
	if err := d.initMysql(); err != nil {
		return nil, nil, err
	}
	if err := d.initRedis(); err != nil {
		return nil, nil, err
	}
	return d, d.cleanup, nil
}

func (d *Data) cleanup() {
	var ctx = context.Background()
	plog.Info(ctx, "closing the data resources")
	if d.redisCli != nil {
		if err := d.redisCli.CloseRedisClient(); err != nil {
			plog.Errorf(ctx, "redis cluster client close err:%s", err)
		}
	}
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

func (d *Data) initRedis() error {
	if redisCfg := d.dataCfg.GetRedis(); redisCfg != nil {
		if rdb, err := cache.NewRedisResource(&cache.RedisConf{
			Addr:         redisCfg.GetAddr(),
			User:         redisCfg.GetUsername(),
			Password:     redisCfg.GetPassword(),
			Db:           redisCfg.GetDb(),
			Pool:         redisCfg.GetPool(),
			ReadTimeout:  redisCfg.GetReadTimeout(),
			WriteTimeout: redisCfg.GetWriteTimeout(),
			IsCluster:    redisCfg.GetIsCluster(),
		}); err != nil {
			return err
		} else {
			d.redisCli = rdb
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

func (d *Data) Redis() redis.Cmdable {
	return d.redisCli.RedisClient()
}

func (d *Data) Query() *dao.Query {
	return d.query
}

package conf

import (
	"context"
	// "reflect"
	"sync"

	plog "demo/pkg/log"

	"github.com/go-kratos/kratos/contrib/config/apollo/v2"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

type AppConfig struct {
	conf      config.Config
	data      *sync.Map
	listeners map[string][]Listener
}

func NewAppConfig(sourceCfg *Config_Source) *AppConfig {
	var source config.Source
	if sourceCfg.Apollo != nil {
		source = apollo.NewSource(
			apollo.WithAppID(sourceCfg.Apollo.AppId),
			apollo.WithCluster(sourceCfg.Apollo.Cluster),
			apollo.WithEndpoint(sourceCfg.Apollo.Endpoint),
			apollo.WithNamespace(sourceCfg.Apollo.Namespace),
			apollo.WithOriginalConfig(),
			apollo.WithEnableBackup(),
			apollo.WithBackupPath(sourceCfg.Apollo.BackupPath),
			apollo.WithSecret(sourceCfg.Apollo.Secret),
		)
	} else {
		source = file.NewSource(sourceCfg.File)
	}
	conf := config.New(config.WithSource(source))
	if err := conf.Load(); err != nil {
		conf.Close()
		panic(err)
	}
	return &AppConfig{conf: conf, listeners: map[string][]Listener{}, data: new(sync.Map)}
}

func (app *AppConfig) GetBootstrap() *Bootstrap {
	var bs Bootstrap
	if err := app.conf.Scan(&bs); err != nil {
		app.conf.Close()
		panic(err)
	}
	if bs.Data == nil {
		app.conf.Close()
		panic("Initialization error for data is empty ")
	}
	return &bs
}

func (app *AppConfig) handlerDataChange(key string, value config.Value) {
	for _, listener := range app.listeners[key] {
		listener.Notify(value)
	}
}

func (app *AppConfig) Register(key string, listener Listener) {
	var ctx = context.Background()
	if listener == nil || len(key) == 0 {
		return
	}
	app.listeners[key] = append(app.listeners[key], listener)
	if err := app.conf.Watch(key, app.handlerDataChange); err != nil {
		plog.Errorf(ctx, "watch failure err:%v", err)
	}
}

type Listener interface {
	Notify(value interface{})
}

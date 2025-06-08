package main

import (
	"context"
	"demo/internal/server"
	"flag"
	"github.com/go-kratos/kratos/v2/config/env"
	"os"

	"demo/internal/conf"
	plog "demo/pkg/log"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "demo"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cronServer *server.CronServerImpl) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
			cronServer,
		),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			env.NewSource("APOLLO_"),
			file.NewSource(flagconf),
		),
	)
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}
	var bc conf.Config
	var ctx = context.Background()
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	logger := plog.NewLogger(Name, bc.Zap.File, bc.Zap.MaxSize, bc.Zap.MaxBackups, bc.Zap.MaxAge,
		plog.WithLevel(bc.Zap.Level), plog.WithConsole(bc.Zap.Console))
	plog.Info(ctx, "init logger.")

	os.Chmod(bc.Zap.File, 0644)

	appConf := conf.NewAppConfig(bc.Source)
	bs := appConf.GetBootstrap()
	app, cleanup, err := wireApp(bs.Server, bs.Data, appConf, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

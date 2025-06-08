package task

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	tLog := log.NewHelper(log.NewStdLogger(os.Stdout))
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx := context.Background()
	srv := NewServer(
		Logger(tLog),
		WithContext(context.Background()),
		//WithRedisLock(redis.Options{}, "test"),
	)
	srv.AddCronTask(
		WithTaskFunc(doWork1(tLog)),
		WithRunSpec("@every 1s"),
		WithTaskName("doWork1"),
		WithTimeout(3*time.Second),
	)
	srv.AddCronTask(
		WithTaskFunc(doWork2(tLog)),
		WithRunSpec("@every 3s"),
		WithTaskName("doWork2"),
		WithTimeout(5*time.Second),
	)
	//srv.AddCronTask(
	//	WithTaskFunc(TroubleMaker),
	//	WithRunSpec("*/10 * * * * *"),
	//	WithTaskName("TroubleMaker"),
	//)
	//srv.AddCronTask(
	//	WithTaskFunc(PanicMaker),
	//	WithRunSpec("@every 3s"),
	//	WithTaskName("PanicMaker"),
	//)
	srv.AddCronTask(
		WithTaskFunc(TimeoutMaker(tLog)),
		WithRunSpec("@every 10s"),
		WithTaskName("TimeoutMaker"),
		WithTimeout(3*time.Second),
	)

	if err := srv.Start(ctx); err != nil {
		panic(err)
	}
	// 起一个goroutine，分批停止定时任务
	go func() {
		time.Sleep(3 * time.Second)
		tLog.Log(log.LevelInfo, "remove doWork1", "done")
		srv.RemoveCronTask("doWork1")
		time.Sleep(30 * time.Second)
		srv.Stop(ctx)
		interrupt <- syscall.SIGTERM
	}()

	defer func() {
		if err := srv.Stop(ctx); err != nil {
			t.Errorf("expected nil got %v", err)
		}
	}()
	<-interrupt
}

func doWork1(tlog *log.Helper) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		tlog.Info("doWork1", "@every 1s")
		return nil
	}
}

func doWork2(tlog *log.Helper) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		tlog.Info("doWork2", "@every 3s")
		return nil
	}
}

func TroubleMaker(ctx context.Context) error {
	return errors.New("OH!! Fuck TroubleMaker")
}

func PanicMaker(ctx context.Context) error {
	panic("OH!! Fuck PanicMaker")
	return nil
}

// 睡眠大于ctx超时时间的定时任务测试
func TimeoutMaker(tlog *log.Helper) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// 模拟一个耗时操作
		sleepDuration := 5 * time.Second
		tlog.Infof("Task started, will sleep for %v\n", sleepDuration)
		time.Sleep(sleepDuration)

		// 检查上下文是否已超时
		select {
		case <-ctx.Done():
			tlog.Info("Task aborted due to timeout")
			return ctx.Err()
		default:
			tlog.Info("Task completed successfully")
			return nil
		}
	}
}

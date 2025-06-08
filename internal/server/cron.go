package server

import (
	"context"
	"demo/internal/server/wire"
	plog "demo/pkg/log"
	"fmt"
	"reflect"
	"time"

	"demo/internal/conf"
	"demo/internal/data"

	redis_lock "demo/pkg/client/cache/locker"
	"demo/pkg/server/task"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/redis/go-redis/v9"
)

// ------------------- 业务说明 -------------------
// 从配置文件中读取定时任务配置，初始化定时任务
// 配置文件中的定时任务配置格式如下：
// cron_tasks:
//   - name: SyncWeChatData  // 任务名称，对应CronService中的方法名
//     spec: "1 1 0 * * ?"   // 任务执行时间规则 https://en.wikipedia.org/wiki/Cron
//     timeout: 30s          // 任务执行超时时间, 为0则不设置超时时间(且并发锁不会生效)
//
// 任务配置会根据配置文件的变更动态加载，不需要重启服务
// 传入的Context具有WithTimeout属性，任务超时会被取消，请保证任务的原子性
// 错误的任务配置加载会导致panic
// ------------------- 业务说明 -------------------

// UpdateTaskStatue 更新任务状态
// func (cs *CronService) UpdateTaskStatue(ctx context.Context) error {
// 	err := cs.useCase.UpdateState(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

type CronServerImpl struct {
	*task.Server
	service     wire.CronService
	cronTaskCfg []*conf.CronTask
}

func NewCronServer(cronSvc wire.CronService, data *data.Data, config *conf.AppConfig) *CronServerImpl {
	locker := NewLocker(data.Redis())
	srv := task.NewServer(
		task.WithContext(context.Background()),
		task.WithLocker(locker, "cron_task"),
	)
	serverCfg := config.GetBootstrap().GetServer()
	cronServer := &CronServerImpl{
		service:     cronSvc,
		Server:      srv,
		cronTaskCfg: serverCfg.GetCronTasks(),
	}
	if err := cronSvc.InitService(); err != nil {
		panic(err)
	}
	if err := cronServer.addCronTask(); err != nil {
		plog.Error(context.Background(), err)
		return nil
	}
	config.Register("server", cronServer)
	return cronServer
}

// Notify 动态加载配置
func (s *CronServerImpl) Notify(value interface{}) {
	if value == nil {
		return
	}
	v, ok := value.(config.Value)
	if !ok {
		return
	}
	var data conf.Server
	if err := v.Scan(&data); err != nil {
		return
	}
	if reflect.DeepEqual(s.cronTaskCfg, data.CronTasks) {
		return
	}
	s.cronTaskCfg = data.CronTasks
	s.Server.RemoveAllTask()
	s.addCronTask()
}

// addCronTask 从配置文件中读取添加定时任务
func (s *CronServerImpl) addCronTask() error {
	value := reflect.ValueOf(s.service)
	for _, cronTask := range s.cronTaskCfg {
		method := value.MethodByName(cronTask.GetName())
		if !method.IsValid() {
			continue
		}
		m, ok := method.Interface().(func(context.Context) error)
		if !ok {
			return fmt.Errorf("cron function type error: %s", cronTask.GetName())
		}
		_, err := s.Server.AddCronTask(
			task.WithTaskFunc(m),
			task.WithRunSpec(cronTask.GetSpec()),
			task.WithTaskName(cronTask.GetName()),
			task.WithTimeout(cronTask.GetTimeout().AsDuration()),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

type Locker struct {
	*redis_lock.Client
}

func NewLocker(client redis.Cmdable) *Locker {
	rclient := redis_lock.NewClient(client)
	return &Locker{Client: rclient}
}

func (l *Locker) Lock(ctx context.Context, key string, timeout time.Duration) (unlock func(context.Context) error, err error) {
	lock, err := l.Client.TryLock(ctx, key, timeout)
	if err != nil {
		return nil, err
	}
	return lock.Unlock, nil
}

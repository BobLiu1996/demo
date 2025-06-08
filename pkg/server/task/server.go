package task

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"demo/pkg/log"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/context"
)

var (
	// "归一化"写法:
	// Server结构体需要实现 transport.Server 这个interface对应的方法
	_ transport.Server = (*Server)(nil)
)

// Server 定时任务服务
type Server struct {
	locker     Locker
	keyPrefix  string
	cron       *cron.Cron
	baseCtx    context.Context
	started    bool
	mapEntryId map[string]cron.EntryID
	err        error
	log        *klog.Helper
}

type funcType func(context.Context) error

// Task 定时任务主体
type Task struct {
	name    string        // 任务名称
	entryID cron.EntryID  // 任务ID
	tFunc   funcType      // 任务执行函数
	spec    string        // 任务执行时间规则
	timeout time.Duration // 任务执行超时时间
	err     error         // 任务执行错误
	errMsg  string        // 任务执行错误信息
}

func NewServer(opts ...ServerOptions) *Server {
	srv := &Server{
		cron:       cron.New(cron.WithSeconds()),
		baseCtx:    context.Background(),
		mapEntryId: make(map[string]cron.EntryID),
		started:    false,
	}
	for _, o := range opts {
		o(srv)
	}
	return srv
}

func (s *Server) Name() string {
	return "CronTaskServer"
}

func (s *Server) GetMapEntryId() map[string]cron.EntryID {
	return s.mapEntryId
}

// AddCronTask is a method of the Server struct that adds a new cron task.
// It accepts a variadic parameter of RunOptions, which are used to configure the task.
// The method returns the ID of the created task and an error if any occurred during the task creation.
//
// The method works as follows:
// 1. It creates a new task using the provided options.
// 2. If an error occurred during the task creation, it returns the error.
// 3. It adds the created task to the cron scheduler.
// 4. If an error occurred while adding the task to the scheduler, it returns the error.
// 5. It logs the created task.
// 6. Finally, it returns the ID of the created task and nil as the error.
//
// Usage:
// entryID, err := server.AddCronTask(
//
//	WithTaskName("MyTask"),
//	WithTaskFunc(myFunc),
//	WithRunSpec("* * * * * ?"),
//	WithTimeout(time.Second * 30),
//
// )
//
//	if err != nil {
//	    log.Fatalf("Failed to add cron task: %v", err)
//	}
func (s *Server) AddCronTask(opts ...RunOptions) (cron.EntryID, error) {
	task := s.createTask(opts...)
	if task.err != nil {
		return 0, task.err
	}

	entryID, err := s.addTaskToCron(task)
	if err != nil {
		return 0, err
	}

	log.Infof(s.baseCtx, "[cron-task] AddCronTask: %+v", task)
	return entryID, nil
}

func (s *Server) createTask(opts ...RunOptions) *Task {
	task := &Task{timeout: 0}
	for _, o := range opts {
		if err := o(task); err != nil {
			task.err = err
			return task
		}
	}
	if task.tFunc == nil {
		task.err = errors.WithStack(ErrTaskFuncIsNil)
		return task
	}
	if task.spec == "" {
		task.err = errors.WithStack(ErrTaskSpecIsNil)
		return task
	}
	if task.name == "" {
		task.name = runtime.FuncForPC(reflect.ValueOf(task.tFunc).Pointer()).Name()
	}
	return task
}

func (s *Server) addTaskToCron(task *Task) (cron.EntryID, error) {
	entryID, err := s.cron.AddFunc(task.spec, s.wrapFunc(task))
	if err != nil {
		return 0, err
	}
	task.entryID = entryID
	s.mapEntryId[task.name] = entryID
	return entryID, nil
}

// wrapFunc is a method of the Server struct that wraps a Task's function with additional functionality.
// It accepts a Task as a parameter and returns a function.
//
// The returned function does the following when called:
//  1. It records the start time of the task.
//  2. It defers a function that does the following:
//     a. It recovers from any panic that might occur during the task execution and logs the panic.
//     b. It calculates the elapsed time since the start of the task.
//     c. If an error occurred during the task execution, it logs the error and resets the task's error.
//     d. If no error occurred, it logs the successful completion of the task.
//  3. It attempts to acquire a Redis lock to prevent concurrent execution of the task.
//     a. If it fails to acquire the lock, it logs the failure and returns.
//     b. If it succeeds in acquiring the lock, it defers a function that releases the lock.
//  4. It sets a timeout for the task execution if a timeout is specified in the task.
//  5. It executes the task's function.
//     a. If an error occurs during the execution, it sets the error on the task.
func (s *Server) wrapFunc(task *Task) func() {
	return func() {
		// 记录任务耗时
		start := time.Now()
		defer func() {
			// 捕获panic
			if r := recover(); r != nil {
				task.err = errors.Errorf("Recover from panic --> %+v", r)
				task.errMsg = fmt.Sprintf("panic stack -->\n%s", string(debug.Stack()))
			}

			elapsed := time.Since(start)
			elapsedInSec := float64(elapsed) / float64(time.Second)
			if task.err != nil {
				log.Errorf(s.baseCtx, "[cron-task] Task Failed: %s [error]: %v [msg]: %s [cost]: %.4fs",
					task.name, task.err, task.errMsg, elapsedInSec)
				task.err = nil
				task.errMsg = ""
				return
			}
			log.Infof(s.baseCtx, "[cron-task] Task Finshed: %s [cost]: %.4fs", task.name, elapsedInSec)
		}()

		// 使用redis锁防止并发执行任务
		if s.locker != nil && len(s.keyPrefix) > 0 && task.timeout > 0 {
			redisKey := fmt.Sprintf("%s_%s", s.keyPrefix, task.name)
			unlock, err := s.locker.Lock(s.baseCtx, redisKey, task.timeout+1*time.Second)
			if err != nil {
				log.Warnf(s.baseCtx, "[cron-task] %s lock fail: %v", task.name, err)
				return
			} else {
				defer func() {
					err := unlock(s.baseCtx)
					if err != nil {
						task.err = err
						task.errMsg = fmt.Sprintf("unlock fail")
					}
				}()
			}
		}

		// 设置超时时间
		ctxInFunc := s.baseCtx
		if task.timeout > 0 {
			ctxInFuncWithTimeOut, cancel := context.WithTimeout(s.baseCtx, task.timeout)
			defer func() {
				cancel()
			}()
			ctxInFunc = ctxInFuncWithTimeOut
			go func() {
				select {
				case <-ctxInFunc.Done():
					if errors.Is(ctxInFunc.Err(), context.DeadlineExceeded) {
						log.Errorf(s.baseCtx, "[cron-task] Task Timeout: %s [cost]: %.4fs", task.name, task.timeout.Seconds())
					}
				}
			}()
		}
		// 执行定时任务
		if err := task.tFunc(ctxInFunc); err != nil {
			task.err = err
			task.errMsg = fmt.Sprintf("task exec fail")
		}
	}
}

// RemoveCronTask is a method of the Server struct that removes a cron task.
// It accepts the name of the task to be removed and returns an error if any occurred during the task removal.
//
// The method works as follows:
// 1. It retrieves the ID of the task to be removed from the MapEntryId map.
// 2. If the task ID is not found in the map, it returns an error.
// 3. It removes the task from the cron scheduler using the retrieved ID.
// 4. If an error occurred while removing the task from the scheduler, it returns the error.
// 5. It logs the removal of the task.
// 6. Finally, it returns nil as the error.
//
// Usage:
// err := server.RemoveCronTask("MyTask")
//
//	if err != nil {
//	    log.Fatalf("Failed to remove cron task: %v", err)
//	}
func (s *Server) RemoveCronTask(name string) error {
	entryID, ok := s.mapEntryId[name]
	if !ok {
		return ErrTaskNotFound
	}
	s.cron.Remove(entryID)
	delete(s.mapEntryId, name)
	log.Infof(s.baseCtx, "[cron-task] RemoveCronTask: %s", name)
	return nil
}

// RemoveAllTask is a method of the Server struct that removes all cron tasks.
func (s *Server) RemoveAllTask() {
	for name, entryID := range s.mapEntryId {
		s.cron.Remove(entryID)
		delete(s.mapEntryId, name)
	}
	log.Infof(s.baseCtx, "[cron-task] Empty Cron Tasks")
}

func (s *Server) Start(ctx context.Context) error {
	if s.err != nil {
		return s.err
	}
	if s.started {
		return nil
	}
	s.cron.Start()
	log.Info(ctx, "[cron-task] server starting")
	s.started = true
	return nil
}

func (s *Server) Stop(_ context.Context) error {
	if !s.started {
		return nil
	}
	log.Info(s.baseCtx, "[cron-task] server stopping")
	s.started = false
	s.cron.Stop()
	return nil
}

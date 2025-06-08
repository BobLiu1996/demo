package task

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type ServerOptions func(o *Server) error

func WithContext(ctx context.Context) ServerOptions {
	return func(s *Server) error {
		s.baseCtx = ctx
		return nil
	}
}

func WithLocker(opt Locker, keyPrefix string) ServerOptions {
	return func(s *Server) error {
		s.locker = opt
		s.keyPrefix = keyPrefix
		return nil
	}
}

func Logger(logger *log.Helper) ServerOptions {
	return func(s *Server) error {
		s.log = logger
		return nil
	}
}

type RunOptions func(o *Task) error

// WithTaskName 设置任务名称
// Default {funcName}
func WithTaskName(name string) RunOptions {
	return func(o *Task) error {
		o.name = name
		return nil
	}
}

// WithTaskFunc 设置任务执行函数
// Must be set
func WithTaskFunc(fun funcType) RunOptions {
	return func(o *Task) error {
		o.tFunc = fun
		return nil
	}
}

// WithRunSpec 设置任务执行时间规则
// Must be set
func WithRunSpec(spec string) RunOptions {
	return func(o *Task) error {
		o.spec = spec
		return nil
	}
}

// WithTimeout 设置任务执行超时时间(单位秒),用于redisLock
// Default 30s
func WithTimeout(timeout time.Duration) RunOptions {
	return func(o *Task) error {
		o.timeout = timeout
		return nil
	}
}

type Locker interface {
	Lock(ctx context.Context, key string, timeout time.Duration) (unlock func(context.Context) error, err error)
}

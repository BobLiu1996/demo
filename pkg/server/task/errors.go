package task

import "fmt"

var (
	ErrRedisNil      = fmt.Errorf("redis is empty")
	ErrRedisInvalid  = fmt.Errorf("redis is invalid")
	ErrExprInvalid   = fmt.Errorf("expr is invalid")
	ErrSaveCron      = fmt.Errorf("save cron failed")
	ErrTaskFuncIsNil = fmt.Errorf("task func is nil")
	ErrTaskSpecIsNil = fmt.Errorf("task spec is nil")
	ErrTaskNotFound  = fmt.Errorf("task not found")
)

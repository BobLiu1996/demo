package log

import (
	"context"
	"fmt"

	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultMessageKey default message key.
var DefaultMessageKey = "M"
var sprint func(...interface{}) string
var sprintf func(format string, a ...interface{}) string

var (
	logger log.Logger
	once   sync.Once
)

type Options struct {
	level   zapcore.Level
	console bool
}

func init() {
	l := NewZapExample()
	logger = log.With(l,
		"LFILE", log.DefaultCaller,
		"TS", log.DefaultTimestamp,
	)
}

type Option func(o *Options)

// Log Print log by level and keyvals.
func Log(ctx context.Context, level log.Level, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(level, keyvals...)
}

// Debug logs a message at debug level.
func Debug(ctx context.Context, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelDebug, DefaultMessageKey, sprint(a...))
}

// Debugf logs a message at debug level.
func Debugf(ctx context.Context, format string, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelDebug, DefaultMessageKey, sprintf(format, a...))
}

// Debugw logs a message at debug level.
func Debugw(ctx context.Context, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelDebug, keyvals...)
}

// Info logs a message at info level.
func Info(ctx context.Context, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelInfo, DefaultMessageKey, sprint(a...))
}

// Infof logs a message at info level.
func Infof(ctx context.Context, format string, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelInfo, DefaultMessageKey, sprintf(format, a...))
}

// Infow logs a message at info level.
func Infow(ctx context.Context, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelInfo, keyvals...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelWarn, DefaultMessageKey, sprint(a...))
}

// Warnf logs a message at warnf level.
func Warnf(ctx context.Context, format string, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelWarn, DefaultMessageKey, sprintf(format, a...))
}

// Warnw logs a message at warnf level.
func Warnw(ctx context.Context, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelWarn, keyvals...)
}

// Error logs a message at error level.
func Error(ctx context.Context, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelError, DefaultMessageKey, sprint(a...))
}

// Errorf logs a message at error level.
func Errorf(ctx context.Context, format string, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelError, DefaultMessageKey, sprintf(format, a...))
}

// Errorw logs a message at error level.
func Errorw(ctx context.Context, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelError, keyvals...)
}

// Fatal logs a message at fatal level.
func Fatal(ctx context.Context, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelFatal, DefaultMessageKey, sprint(a...))
	os.Exit(1)
}

// Fatalf logs a message at fatal level.
func Fatalf(ctx context.Context, format string, a ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelFatal, DefaultMessageKey, sprintf(format, a...))
	os.Exit(1)
}

// Fatalw logs a message at fatal level.
func Fatalw(ctx context.Context, keyvals ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelFatal, keyvals...)
	os.Exit(1)
}

func NewLogger(appName, fileName string, maxSize, maxBackups, MaxAge int32, opt ...Option) log.Logger {
	once.Do(func() {
		var opts = Options{
			level: zapcore.DebugLevel,
		}
		for _, o := range opt {
			o(&opts)
		}
		logger = initLog(appName, fileName, maxSize, maxBackups, MaxAge, opts)
	})
	log.SetLogger(logger)
	return logger
}

func initLog(appName, fileName string, maxSize, maxBackups, MaxAge int32, opts Options) log.Logger {
	encoder := zapcore.EncoderConfig{
		//TimeKey:        "ts",
		NameKey:        "logger",
		StacktraceKey:  "stack",
		LevelKey:       "L",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	zLogger := NewZapLogger(
		encoder,
		zap.NewAtomicLevelAt(opts.level),
		fileName,
		int(maxSize),
		int(maxBackups),
		int(MaxAge),
		opts.console,
		zap.AddStacktrace(
			zap.NewAtomicLevelAt(zapcore.FatalLevel)),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.Development(),
	)

	logger = log.With(zLogger,
		"LFILE", log.DefaultCaller,
		"TS", log.DefaultTimestamp,
		"TRACE_ID", tracing.TraceID(),
		"SPAN_ID", tracing.SpanID(),
		"LAPP", appName,
	)
	sprint = fmt.Sprint
	sprintf = fmt.Sprintf
	return logger
}

func WithLevel(level string) Option {
	return func(o *Options) {
		l, err := zapcore.ParseLevel(level)
		if err == nil {
			o.level = zapcore.Level(l)
		}
	}
}

func WithConsole(console bool) Option {
	return func(o *Options) {
		o.console = console
	}
}

func WithLogger() {
	kratos.New(kratos.Logger(logger))
}

func GetLogger() log.Logger {
	return logger
}

// SetLogger 外部注入日志句柄，无需新建初始化一个
func SetLogger(l log.Logger) {
	once.Do(func() {
		logger = l
		//log.SetLogger(logger)
		sprint = fmt.Sprint
		sprintf = fmt.Sprintf
	})
}

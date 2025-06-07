package log

import (
	"context"
	"fmt"
	gorm_log "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
	// 默认配置 gorm_log.Info
	LogLevel gorm_log.LogLevel
}

func NewZapExample() *ZapLogger {
	e := zap.NewExample()
	return &ZapLogger{log: e, Sync: e.Sync, LogLevel: gorm_log.Info}
}

// NewZapLogger return a zap logger.
func NewZapLogger(encoder zapcore.EncoderConfig, level zap.AtomicLevel, name string, maxFileSize, MaxBackups, maxAge int, console bool, opts ...zap.Option) *ZapLogger {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   name,
		MaxSize:    maxFileSize,
		MaxBackups: MaxBackups,
		MaxAge:     maxAge,
		Compress:   true,
		LocalTime:  true,
	}
	writerSyncs := make([]zapcore.WriteSyncer, 0, 2)
	writerSyncs = append(writerSyncs, zapcore.AddSync(lumberJackLogger))
	if console {
		writerSyncs = append(writerSyncs, zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(writerSyncs...), level)
	zapLogger := zap.New(core, opts...)
	return &ZapLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}
	// Zap.Field is used when keyvals pairs appear
	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("", data...)
	case log.LevelInfo:
		l.log.Info("", data...)
	case log.LevelWarn:
		l.log.Warn("", data...)
	case log.LevelError:
		l.log.Error("", data...)
	case log.LevelFatal:
		l.log.Fatal("", data...)
	}
	return nil
}

// !______________补充接口实现,适配GORM______________________!
// https://gorm.io/zh_CN/docs/logger.html
// Logger 需要实现以下接口，它接受 context，所以你可以用它来追踪日志
var GormDefaultMessageKey = "GORM"

// LogMode log mode
func (l *ZapLogger) LogMode(level gorm_log.LogLevel) gorm_log.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (l *ZapLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelInfo, GormDefaultMessageKey, sprintf(msg, args...))
}

// Warn print warn messages
func (l *ZapLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelWarn, GormDefaultMessageKey, sprintf(msg, args...))
}

// Error print error messages
func (l *ZapLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	_ = log.WithContext(ctx, logger).Log(log.LevelError, GormDefaultMessageKey, sprintf(msg, args...))
}

// Trace print sql message
func (l *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	fields := make(map[string]string)
	elapsed := time.Since(begin)
	fields["sql"] = sql
	fields["rows"] = fmt.Sprintf("%d", rows)
	fields["file"] = utils.FileWithLineNum()
	fields["latency"] = fmt.Sprintf("%d", elapsed)

	if err != nil {
		fields["error"] = err.Error()
		_ = log.WithContext(ctx, logger).Log(log.LevelError, "GORM", fields)
	} else {
		_ = log.WithContext(ctx, logger).Log(log.LevelInfo, "GORM", fields)
	}

}

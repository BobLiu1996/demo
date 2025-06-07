package db

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Option func(*options)

type options struct {
	source      string
	maxConn     int
	maxIdleConn int
	maxLifeTime time.Duration
	logger      logger.Interface
	logLevel    logger.LogLevel
}

func WithSource(source string) Option {
	return func(o *options) {
		o.source = source
	}
}

func WithLogger(logger logger.Interface) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithMaxConn(maxConn int) Option {
	return func(o *options) {
		if maxConn > 0 {
			o.maxConn = maxConn
		}
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return func(o *options) {
		if maxIdleConn > 0 {
			o.maxIdleConn = maxIdleConn
		}
	}
}

func WithMaxLifeTime(maxLifeTime time.Duration) Option {
	return func(o *options) {
		if maxLifeTime > 0 {
			o.maxLifeTime = maxLifeTime
		}
	}
}

func WithLogLevel(logLevel logger.LogLevel) Option {
	return func(o *options) {
		o.logLevel = logLevel
	}
}

func NewMysqlClient(opts ...Option) (*gorm.DB, error) {
	options := options{
		maxConn:     100,
		maxIdleConn: 10,
		maxLifeTime: time.Duration(300) * time.Second,
		logger:      nil,
		logLevel:    logger.Silent,
	}
	for _, o := range opts {
		o(&options)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       options.source, // DSN data source name
		DefaultStringSize:         256,            // string 类型字段的默认长度
		DisableDatetimePrecision:  true,           // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,           // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,           // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,          // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: options.logger,
	})
	if err != nil {
		return nil, err
	}
	sdb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sdb.SetMaxOpenConns(options.maxConn)
	sdb.SetMaxIdleConns(options.maxIdleConn)
	sdb.SetConnMaxLifetime(options.maxLifeTime)
	return db, nil
}

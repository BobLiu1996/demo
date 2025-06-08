package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/durationpb"
	"strings"
	"time"
)

type OptionOpt func(*rdsOptions)

type rdsOptions struct {
	user     string
	password string
	db       int
	poolSize int
	readTO   time.Duration
	writeTO  time.Duration
}

func WithPassword(password string) OptionOpt {
	return func(o *rdsOptions) {
		o.password = password
	}
}

func WithUser(user string) OptionOpt {
	return func(o *rdsOptions) {
		o.user = user
	}
}

func WithDb(db int) OptionOpt {
	return func(o *rdsOptions) {
		o.db = db
	}
}

func WithReadTO(to time.Duration) OptionOpt {
	return func(o *rdsOptions) {
		if to.Seconds() > 0 {
			o.readTO = to
		}
	}
}

func WithWriteTO(to time.Duration) OptionOpt {
	return func(o *rdsOptions) {
		if to.Seconds() > 0 {
			o.writeTO = to
		}
	}
}

func WithPoolSize(poolSize int) OptionOpt {
	return func(o *rdsOptions) {
		if poolSize > 0 {
			o.poolSize = poolSize
		}
	}
}

// RedisResource 测试环境是redis集群, 产线是redis单机，只能二选一
type RedisResource struct {
	clusterClient *redis.ClusterClient
	client        *redis.Client
}

type RedisConf struct {
	Addr         string               `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Password     string               `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	Db           int32                `protobuf:"varint,3,opt,name=db,proto3" json:"db,omitempty"`
	Pool         int32                `protobuf:"varint,4,opt,name=pool,proto3" json:"pool,omitempty"`
	ReadTimeout  *durationpb.Duration `protobuf:"bytes,5,opt,name=read_timeout,json=readTimeout,proto3" json:"read_timeout,omitempty"`
	WriteTimeout *durationpb.Duration `protobuf:"bytes,6,opt,name=write_timeout,json=writeTimeout,proto3" json:"write_timeout,omitempty"`
	IsCluster    bool                 `protobuf:"varint,7,opt,name=is_cluster,json=isCluster,proto3" json:"is_cluster,omitempty"`
	User         string               `protobuf:"bytes,8,opt,name=user,proto3" json:"user,omitempty"`
}

func NewRedisResource(redisConf *RedisConf) (*RedisResource, error) {
	var rdb *RedisResource
	var err error
	if redisConf.IsCluster {
		rdb, err = newRedisResourceWithClusterClient(
			redisConf.Addr,
			WithUser(redisConf.User),
			WithPassword(redisConf.Password),
			WithPoolSize(int(redisConf.Pool)),
			WithReadTO(redisConf.ReadTimeout.AsDuration()),
			WithWriteTO(redisConf.WriteTimeout.AsDuration()),
		)
		if err != nil {
			return nil, err
		}
	} else {
		rdb, err = newRedisResourceWithSingleClient(
			redisConf.Addr,
			WithUser(redisConf.User),
			WithPassword(redisConf.Password),
			WithPoolSize(int(redisConf.Pool)),
			WithDb(int(redisConf.Db)),
			WithReadTO(redisConf.ReadTimeout.AsDuration()),
			WithWriteTO(redisConf.WriteTimeout.AsDuration()),
		)
		if err != nil {
			return nil, err
		}
	}
	return rdb, nil
}

func newRedisResourceWithClusterClient(addrs string, opts ...OptionOpt) (*RedisResource, error) {
	options := rdsOptions{
		poolSize: 100,
	}
	for _, o := range opts {
		o(&options)
	}
	clusterClientOpt := &redis.ClusterOptions{
		PoolSize: options.poolSize,
		Addrs:    strings.Split(addrs, ","),
		Password: options.password,
		Username: options.user,
	}
	if options.readTO.Seconds() > 0 {
		clusterClientOpt.ReadTimeout = options.readTO
	}
	if options.writeTO.Seconds() > 0 {
		clusterClientOpt.WriteTimeout = options.writeTO
	}
	cli := redis.NewClusterClient(clusterClientOpt)
	_, err := cli.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	r := &RedisResource{
		clusterClient: cli,
	}
	return r, nil
}

func newRedisResourceWithSingleClient(addr string, opts ...OptionOpt) (*RedisResource, error) {
	options := rdsOptions{
		poolSize: 100,
	}
	for _, o := range opts {
		o(&options)
	}
	clientOpt := &redis.Options{
		PoolSize: options.poolSize,
		Addr:     addr,
		Password: options.password,
		DB:       options.db,
		Username: options.user,
	}
	if options.readTO.Seconds() > 0 {
		clientOpt.ReadTimeout = options.readTO
	}
	if options.writeTO.Seconds() > 0 {
		clientOpt.WriteTimeout = options.writeTO
	}
	cli := redis.NewClient(clientOpt)
	_, err := cli.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	r := &RedisResource{
		client: cli,
	}
	return r, nil
}

func (r *RedisResource) CloseRedisClient() error {
	if r.client != nil {
		return r.client.Close()
	}
	if r.clusterClient != nil {
		return r.clusterClient.Close()
	}
	return nil
}

func (r *RedisResource) RedisClient() redis.Cmdable {
	if r.clusterClient != nil {
		return r.clusterClient
	} else {
		return r.client
	}
}

func (r *RedisResource) RedisUniversalClient() redis.UniversalClient {
	if r.clusterClient != nil {
		return r.clusterClient
	} else {
		return r.client
	}
}

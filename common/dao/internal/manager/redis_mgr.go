package manager

import (
	"context"
	"poker_server/common/dao/internal/redis"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/library/uerror"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

var (
	redisPool = make(map[string]*redis.RedisClient)
)

func InitRedis(cfgs map[int32]*yaml.DbConfig) error {
	if len(cfgs) <= 0 {
		return uerror.New(1, pb.ErrorCode_CONFIG_NOT_FOUND, "redis配置为空")
	}
	for _, cfg := range cfgs {
		// 建立redis连接
		cli := goredis.NewClient(&goredis.Options{
			IdleTimeout:  1 * time.Minute,
			MinIdleConns: 100,
			DB:           cfg.Db,
			ReadTimeout:  -1,
			WriteTimeout: -1,
			Addr:         cfg.Host,
			Username:     cfg.User,
			Password:     cfg.Password,
			OnConnect:    func(ctx context.Context, cn *goredis.Conn) error { return nil },
		})
		// 连接到redis服务器，测试连通性
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := cli.Ping(ctx).Result(); err != nil {
			return uerror.New(1, pb.ErrorCode_PING_FAILED, "ping测试失败：%v", err)
		}
		redisPool[cfg.DbName] = redis.NewRedisClient(cli, cfg.DbName)
	}
	return nil
}

func GetRedis(dbid string) *redis.RedisClient {
	return redisPool[dbid]
}

package dao

import (
	"poker_server/common/dao/internal/manager"
	mredis "poker_server/common/dao/internal/redis"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/library/uerror"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis(cfgs map[int32]*yaml.DbConfig) error {
	return manager.InitRedis(cfgs)
}

func GetRedisClient(db string) *mredis.RedisClient {
	return manager.GetRedis(db)
}

func IncrBy(dbid string, key string, val int64) (ret int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.IncrBy(key, val)
	return
}

func Get(dbid string, key string) (str string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	str, err = cli.Get(key)
	return
}

func Set(dbid string, key string, val interface{}) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.Set(key, val)
	return
}

// 不存在key时，设置该key的值未val
func SetNX(dbid string, key string, val interface{}) (exist bool, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	exist, err = cli.SetNX(key, val)
	return
}

// 设置key的过期时间
func SetEX(dbid string, key string, val interface{}, ttl time.Duration) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.SetEX(key, val, ttl)
	return
}

// 批量读取操作
func MGet(dbid string, keys ...string) (rets []interface{}, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	args := []string{}
	for i := 0; i < len(keys); i++ {
		args = append(args, (keys[i]))
	}
	rets, err = cli.MGet(args...)
	return
}

// 批量设置键值对
func MSet(dbid string, args ...interface{}) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.MSet(args...)
	return
}

func HGetAll(dbid string, key string) (ret map[string]string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.HGetAll(key)
	return
}

func HGet(dbid string, key string, field string) (ret string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.HGet(key, field)
	return
}

func HDel(dbid string, key string, fields ...string) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.HDel(key, fields...)
	return
}

func HKeys(dbid string, key string) (rets []string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	rets, err = cli.HKeys(key)
	return
}

func HIncrBy(dbid string, key string, field string, incr int64) (int64, error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		return 0, uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
	}
	return cli.HIncrBy(key, field, incr)
}

func HSet(dbid string, key string, field string, val interface{}) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.HSet(key, field, val)
	return
}

func HMSet(dbid string, key string, vals ...interface{}) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.HMSet(key, vals...)
	return
}

func ZAdd(dbid string, key string, members ...*redis.Z) (err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	err = cli.ZAdd(key, members...)
	return
}

func ZCard(dbid string, key string) (count int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	count, err = cli.ZCard(key)
	return
}

// 返回有序集合中指定成员的排名，有序集合成员按分数值递减排序
func ZRevRank(dbid string, key string, member string) (count int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	count, err = cli.ZRevRank(key, member)
	return
}

// 返回有序集合中指定区间内的成员
func ZRevRange(dbid string, key string, start, stop int64) (members []string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	members, err = cli.ZRevRange(key, start, stop)
	return
}

func ZScore(dbid string, key string, member string) (score float64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	score, err = cli.ZScore(key, member)
	return
}

func ZRevRangeWithScores(dbid string, key string, start, stop int64) (members []redis.Z, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	members, err = cli.ZRevRangeWithScores(key, start, stop)
	return
}

func RPush(dbid string, key string, values ...interface{}) (ret int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.RPush(key, values...)
	return
}

func RPop(dbid string, key string) (ret string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.RPop(key)
	return
}

func LLen(dbid string, key string) (ret int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.LLen(key)
	return
}

func LRange(dbid string, key string, start, stop int64) (ret []string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.LRange(key, start, stop)
	return
}

func LTrim(dbid string, key string, start, stop int64) (ret string, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.LTrim(key, start, stop)
	return
}

func LRem(dbid string, key string, val interface{}) (ret int64, err error) {
	cli := manager.GetRedis(dbid)
	if cli == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "dbid:%s", dbid)
		return
	}
	ret, err = cli.LRem(key, val)
	return
}

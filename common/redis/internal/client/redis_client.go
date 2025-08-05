package client

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *goredis.Client
	preKey string
}

func NewRedisClient(cli *goredis.Client, key string) *RedisClient {
	return &RedisClient{client: cli, preKey: key}
}

func (d *RedisClient) GetClient() *goredis.Client {
	return d.client
}

func (d *RedisClient) getKey(key string) string {
	if len(d.preKey) > 0 {
		return fmt.Sprintf("%s_%s", d.preKey, key)
	}
	return key
}

func (d *RedisClient) Del(key string) error {
	_, err := d.client.Del(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return err
}

func (d *RedisClient) Incr(key string) (ret int64, err error) {
	ret, err = d.client.Incr(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) IncrBy(key string, val int64) (ret int64, err error) {
	ret, err = d.client.IncrBy(context.Background(), d.getKey(key), val).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) Get(key string) (str string, err error) {
	str, err = d.client.Get(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) Set(key string, val interface{}) (err error) {
	_, err = d.client.Set(context.Background(), d.getKey(key), val, 0).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 不存在key时，设置该key的值未val
func (d *RedisClient) SetNX(key string, val interface{}) (exist bool, err error) {
	exist, err = d.client.SetNX(context.Background(), d.getKey(key), val, 0).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 设置key的过期时间
func (d *RedisClient) SetEX(key string, val interface{}, ttl time.Duration) (err error) {
	_, err = d.client.SetEX(context.Background(), d.getKey(key), val, ttl).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 批量读取操作
func (d *RedisClient) MGet(keys ...string) (rets []interface{}, err error) {
	args := []string{}
	for i := 0; i < len(keys); i++ {
		args = append(args, d.getKey(keys[i]))
	}
	rets, err = d.client.MGet(context.Background(), args...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 批量设置键值对
func (d *RedisClient) MSet(args ...interface{}) (err error) {
	_, err = d.client.MSet(context.Background(), args...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HGetAll(key string) (ret map[string]string, err error) {
	ret, err = d.client.HGetAll(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HGet(key string, field string) (ret string, err error) {
	ret, err = d.client.HGet(context.Background(), d.getKey(key), field).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HDel(key string, fields ...string) (err error) {
	_, err = d.client.HDel(context.Background(), d.getKey(key), fields...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HKeys(key string) (rets []string, err error) {
	rets, err = d.client.HKeys(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HIncrBy(key string, field string, incr int64) (newIncr int64, err error) {
	newIncr, err = d.client.HIncrBy(context.Background(), d.getKey(key), field, incr).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HSet(key string, field string, val interface{}) (err error) {
	_, err = d.client.HSet(context.Background(), d.getKey(key), field, val).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	rets, err := d.client.HMGet(context.Background(), d.getKey(key), fields...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return rets, err
}

func (d *RedisClient) HMSet(key string, vals ...interface{}) (err error) {
	_, err = d.client.HMSet(context.Background(), d.getKey(key), vals...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) ZAdd(key string, members ...*goredis.Z) (err error) {
	_, err = d.client.ZAdd(context.Background(), d.getKey(key), members...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) ZCard(key string) (count int64, err error) {
	count, err = d.client.ZCard(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 返回有序集合中指定成员的排名，有序集合成员按分数值递减排序
func (d *RedisClient) ZRevRank(key string, member string) (count int64, err error) {
	count, err = d.client.ZRevRank(context.Background(), d.getKey(key), member).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

// 返回有序集合中指定区间内的成员
func (d *RedisClient) ZRevRange(key string, start, stop int64) (members []string, err error) {
	members, err = d.client.ZRevRange(context.Background(), d.getKey(key), start, stop).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) ZScore(key string, member string) (score float64, err error) {
	score, err = d.client.ZScore(context.Background(), d.getKey(key), member).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) ZRevRangeWithScores(key string, start, stop int64) (members []goredis.Z, err error) {
	members, err = d.client.ZRevRangeWithScores(context.Background(), d.getKey(key), start, stop).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) RPush(key string, values ...interface{}) (ret int64, err error) {
	ret, err = d.client.RPush(context.Background(), d.getKey(key), values...).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) RPop(key string) (ret string, err error) {
	ret, err = d.client.RPop(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) LLen(key string) (ret int64, err error) {
	ret, err = d.client.LLen(context.Background(), d.getKey(key)).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) LRange(key string, start, stop int64) (ret []string, err error) {
	ret, err = d.client.LRange(context.Background(), d.getKey(key), start, stop).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) LTrim(key string, start, stop int64) (ret string, err error) {
	ret, err = d.client.LTrim(context.Background(), d.getKey(key), start, stop).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

func (d *RedisClient) LRem(key string, val interface{}) (ret int64, err error) {
	ret, err = d.client.LRem(context.Background(), d.getKey(key), 0, val).Result()
	if err == goredis.Nil {
		err = nil
	}
	return
}

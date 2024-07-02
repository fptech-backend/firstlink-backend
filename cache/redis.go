package cache

import (
	"certification/logger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	RDB *redis.Client
}

var Redis Cache

// ----------------- Redis -----------------

func (c *Cache) ConnectRedis(host, port, password string, db int) error {
	c.RDB = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	_, err := c.RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) CloseRedis() {
	err := c.RDB.Close()
	if err != nil {
		logger.Log.Error(err)
	}
}

func _setCache(rdb *redis.Client, key string, dataJson interface{}, expiration time.Duration) error {
	_data, err := json.Marshal(dataJson)
	if err != nil {
		return err
	}
	return rdb.Set(context.Background(), key, _data, expiration).Err()
}

func (c *Cache) SetCache(key string, dataJson interface{}, expiration time.Duration) error {
	_key := fmt.Sprintf("%s:-", key)
	return _setCache(c.RDB, _key, dataJson, expiration)
}

func (c *Cache) SetCacheById(key, id string, dataJson interface{}, expiration time.Duration) error {
	_key := fmt.Sprintf("%s:%s", key, id)
	return _setCache(c.RDB, _key, dataJson, expiration)
}

func (c *Cache) SetCacheByIdForId(key, id, user_id string, dataJson interface{}, expiration time.Duration) error {
	_key := fmt.Sprintf("%s:%s:%s", key, id, user_id)
	return _setCache(c.RDB, _key, dataJson, expiration)
}

func _getCache(rdb *redis.Client, key string) (map[string]interface{}, error) {
	results, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(results), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Cache) GetCache(key string) (interface{}, error) {
	_key := fmt.Sprintf("%s:-", key)
	return _getCache(c.RDB, _key)
}

func (c *Cache) GetCacheById(key, id string) (map[string]interface{}, error) {
	_key := fmt.Sprintf("%s:%s", key, id)
	return _getCache(c.RDB, _key)
}

func (c *Cache) GetCacheByIdForId(key, id, user_id string) (map[string]interface{}, error) {
	_key := fmt.Sprintf("%s:%s:%s", key, id, user_id)
	return _getCache(c.RDB, _key)
}

func (c *Cache) DeleteCache(key string) error {
	_key := fmt.Sprintf("%s:-", key)
	return c.RDB.Del(context.Background(), _key).Err()
}

func (c *Cache) DeleteCacheById(key, id string) error {
	_key := fmt.Sprintf("%s:%s", key, id)
	return c.RDB.Del(context.Background(), _key).Err()
}

func (c *Cache) DeleteCacheByIdForId(key, id, user_id string) error {
	_key := fmt.Sprintf("%s:%s:%s", key, id, user_id)
	return c.RDB.Del(context.Background(), _key).Err()
}

func (c *Cache) DeleteCacheByUserId(key, user_id string) error {
	_key := fmt.Sprintf("%s:%s", key, user_id)
	return c.RDB.Del(context.Background(), _key).Err()
}

func (c *Cache) DeleteCacheGroup(key string) error {
	_key := fmt.Sprintf("%s:*", key)
	keys, err := c.RDB.Keys(context.Background(), _key).Result()
	if err != nil {
		return err
	}
	return c.RDB.Del(context.Background(), keys...).Err()
}

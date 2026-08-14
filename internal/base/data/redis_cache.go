/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package data

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentfault/pacman/cache"
)

const (
	redisPingTimeout = 5 * time.Second
	redisScanSize    = int64(500)
)

var _ cache.Cache = (*RedisCache)(nil)
var _ AtomicCache = (*RedisCache)(nil)

var changeInt64Script = redis.NewScript(`
  if redis.call("EXISTS", KEYS[1]) == 0 then
	return redis.error_reply("cache key does not exist")
  end
  return redis.call("INCRBY", KEYS[1], ARGV[1])
  `)

var slidingWindowScript = redis.NewScript(`
  local now = tonumber(ARGV[1])
  local member = ARGV[2]
  local max_retry_after = 0

  for index, key in ipairs(KEYS) do
    local limit = tonumber(ARGV[2 + index * 2 - 1])
    local window = tonumber(ARGV[2 + index * 2])
    redis.call("ZREMRANGEBYSCORE", key, "-inf", now - window)
    local count = redis.call("ZCARD", key)
    if count >= limit then
      local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
      if #oldest > 0 then
        local retry_after = tonumber(oldest[2]) + window - now
        if retry_after > max_retry_after then
          max_retry_after = retry_after
        end
      end
    end
  end

  if max_retry_after > 0 then
    return max_retry_after
  end

  for index, key in ipairs(KEYS) do
    local window = tonumber(ARGV[2 + index * 2])
    redis.call("ZADD", key, now, member)
    redis.call("PEXPIRE", key, window)
  end
  return 0
  `)

var compareAndDeleteScript = redis.NewScript(`
  local value = redis.call("GET", KEYS[1])
  if value and value == ARGV[1] then
    redis.call("DEL", KEYS[1])
    return 1
  end
  return 0
  `)

type RedisCache struct {
	client    *redis.Client
	keyPrefix string
}

func NewRedisCache(conf RedisCacheConf) (*RedisCache, error) {
	host := strings.TrimSpace(conf.Host)
	keyPrefix := strings.TrimSpace(conf.KeyPrefix)

	if host == "" {
		return nil, errors.New("redis cache host is required")
	}
	if conf.Port < 1 || conf.Port > 65535 {
		return nil, fmt.Errorf("redis cache port must be between 1 and 65535, got %d", conf.Port)
	}
	if conf.DB < 0 {
		return nil, fmt.Errorf("redis cache db must be non-negative, got %d", conf.DB)
	}
	if conf.PoolSize < 0 {
		return nil, fmt.Errorf("redis cache pool size must be non-negative, got %d", conf.PoolSize)
	}
	if err := validateRedisKeyPrefix(keyPrefix); err != nil {
		return nil, err
	}

	poolSize := conf.PoolSize
	if poolSize <= 0 {
		poolSize = DefaultRedisPoolSize
	}
	options := &redis.Options{
		Addr:         net.JoinHostPort(host, strconv.Itoa(conf.Port)),
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           conf.DB,
		PoolSize:     poolSize,
		DialTimeout:  redisPingTimeout,
		ReadTimeout:  redisPingTimeout,
		WriteTimeout: redisPingTimeout,
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to redis at %s: %w", options.Addr, err)
	}
	return &RedisCache{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

// validateRedisKeyPrefix prevents one deployment from clearing another
// deployment's keys when Flush scans the shared Redis database.
func validateRedisKeyPrefix(prefix string) error {
	if prefix == "" {
		return errors.New("redis cache key prefix is required")
	}
	if !strings.HasSuffix(prefix, ":") {
		return errors.New("redis cache key prefix must end with ':'")
	}
	if strings.ContainsAny(prefix, "*?[]\\") {
		return errors.New("redis cache key prefix contains unsupported pattern characters")
	}
	return nil
}

func (c *RedisCache) key(key string) string {
	return c.keyPrefix + key
}

// GetString returns a cached string.
func (c *RedisCache) GetString(ctx context.Context, key string) (string, bool, error) {
	value, err := c.client.Get(ctx, c.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// SetString stores a string with an optional TTL. A zero TTL does not expire,
// matching the existing memory cache behavior.
func (c *RedisCache) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

// SetIfAbsent stores a value only when the key does not already exist.
func (c *RedisCache) SetIfAbsent(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.key(key), value, ttl).Result()
}

// GetInt64 returns a cached int64.
func (c *RedisCache) GetInt64(ctx context.Context, key string) (int64, bool, error) {
	value, exist, err := c.GetString(ctx, key)
	if err != nil || !exist {
		return 0, exist, err
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, true, fmt.Errorf("parse cached int64: %w", err)
	}

	return result, true, nil
}

// SetInt64 stores an int64 with an optional TTL.
func (c *RedisCache) SetInt64(ctx context.Context, key string, value int64, ttl time.Duration) error {
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

// Increase atomically increments an existing int64 and preserves its TTL.
func (c *RedisCache) Increase(ctx context.Context, key string, value int64) (int64, error) {
	return c.changeExistingInt64(ctx, key, value)
}

// Decrease atomically decrements an existing int64 and preserves its TTL.
func (c *RedisCache) Decrease(ctx context.Context, key string, value int64) (int64, error) {
	return c.changeExistingInt64(ctx, key, -value)
}

func (c *RedisCache) changeExistingInt64(ctx context.Context, key string, delta int64) (int64, error) {
	result, err := changeInt64Script.Run(ctx, c.client, []string{c.key(key)}, delta).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Del deletes one cache key.
func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

// CheckAndRecordSlidingWindows checks all rules and records the request only
// when every rule allows it. The Lua script keeps this atomic across instances.
func (c *RedisCache) CheckAndRecordSlidingWindows(
	ctx context.Context,
	member string,
	rules []SlidingWindowRule,
) (time.Duration, error) {
	if len(rules) == 0 {
		return 0, nil
	}
	if member == "" {
		return 0, errors.New("sliding window member is required")
	}

	keys := make([]string, 0, len(rules))
	args := make([]any, 0, 2+len(rules)*2)
	args = append(args, time.Now().UnixMilli(), member)
	for _, rule := range rules {
		if rule.Key == "" || rule.Limit <= 0 || rule.Window <= 0 {
			return 0, errors.New("invalid sliding window rule")
		}
		keys = append(keys, c.key(rule.Key))
		args = append(args, rule.Limit, rule.Window.Milliseconds())
	}

	retryAfterMilliseconds, err := slidingWindowScript.Run(ctx, c.client, keys, args...).Int64()
	if err != nil {
		return 0, err
	}
	return time.Duration(retryAfterMilliseconds) * time.Millisecond, nil
}

// CompareAndDelete consumes a value only when it matches the expected value.
func (c *RedisCache) CompareAndDelete(ctx context.Context, key, expected string) (bool, error) {
	result, err := compareAndDeleteScript.Run(ctx, c.client, []string{c.key(key)}, expected).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// Flush deletes only keys in this deployment's namespace. It never issues
// FLUSHDB or FLUSHALL because the Redis database may be shared.
func (c *RedisCache) Flush(ctx context.Context) error {
	var cursor uint64
	pattern := c.keyPrefix + "*"

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, redisScanSize).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			return nil
		}
	}
}

// Close releases the Redis connection pool.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

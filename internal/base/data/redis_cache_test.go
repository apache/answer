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
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func newTestRedisCache(
	t *testing.T,
) (*RedisCache, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)

	redisCache, err := NewRedisCache(RedisCacheConf{
		Host:      server.Host(),
		Port:      mustRedisPort(t, server.Port()),
		KeyPrefix: "hnu-forum:test:",
		PoolSize:  2,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, redisCache.Close())
	})

	return redisCache, server
}

func mustRedisPort(t *testing.T, port string) int {
	t.Helper()
	value, err := strconv.Atoi(port)
	require.NoError(t, err)
	return value
}

func TestRedisCacheString(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)
	ctx := context.Background()

	value, exist, err := redisCache.GetString(ctx, "missing")
	require.NoError(t, err)
	require.False(t, exist)
	require.Empty(t, value)

	require.NoError(
		t,
		redisCache.SetString(ctx, "name", "answer", time.Minute),
	)

	value, exist, err = redisCache.GetString(ctx, "name")
	require.NoError(t, err)
	require.True(t, exist)
	require.Equal(t, "answer", value)

	require.NoError(t, redisCache.Del(ctx, "name"))

	_, exist, err = redisCache.GetString(ctx, "name")
	require.NoError(t, err)
	require.False(t, exist)
}

func TestRedisCacheTTL(t *testing.T) {
	redisCache, server := newTestRedisCache(t)
	ctx := context.Background()

	require.NoError(
		t,
		redisCache.SetString(ctx, "temporary", "value", time.Minute),
	)

	server.FastForward(time.Minute + time.Second)

	_, exist, err := redisCache.GetString(ctx, "temporary")
	require.NoError(t, err)
	require.False(t, exist)
}

func TestRedisCacheSetIfAbsent(t *testing.T) {
	redisCache, server := newTestRedisCache(t)
	ctx := context.Background()

	stored, err := redisCache.SetIfAbsent(ctx, "lock", "first", time.Minute)
	require.NoError(t, err)
	require.True(t, stored)

	stored, err = redisCache.SetIfAbsent(ctx, "lock", "second", time.Minute)
	require.NoError(t, err)
	require.False(t, stored)

	value, exists, err := redisCache.GetString(ctx, "lock")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "first", value)

	server.FastForward(time.Minute + time.Second)
	stored, err = redisCache.SetIfAbsent(ctx, "lock", "third", time.Minute)
	require.NoError(t, err)
	require.True(t, stored)
}

func TestRedisCacheAuthentication(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("test-password")

	redisCache, err := NewRedisCache(RedisCacheConf{
		Host:      server.Host(),
		Port:      mustRedisPort(t, server.Port()),
		Password:  "test-password",
		KeyPrefix: "hnu-forum:test:",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, redisCache.Close())
	})

	require.NoError(t, redisCache.SetString(context.Background(), "authenticated", "yes", time.Minute))
}

func TestRedisCacheInt64(t *testing.T) {
	redisCache, server := newTestRedisCache(t)
	ctx := context.Background()

	require.NoError(
		t,
		redisCache.SetInt64(ctx, "counter", 10, time.Minute),
	)

	ttlBefore := server.TTL(redisCache.key("counter"))

	value, err := redisCache.Increase(ctx, "counter", 5)
	require.NoError(t, err)
	require.Equal(t, int64(15), value)

	value, err = redisCache.Decrease(ctx, "counter", 3)
	require.NoError(t, err)
	require.Equal(t, int64(12), value)

	storedValue, exist, err := redisCache.GetInt64(ctx, "counter")
	require.NoError(t, err)
	require.True(t, exist)
	require.Equal(t, int64(12), storedValue)

	ttlAfter := server.TTL(redisCache.key("counter"))
	require.Equal(t, ttlBefore, ttlAfter)
}

func TestRedisCacheIncreaseMissingKey(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)

	_, err := redisCache.Increase(
		context.Background(),
		"missing-counter",
		1,
	)
	require.Error(t, err)
}

func TestRedisCacheSlidingWindows(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)
	ctx := context.Background()
	rules := []SlidingWindowRule{
		{Key: "rate:email:minute", Limit: 1, Window: time.Minute},
		{Key: "rate:email:hour", Limit: 5, Window: time.Hour},
		{Key: "rate:ip:hour", Limit: 100, Window: time.Hour},
	}

	retryAfter, err := redisCache.CheckAndRecordSlidingWindows(ctx, "request-1", rules)
	require.NoError(t, err)
	require.Zero(t, retryAfter)

	retryAfter, err = redisCache.CheckAndRecordSlidingWindows(ctx, "request-2", rules)
	require.NoError(t, err)
	require.Greater(t, retryAfter, time.Duration(0))
	require.LessOrEqual(t, retryAfter, time.Minute)
}

func TestRedisCacheSlidingWindowsDoNotPartiallyRecordRejectedRequest(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)
	ctx := context.Background()

	_, err := redisCache.CheckAndRecordSlidingWindows(ctx, "request-1", []SlidingWindowRule{
		{Key: "rate:strict", Limit: 1, Window: time.Hour},
		{Key: "rate:loose", Limit: 10, Window: time.Hour},
	})
	require.NoError(t, err)

	retryAfter, err := redisCache.CheckAndRecordSlidingWindows(ctx, "request-2", []SlidingWindowRule{
		{Key: "rate:strict", Limit: 1, Window: time.Hour},
		{Key: "rate:loose", Limit: 10, Window: time.Hour},
	})
	require.NoError(t, err)
	require.Greater(t, retryAfter, time.Duration(0))
	require.Equal(t, int64(1), redisCache.client.ZCard(ctx, redisCache.key("rate:loose")).Val())
}

func TestRedisCacheCompareAndDelete(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)
	ctx := context.Background()
	require.NoError(t, redisCache.SetString(ctx, "one-time-code", "digest", time.Minute))

	matched, err := redisCache.CompareAndDelete(ctx, "one-time-code", "wrong")
	require.NoError(t, err)
	require.False(t, matched)

	matched, err = redisCache.CompareAndDelete(ctx, "one-time-code", "digest")
	require.NoError(t, err)
	require.True(t, matched)

	matched, err = redisCache.CompareAndDelete(ctx, "one-time-code", "digest")
	require.NoError(t, err)
	require.False(t, matched)
}

func TestRedisCacheFlushOnlyOwnNamespace(t *testing.T) {
	redisCache, _ := newTestRedisCache(t)
	ctx := context.Background()

	require.NoError(
		t,
		redisCache.SetString(ctx, "owned", "value", time.Hour),
	)
	require.NoError(
		t,
		redisCache.client.Set(
			ctx,
			"another-project:key",
			"value",
			time.Hour,
		).Err(),
	)

	require.NoError(t, redisCache.Flush(ctx))

	_, exist, err := redisCache.GetString(ctx, "owned")
	require.NoError(t, err)
	require.False(t, exist)

	externalValue, err := redisCache.client.Get(
		ctx,
		"another-project:key",
	).Result()
	require.NoError(t, err)
	require.Equal(t, "value", externalValue)
}

func TestRedisCacheConfigValidation(t *testing.T) {
	testCases := []struct {
		name string
		conf RedisCacheConf
	}{
		{
			name: "missing host",
			conf: RedisCacheConf{
				Port:      6379,
				KeyPrefix: "hnu-forum:test:",
			},
		},
		{
			name: "invalid port",
			conf: RedisCacheConf{
				Host:      "127.0.0.1",
				Port:      70000,
				KeyPrefix: "hnu-forum:test:",
			},
		},
		{
			name: "invalid database",
			conf: RedisCacheConf{
				Host:      "127.0.0.1",
				Port:      6379,
				DB:        -1,
				KeyPrefix: "hnu-forum:test:",
			},
		},
		{
			name: "invalid pool size",
			conf: RedisCacheConf{
				Host:      "127.0.0.1",
				Port:      6379,
				KeyPrefix: "hnu-forum:test:",
				PoolSize:  -1,
			},
		},
		{
			name: "missing prefix",
			conf: RedisCacheConf{
				Host: "127.0.0.1",
				Port: 6379,
			},
		},
		{
			name: "prefix without separator",
			conf: RedisCacheConf{
				Host:      "127.0.0.1",
				Port:      6379,
				KeyPrefix: "hnu-forum",
			},
		},
		{
			name: "prefix with pattern",
			conf: RedisCacheConf{
				Host:      "127.0.0.1",
				Port:      6379,
				KeyPrefix: "hnu-forum:*:",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewRedisCache(testCase.conf)
			require.Error(t, err)
		})
	}
}

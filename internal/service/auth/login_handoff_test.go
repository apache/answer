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

package auth

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/segmentfault/pacman/cache"
)

type atomicTestCache struct {
	cache.Cache
	consumeCalled bool
}

func (c *atomicTestCache) ConsumeString(ctx context.Context, key string) (string, bool, error) {
	c.consumeCalled = true
	value, exists, err := c.GetString(ctx, key)
	if err != nil || !exists {
		return value, exists, err
	}
	return value, true, c.Del(ctx, key)
}

func TestLoginHandoffServiceConsumesCodeOnlyOnce(t *testing.T) {
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	service := NewLoginHandoffService(&data.Data{Cache: cache})
	const accessToken = "long-lived-session-token"
	code, err := service.Create(context.Background(), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	if code == "" || strings.Contains(code, accessToken) {
		t.Fatalf("handoff code must be opaque: %q", code)
	}

	got, exists, err := service.Consume(context.Background(), code)
	if err != nil || !exists || got != accessToken {
		t.Fatalf("first consume = (%q, %v, %v), want (%q, true, nil)", got, exists, err, accessToken)
	}

	got, exists, err = service.Consume(context.Background(), code)
	if err != nil || exists || got != "" {
		t.Fatalf("second consume = (%q, %v, %v), want (empty, false, nil)", got, exists, err)
	}
}

func TestLoginHandoffServiceConcurrentConsumeReturnsTokenOnce(t *testing.T) {
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	service := NewLoginHandoffService(&data.Data{Cache: cache})
	code, err := service.Create(context.Background(), "session-token")
	if err != nil {
		t.Fatal(err)
	}

	const consumers = 16
	var wg sync.WaitGroup
	results := make(chan bool, consumers)
	for range consumers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, exists, err := service.Consume(context.Background(), code)
			if err != nil {
				t.Errorf("consume: %v", err)
				return
			}
			results <- exists
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	for exists := range results {
		if exists {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful consumes = %d, want 1", successes)
	}
}

func TestLoginHandoffServiceUsesAtomicCacheConsumer(t *testing.T) {
	memoryCache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	atomicCache := &atomicTestCache{Cache: memoryCache}
	service := NewLoginHandoffService(&data.Data{Cache: atomicCache})
	code, err := service.Create(context.Background(), "session-token")
	if err != nil {
		t.Fatal(err)
	}
	_, exists, err := service.Consume(context.Background(), code)
	if err != nil || !exists {
		t.Fatalf("consume = (_, %v, %v), want (_, true, nil)", exists, err)
	}
	if !atomicCache.consumeCalled {
		t.Fatal("atomic cache consumer was not used")
	}
}

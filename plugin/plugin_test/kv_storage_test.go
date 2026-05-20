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

package plugin_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
	_ "modernc.org/sqlite"
)

var (
	testPlugin *TestKVStoragePlugin
)

// Helper functions for testing
func mustSet(t *testing.T, kv *plugin.KVOperator, ctx context.Context, group, key, value string) {
	if err := kv.Set(ctx, plugin.KVParams{Group: group, Key: key, Value: value}); err != nil {
		t.Fatalf("Failed to set %s/%s: %v", group, key, err)
	}
}

func mustGet(t *testing.T, kv *plugin.KVOperator, ctx context.Context, group, key, expected string) {
	val, err := kv.Get(ctx, plugin.KVParams{Group: group, Key: key})
	if err != nil {
		t.Fatalf("Failed to get %s/%s: %v", group, key, err)
	}
	if val != expected {
		t.Errorf("Expected '%s' for %s/%s, got '%s'", expected, group, key, val)
	}
}

func mustDel(t *testing.T, kv *plugin.KVOperator, ctx context.Context, group, key string) {
	if err := kv.Del(ctx, plugin.KVParams{Group: group, Key: key}); err != nil {
		t.Fatalf("Failed to delete %s/%s: %v", group, key, err)
	}
}

func assertNotFound(t *testing.T, kv *plugin.KVOperator, ctx context.Context, group, key string) {
	val, err := kv.Get(ctx, plugin.KVParams{Group: group, Key: key})
	if err != plugin.ErrKVKeyNotFound {
		t.Errorf("Expected ErrKVKeyNotFound for %s/%s, got: %v", group, key, err)
	}
	if val != "" {
		t.Errorf("Expected empty value for %s/%s, got: '%s'", group, key, val)
	}
}

func assertError(t *testing.T, err error, expected error, msg string) {
	if err != expected {
		t.Errorf("%s: expected %v, got %v", msg, expected, err)
	}
}

// TestKVStoragePlugin implements KVStorage interface for testing
type TestKVStoragePlugin struct {
	operator *plugin.KVOperator
}

// Info returns plugin information
func (p *TestKVStoragePlugin) Info() plugin.Info {
	return plugin.Info{
		Name:        plugin.MakeTranslator("test_kv_storage_name"),
		SlugName:    "test_kv_storage",
		Description: plugin.MakeTranslator("test_kv_storage_desc"),
		Author:      "Answer Team",
		Version:     "1.0.0",
		Link:        "https://github.com/apache/answer",
	}
}

// SetOperator sets KV operator
func (p *TestKVStoragePlugin) SetOperator(operator *plugin.KVOperator) {
	p.operator = operator
}

// setupTestEnvironment sets up test environment
func setupTestEnvironment() {
	// Initialize only once
	if testPlugin != nil {
		return
	}

	// Create and register test plugin
	testPlugin = &TestKVStoragePlugin{}
	plugin.Register(testPlugin)

	// Enable plugin
	plugin.StatusManager.Enable("test_kv_storage", true)

	// Initialize plugin data, refer to plugin_common_service.go implementation
	_ = plugin.CallKVStorage(func(k plugin.KVStorage) error {
		k.SetOperator(plugin.NewKVOperator(
			testDataSource.DB,
			testDataSource.Cache,
			k.Info().SlugName,
		))
		return nil
	})
}

// Test basic operations including CRUD and edge cases
func TestBasicOperations(t *testing.T) {
	setupTestEnvironment()
	kv := testPlugin.operator
	ctx := context.Background()

	t.Run("BasicCRUD", func(t *testing.T) {
		// Set/Get
		mustSet(t, kv, ctx, "group1", "key1", "value1")
		mustGet(t, kv, ctx, "group1", "key1", "value1")

		// Update
		mustSet(t, kv, ctx, "group1", "key1", "new_value")
		mustGet(t, kv, ctx, "group1", "key1", "new_value")

		// Delete
		mustDel(t, kv, ctx, "group1", "key1")
		assertNotFound(t, kv, ctx, "group1", "key1")

		// Group operation
		mustSet(t, kv, ctx, "group1", "key2", "value2")
		mustSet(t, kv, ctx, "group1", "key3", "value3")
		groupData, err := kv.GetByGroup(ctx, plugin.KVParams{Group: "group1", Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("Failed to get group data: %v", err)
		}

		// the groupData should only have key2 and key3 because key1 is deleted
		if len(groupData) != 2 {
			t.Errorf("Expected 2 items, got %d", len(groupData))
		}
		if groupData["key2"] != "value2" || groupData["key3"] != "value3" {
			t.Errorf("Unexpected group data: %v", groupData)
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Empty key
		err := kv.Set(ctx, plugin.KVParams{Group: "group", Key: "", Value: "value"})
		assertError(t, err, plugin.ErrKVKeyEmpty, "Empty key test")

		// Empty group query
		_, err = kv.GetByGroup(ctx, plugin.KVParams{Group: "", Page: 1, PageSize: 10})
		assertError(t, err, plugin.ErrKVGroupEmpty, "Empty group test")

		// Non-existent key
		assertNotFound(t, kv, ctx, "non_exist_group", "non_exist_key")

		// Cache penetration protection
		key := fmt.Sprintf("non_exist_key_%d", time.Now().UnixNano())
		assertNotFound(t, kv, ctx, "cache_penetration", key)
	})

	t.Run("CacheConsistency", func(t *testing.T) {
		mustSet(t, kv, ctx, "cache_group", "cache_key", "cache_value")
		mustGet(t, kv, ctx, "cache_group", "cache_key", "cache_value")

		// Update and verify immediate consistency
		mustSet(t, kv, ctx, "cache_group", "cache_key", "updated_value")
		mustGet(t, kv, ctx, "cache_group", "cache_key", "updated_value")
	})
}

// Test transactions including rollback and nested transactions
func TestTransactions(t *testing.T) {
	setupTestEnvironment()
	kv := testPlugin.operator
	ctx := context.Background()

	t.Run("SuccessfulTransaction", func(t *testing.T) {
		err := kv.Tx(ctx, func(ctx context.Context, txKv *plugin.KVOperator) error {
			if err := txKv.Set(ctx, plugin.KVParams{Group: "tx_group", Key: "tx_key1", Value: "tx_value1"}); err != nil {
				return err
			}
			if err := txKv.Set(ctx, plugin.KVParams{Group: "tx_group", Key: "tx_key2", Value: "tx_value2"}); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Successful transaction failed: %v", err)
		}

		mustGet(t, kv, ctx, "tx_group", "tx_key1", "tx_value1")
		mustGet(t, kv, ctx, "tx_group", "tx_key2", "tx_value2")
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		err := kv.Tx(ctx, func(ctx context.Context, txKv *plugin.KVOperator) error {
			if err := txKv.Set(ctx, plugin.KVParams{Group: "tx_group", Key: "tx_key3", Value: "tx_value3"}); err != nil {
				return err
			}
			return fmt.Errorf("mock error")
		})
		if err == nil {
			t.Error("Expected transaction to fail but it succeeded")
		}

		assertNotFound(t, kv, ctx, "tx_group", "tx_key3")
	})

	t.Run("NestedTransactions", func(t *testing.T) {
		err := kv.Tx(ctx, func(ctx context.Context, txKv *plugin.KVOperator) error {
			if err := txKv.Set(ctx, plugin.KVParams{Group: "nested", Key: "key1", Value: "value1"}); err != nil {
				return err
			}

			return txKv.Tx(ctx, func(ctx context.Context, nestedKv *plugin.KVOperator) error {
				if err := nestedKv.Set(ctx, plugin.KVParams{Group: "nested", Key: "key2", Value: "value2"}); err != nil {
					return err
				}
				return fmt.Errorf("mock nested error")
			})
		})
		if err == nil {
			t.Error("Expected nested transaction to fail but it succeeded")
		}

		// Verify outer transaction also rolled back
		assertNotFound(t, kv, ctx, "nested", "key1")
		assertNotFound(t, kv, ctx, "nested", "key2")
	})
}

// Test pagination in GetByGroup
func TestPagination(t *testing.T) {
	setupTestEnvironment()
	kv := testPlugin.operator
	ctx := context.Background()
	totalItems := 25

	for i := range totalItems {
		mustSet(t, kv, ctx, "pagination", fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Test pagination
	page1, err := kv.GetByGroup(ctx, plugin.KVParams{Group: "pagination", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}
	if len(page1) != 10 {
		t.Errorf("Page 1: expected 10 items, got %d", len(page1))
	}

	page2, err := kv.GetByGroup(ctx, plugin.KVParams{Group: "pagination", Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}
	if len(page2) != 10 {
		t.Errorf("Page 2: expected 10 items, got %d", len(page2))
	}

	page3, err := kv.GetByGroup(ctx, plugin.KVParams{Group: "pagination", Page: 3, PageSize: 10})
	if err != nil {
		t.Fatalf("Failed to get page 3: %v", err)
	}
	if len(page3) != 5 {
		t.Errorf("Page 3: expected 5 items, got %d", len(page3))
	}

	// Verify different keys on different pages
	for i := range 10 {
		key := fmt.Sprintf("key%d", i)
		if _, ok := page1[key]; !ok {
			t.Errorf("Pagination test failed, key %s should be on page 1", key)
		}
	}
	for i := range 10 {
		key := fmt.Sprintf("key%d", i+10)
		if _, ok := page2[key]; !ok {
			t.Errorf("Pagination test failed, key %s should be on page 2", key)
		}
	}
}

// Test concurrent operations and performance
func TestConcurrency(t *testing.T) {
	setupTestEnvironment()
	kv := testPlugin.operator
	ctx := context.Background()

	t.Run("BasicConcurrency", func(t *testing.T) {
		parallel := 10
		var wg sync.WaitGroup
		wg.Add(parallel)

		for i := range parallel {
			go func(index int) {
				defer wg.Done()
				time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
				mustSet(t, kv, ctx, "concurrent", fmt.Sprintf("key%d", index), "value")
			}(i)
		}
		wg.Wait()

		// Verify results
		wg.Add(parallel)
		for i := range parallel {
			go func(index int) {
				defer wg.Done()
				time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
				mustGet(t, kv, ctx, "concurrent", fmt.Sprintf("key%d", index), "value")
			}(i)
		}
		wg.Wait()
	})

	t.Run("StressTest", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping stress test in short mode")
		}

		totalOps := 1000
		workerCount := 20
		prefix := "stress_test"
		opsPerWorker := totalOps / workerCount

		log.Info("Starting KV storage stress test...")
		startTime := time.Now()

		// Concurrent write test
		var wg sync.WaitGroup
		errorCount := int64(0)

		for w := range workerCount {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				startIdx := workerID * opsPerWorker

				for i := range opsPerWorker {
					i := startIdx + i
					err := kv.Set(ctx, plugin.KVParams{
						Group: prefix,
						Key:   fmt.Sprintf("key%d", i),
						Value: fmt.Sprintf("value%d", i),
					})
					if err != nil {
						log.Warnf("Write error: %v", err)
						errorCount++
					}
				}
			}(w)
		}
		wg.Wait()

		writeTime := time.Since(startTime)

		// Verify data integrity
		groupData, err := kv.GetByGroup(ctx, plugin.KVParams{Group: prefix, Page: 1, PageSize: totalOps})
		if err != nil {
			t.Fatalf("Failed to verify data: %v", err)
		}
		if len(groupData) != totalOps {
			t.Errorf("Data loss: expected %d items, got %d", totalOps, len(groupData))
		}

		// Concurrent read test
		startTime = time.Now()
		readErrors := int64(0)

		wg.Add(workerCount)
		for range workerCount {
			go func() {
				defer wg.Done()
				for range opsPerWorker {
					keyIdx := rand.Intn(totalOps)
					key := fmt.Sprintf("key%d", keyIdx)
					expected := fmt.Sprintf("value%d", keyIdx)

					val, err := kv.Get(ctx, plugin.KVParams{Group: prefix, Key: key})
					if err != nil {
						readErrors++
					} else if val != expected {
						t.Errorf("Data inconsistency: key=%s, expected=%s, got=%s", key, expected, val)
					}
				}
			}()
		}
		wg.Wait()

		readTime := time.Since(startTime)

		log.Infof("Stress test completed:")
		log.Infof("  Write: %d ops in %v (%.1f ops/sec), %d errors", totalOps, writeTime, float64(totalOps)/writeTime.Seconds(), errorCount)
		log.Infof("  Read: %d ops in %v (%.1f ops/sec), %d errors", totalOps, readTime, float64(totalOps)/readTime.Seconds(), readErrors)
	})
}

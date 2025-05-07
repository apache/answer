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

package plugin

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/cache"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Error variables for KV storage operations
var (
	// ErrKVKeyNotFound is returned when the requested key does not exist in the KV storage
	ErrKVKeyNotFound = fmt.Errorf("key not found in KV storage")
	// ErrKVGroupEmpty is returned when a required group name is empty
	ErrKVGroupEmpty = fmt.Errorf("group name is empty")
	// ErrKVKeyEmpty is returned when a required key name is empty
	ErrKVKeyEmpty = fmt.Errorf("key name is empty")
	// ErrKVKeyAndGroupEmpty is returned when both key and group names are empty
	ErrKVKeyAndGroupEmpty = fmt.Errorf("both key and group are empty")
	// ErrKVTransactionFailed is returned when a KV storage transaction operation fails
	ErrKVTransactionFailed = fmt.Errorf("KV storage transaction failed")
)

// KVParams is the parameters for KV storage operations
type KVParams struct {
	Group    string
	Key      string
	Value    string
	Page     int
	PageSize int
}

// KVOperator provides methods to interact with the key-value storage system for plugins
type KVOperator struct {
	data           *Data
	session        *xorm.Session
	pluginSlugName string
	cacheTTL       time.Duration
}

// KVStorageOption defines a function type that configures a KVOperator
type KVStorageOption func(*KVOperator)

// WithCacheTTL is the option to set the cache TTL; the default value is 30 minutes.
// If ttl is less than 0, the cache will not be used
func WithCacheTTL(ttl time.Duration) KVStorageOption {
	return func(kv *KVOperator) {
		kv.cacheTTL = ttl
	}
}

// Option is used to set the options for the KV storage
func (kv *KVOperator) Option(opts ...KVStorageOption) {
	for _, opt := range opts {
		opt(kv)
	}
}

func (kv *KVOperator) getSession(ctx context.Context) (*xorm.Session, func()) {
	session := kv.session
	cleanup := func() {}
	if session == nil {
		session = kv.data.DB.NewSession().Context(ctx)
		cleanup = func() {
			if session != nil {
				session.Close()
			}
		}
	}
	return session, cleanup
}

func (kv *KVOperator) getCacheKey(params KVParams) string {
	return fmt.Sprintf("plugin_kv_storage:%s:group:%s:key:%s", kv.pluginSlugName, params.Group, params.Key)
}

func (kv *KVOperator) setCache(ctx context.Context, params KVParams) {
	if kv.cacheTTL < 0 {
		return
	}

	ttl := kv.cacheTTL
	if ttl > 10 {
		ttl += time.Duration(float64(ttl) * 0.1 * (1 - rand.Float64()))
	}

	cacheKey := kv.getCacheKey(params)
	if err := kv.data.Cache.SetString(ctx, cacheKey, params.Value, ttl); err != nil {
		log.Warnf("cache set failed: %v, key: %s", err, cacheKey)
	}
}

func (kv *KVOperator) getCache(ctx context.Context, params KVParams) (string, bool, error) {
	if kv.cacheTTL < 0 {
		return "", false, nil
	}

	cacheKey := kv.getCacheKey(params)
	return kv.data.Cache.GetString(ctx, cacheKey)
}

func (kv *KVOperator) cleanCache(ctx context.Context, params KVParams) {
	if kv.cacheTTL < 0 {
		return
	}

	if err := kv.data.Cache.Del(ctx, kv.getCacheKey(params)); err != nil {
		log.Warnf("Failed to delete cache for key %s: %v", params.Key, err)
	}
}

// Get retrieves a value from KV storage by group and key.
// Returns the value as a string or an error if the key is not found.
func (kv *KVOperator) Get(ctx context.Context, params KVParams) (string, error) {
	if params.Key == "" {
		return "", ErrKVKeyEmpty
	}

	if value, exist, err := kv.getCache(ctx, params); err == nil && exist {
		return value, nil
	}

	// query
	data := entity.PluginKVStorage{}
	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	query.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
		"`group`":          params.Group,
		"`key`":            params.Key,
	})

	has, err := query.Get(&data)
	if err != nil {
		return "", err
	}
	if !has {
		return "", ErrKVKeyNotFound
	}

	params.Value = data.Value
	kv.setCache(ctx, params)

	return data.Value, nil
}

// Set stores a value in KV storage with the specified group and key.
// Updates the value if it already exists.
func (kv *KVOperator) Set(ctx context.Context, params KVParams) error {
	if params.Key == "" {
		return ErrKVKeyEmpty
	}

	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	data := &entity.PluginKVStorage{
		PluginSlugName: kv.pluginSlugName,
		Group:          params.Group,
		Key:            params.Key,
		Value:          params.Value,
	}

	kv.cleanCache(ctx, params)

	affected, err := query.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
		"`group`":          params.Group,
		"`key`":            params.Key,
	}).Cols("value").Update(data)
	if err != nil {
		return err
	}

	if affected == 0 {
		_, err = query.Insert(data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Del removes values from KV storage by group and/or key.
// If both group and key are provided, only that specific entry is deleted.
// If only group is provided, all entries in that group are deleted.
// At least one of group or key must be provided.
func (kv *KVOperator) Del(ctx context.Context, params KVParams) error {
	if params.Key == "" && params.Group == "" {
		return ErrKVKeyAndGroupEmpty
	}

	kv.cleanCache(ctx, params)

	session, cleanup := kv.getSession(ctx)
	defer cleanup()

	session.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
	})
	if params.Group != "" {
		session.Where(builder.Eq{"`group`": params.Group})
	}
	if params.Key != "" {
		session.Where(builder.Eq{"`key`": params.Key})
	}

	_, err := session.Delete(&entity.PluginKVStorage{})
	return err
}

// GetByGroup retrieves all key-value pairs for a specific group with pagination support.
// Returns a map of keys to values or an error if the group is empty or not found.
func (kv *KVOperator) GetByGroup(ctx context.Context, params KVParams) (map[string]string, error) {
	if params.Group == "" {
		return nil, ErrKVGroupEmpty
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}

	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	var items []entity.PluginKVStorage
	err := query.Where(builder.Eq{"plugin_slug_name": kv.pluginSlugName, "`group`": params.Group}).
		Limit(params.PageSize, (params.Page-1)*params.PageSize).
		OrderBy("id ASC").
		Find(&items)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.Key] = item.Value
	}

	return result, nil
}

// Tx executes a function within a transaction context. If the KVOperator already has a session,
// it will use that session. Otherwise, it creates a new transaction session.
// The transaction will be committed if the function returns nil, or rolled back if it returns an error.
func (kv *KVOperator) Tx(ctx context.Context, fn func(ctx context.Context, kv *KVOperator) error) error {
	var (
		txKv         = kv
		shouldCommit bool
	)

	if kv.session == nil {
		session := kv.data.DB.NewSession().Context(ctx)
		if err := session.Begin(); err != nil {
			session.Close()
			return fmt.Errorf("%w: begin transaction failed: %v", ErrKVTransactionFailed, err)
		}

		defer func() {
			if !shouldCommit {
				if rollbackErr := session.Rollback(); rollbackErr != nil {
					log.Errorf("rollback failed: %v", rollbackErr)
				}
			}
			session.Close()
		}()

		txKv = &KVOperator{
			session:        session,
			data:           kv.data,
			pluginSlugName: kv.pluginSlugName,
		}
		shouldCommit = true
	}

	if err := fn(ctx, txKv); err != nil {
		return fmt.Errorf("%w: %v", ErrKVTransactionFailed, err)
	}

	if shouldCommit {
		if err := txKv.session.Commit(); err != nil {
			return fmt.Errorf("%w: commit failed: %v", ErrKVTransactionFailed, err)
		}
	}
	return nil
}

// KVStorage defines the interface for plugins that need data storage capabilities
type KVStorage interface {
	Info() Info
	SetOperator(operator *KVOperator)
}

var (
	CallKVStorage,
	registerKVStorage = MakePlugin[KVStorage](true)
)

// NewKVOperator creates a new KV storage operator with the specified database engine, cache and plugin name.
// It returns a KVOperator instance that can be used to interact with the plugin's storage.
func NewKVOperator(db *xorm.Engine, cache cache.Cache, pluginSlugName string) *KVOperator {
	return &KVOperator{
		data:           &Data{DB: db, Cache: cache},
		pluginSlugName: pluginSlugName,
		cacheTTL:       30 * time.Minute,
	}
}

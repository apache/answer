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
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// define error
var (
	ErrKVKeyNotFound        = fmt.Errorf("key not found in KV storage")
	ErrKVGroupEmpty         = fmt.Errorf("group name is empty")
	ErrKVKeyEmpty           = fmt.Errorf("key name is empty")
	ErrKVKeyAndGroupEmpty   = fmt.Errorf("both key and group are empty")
	ErrKVTransactionFailed  = fmt.Errorf("KV storage transaction failed")
	ErrKVDataNotInitialized = fmt.Errorf("KV storage data not initialized")
	ErrKVDBNotInitialized   = fmt.Errorf("KV storage database connection not initialized")
)

type KVOperator struct {
	data           *Data
	session        *xorm.Session
	pluginSlugName string
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

func (kv *KVOperator) getCacheTTL() time.Duration {
	return 30*time.Minute + time.Duration(rand.Intn(300))*time.Second
}

func (kv *KVOperator) getCacheKey(group, key string) string {
	if group == "" {
		return fmt.Sprintf("plugin_kv_storage:%s:key:%s", kv.pluginSlugName, key)
	}
	if key == "" {
		return fmt.Sprintf("plugin_kv_storage:%s:group:%s", kv.pluginSlugName, group)
	}
	return fmt.Sprintf("plugin_kv_storage:%s:group:%s:key:%s", kv.pluginSlugName, group, key)
}

func (kv *KVOperator) Get(ctx context.Context, group, key string) (string, error) {
	// validate
	if key == "" {
		return "", ErrKVKeyEmpty
	}

	cacheKey := kv.getCacheKey(group, key)
	if value, exist, err := kv.data.Cache.GetString(ctx, cacheKey); err == nil && exist {
		return value, nil
	}

	// query
	data := entity.PluginKVStorage{}
	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	query.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
		"`group`":          group,
		"`key`":            key,
	})

	has, err := query.Get(&data)
	if err != nil {
		return "", err
	}
	if !has {
		return "", ErrKVKeyNotFound
	}

	if err := kv.data.Cache.SetString(ctx, cacheKey, data.Value, kv.getCacheTTL()); err != nil {
		log.Error(err)
	}

	return data.Value, nil
}

func (kv *KVOperator) Set(ctx context.Context, group, key, value string) error {
	if key == "" {
		return ErrKVKeyEmpty
	}

	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	data := &entity.PluginKVStorage{
		PluginSlugName: kv.pluginSlugName,
		Group:          group,
		Key:            key,
		Value:          value,
	}

	kv.cleanCache(ctx, group, key)

	affected, err := query.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
		"`group`":          group,
		"`key`":            key,
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

func (kv *KVOperator) Del(ctx context.Context, group, key string) error {
	if key == "" && group == "" {
		return ErrKVKeyAndGroupEmpty
	}

	kv.cleanCache(ctx, group, key)

	session, cleanup := kv.getSession(ctx)
	defer cleanup()

	session.Where(builder.Eq{
		"plugin_slug_name": kv.pluginSlugName,
	})
	if group != "" {
		session.Where(builder.Eq{"`group`": group})
	}
	if key != "" {
		session.Where(builder.Eq{"`key`": key})
	}

	_, err := session.Delete(&entity.PluginKVStorage{})
	return err
}

func (kv *KVOperator) cleanCache(ctx context.Context, group, key string) {
	if key != "" {
		if err := kv.data.Cache.Del(ctx, kv.getCacheKey("", key)); err != nil {
			log.Warnf("Failed to delete cache for key %s: %v", key, err)
		}

		if group != "" {
			if err := kv.data.Cache.Del(ctx, kv.getCacheKey(group, key)); err != nil {
				log.Warnf("Failed to delete cache for group %s, key %s: %v", group, key, err)
			}
		}
	}

	if group != "" {
		if err := kv.data.Cache.Del(ctx, kv.getCacheKey(group, "")); err != nil {
			log.Warnf("Failed to delete cache for group %s: %v", group, err)
		}
	}
}

func (kv *KVOperator) GetByGroup(ctx context.Context, group string, page, pageSize int) (map[string]string, error) {
	if group == "" {
		return nil, ErrKVGroupEmpty
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	cacheKey := kv.getCacheKey(group, "")
	if value, exist, err := kv.data.Cache.GetString(ctx, cacheKey); err == nil && exist {
		result := make(map[string]string)
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result, nil
		}
	}

	query, cleanup := kv.getSession(ctx)
	defer cleanup()

	var items []entity.PluginKVStorage
	err := query.Where(builder.Eq{"plugin_slug_name": kv.pluginSlugName, "`group`": group}).
		Limit(pageSize, (page-1)*pageSize).
		OrderBy("id ASC").
		Find(&items)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.Key] = item.Value
		if err := kv.data.Cache.SetString(ctx, kv.getCacheKey(group, item.Key), item.Value, kv.getCacheTTL()); err != nil {
			log.Warnf("Failed to set cache for group %s, key %s: %v", group, item.Key, err)
		}
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		_ = kv.data.Cache.SetString(ctx, cacheKey, string(resultJSON), kv.getCacheTTL())
	}

	return result, nil
}

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

// PluginData defines the interface for plugins that need data storage capabilities
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

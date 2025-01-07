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
	"path/filepath"
	"time"

	"github.com/apache/answer/pkg/dir"
	"github.com/apache/answer/plugin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/segmentfault/pacman/cache"
	"github.com/segmentfault/pacman/contrib/cache/memory"
	"github.com/segmentfault/pacman/log"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
	ormlog "xorm.io/xorm/log"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

// Data data
type Data struct {
	DB    *xorm.Engine
	Cache cache.Cache
}

// NewData new data instance
func NewData(db *xorm.Engine, cache cache.Cache) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
		db.Close()
	}
	return &Data{DB: db, Cache: cache}, cleanup, nil
}

// NewDB new database instance
func NewDB(debug bool, dataConf *Database) (*xorm.Engine, error) {
	if dataConf.Driver == "" {
		dataConf.Driver = string(schemas.MYSQL)
	}
	if dataConf.Driver == string(schemas.SQLITE) {
		dataConf.Driver = "sqlite"
		dbFileDir := filepath.Dir(dataConf.Connection)
		log.Debugf("try to create database directory %s", dbFileDir)
		if err := dir.CreateDirIfNotExist(dbFileDir); err != nil {
			log.Errorf("create database dir failed: %s", err)
		}
		dataConf.MaxOpenConn = 1
	}
	engine, err := xorm.NewEngine(dataConf.Driver, dataConf.Connection)
	if err != nil {
		return nil, err
	}

	if debug {
		engine.ShowSQL(true)
	} else {
		engine.SetLogLevel(ormlog.LOG_ERR)
	}

	if err = engine.Ping(); err != nil {
		return nil, err
	}

	if dataConf.MaxIdleConn > 0 {
		engine.SetMaxIdleConns(dataConf.MaxIdleConn)
	}
	if dataConf.MaxOpenConn > 0 {
		engine.SetMaxOpenConns(dataConf.MaxOpenConn)
	}
	if dataConf.ConnMaxLifeTime > 0 {
		engine.SetConnMaxLifetime(time.Duration(dataConf.ConnMaxLifeTime) * time.Second)
	}
	engine.SetColumnMapper(names.GonicMapper{})
	return engine, nil
}

// NewCache new cache instance
func NewCache(c *CacheConf) (cache.Cache, func(), error) {
	var pluginCache plugin.Cache
	_ = plugin.CallCache(func(fn plugin.Cache) error {
		pluginCache = fn
		return nil
	})
	if pluginCache != nil {
		return pluginCache, func() {}, nil
	}

	// TODO What cache type should be initialized according to the configuration file
	memCache := memory.NewCache()

	if len(c.FilePath) > 0 {
		cacheFileDir := filepath.Dir(c.FilePath)
		log.Debugf("try to create cache directory %s", cacheFileDir)
		err := dir.CreateDirIfNotExist(cacheFileDir)
		if err != nil {
			log.Errorf("create cache dir failed: %s", err)
		}
		log.Infof("try to load cache file from %s", c.FilePath)
		if err := memory.Load(memCache, c.FilePath); err != nil {
			log.Warn(err)
		}
		go func() {
			ticker := time.Tick(time.Minute)
			for range ticker {
				if err := memory.Save(memCache, c.FilePath); err != nil {
					log.Warn(err)
				}
			}
		}()
	}
	cleanup := func() {
		log.Infof("try to save cache file to %s", c.FilePath)
		if err := memory.Save(memCache, c.FilePath); err != nil {
			log.Warn(err)
		}
	}
	return memCache, cleanup, nil
}

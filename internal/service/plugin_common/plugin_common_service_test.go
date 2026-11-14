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

package plugin_common

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	basedata "github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	configRepo "github.com/apache/answer/internal/repo/config"
	"github.com/apache/answer/internal/schema"
	configService "github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/contrib/cache/memory"
)

func TestUpdatePluginStatusMergesConcurrentReplicaChanges(t *testing.T) {
	baseService, db := newPluginCommonTestService(t, &pluginCommonTestConfigRepo{})
	defer func() {
		if err := db.DB.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	}()
	if _, err := baseService.configService.GetStringValue(context.Background(), constant.PluginStatus); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	for _, slugName := range []string{"replica-a-plugin", "replica-b-plugin"} {
		replicaService := &PluginCommonService{configService: baseService.configService}
		go func(slugName string, service *PluginCommonService) {
			<-start
			results <- service.UpdatePluginStatus(context.Background(), slugName, true)
		}(slugName, replicaService)
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}

	stored, err := baseService.configService.GetStringValueFromDB(context.Background(), constant.PluginStatus)
	if err != nil {
		t.Fatal(err)
	}
	statuses := make(map[string]bool)
	if err := json.Unmarshal([]byte(stored), &statuses); err != nil {
		t.Fatal(err)
	}
	if !statuses["replica-a-plugin"] || !statuses["replica-b-plugin"] {
		t.Fatalf("status map lost a concurrent update: %s", stored)
	}
	cached, err := baseService.configService.GetStringValue(context.Background(), constant.PluginStatus)
	if err != nil {
		t.Fatal(err)
	}
	if cached != stored {
		t.Fatalf("cached status = %s, want latest database value %s", cached, stored)
	}
}

func TestRefreshPluginStatusAndConfigFromDatabase(t *testing.T) {
	registerPluginCommonConfigTestPlugin(t)
	oldStatus, err := plugin.StatusManager.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = plugin.StatusManager.UnmarshalJSON(oldStatus) })

	repo := &pluginCommonTestConfigRepo{configs: map[string]string{
		pluginCommonConfigTestSlug: `{"value":"old"}`,
	}}
	service, db := newPluginCommonTestService(t, repo)
	defer func() {
		if err := db.DB.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	}()

	ctx := context.Background()
	if err := plugin.StatusManager.UnmarshalJSON([]byte(`{"plugin-common-config-refresh-test":false}`)); err != nil {
		t.Fatal(err)
	}
	if err := service.refreshPluginConfigs(ctx, true, false); err != nil {
		t.Fatal(err)
	}
	if got := pluginCommonConfigTestInstance.configValue(); got != `{"value":"old"}` {
		t.Fatalf("initial plugin config = %s, want old config", got)
	}

	if err := service.configService.UpdateConfig(ctx, constant.PluginStatus,
		`{"plugin-common-config-refresh-test":true}`); err != nil {
		t.Fatal(err)
	}
	repo.set(pluginCommonConfigTestSlug, `{"value":"new"}`)

	service.refreshPluginStatus(ctx)
	if err := service.refreshPluginConfigs(ctx, false, false); err != nil {
		t.Fatal(err)
	}
	if !plugin.StatusManager.IsEnabled(pluginCommonConfigTestSlug) {
		t.Fatal("plugin status was not refreshed from the database")
	}
	if got := pluginCommonConfigTestInstance.configValue(); got != `{"value":"new"}` {
		t.Fatalf("refreshed plugin config = %s, want new config", got)
	}
}

func newPluginCommonTestService(
	t *testing.T,
	pluginConfigRepo PluginConfigRepo,
) (*PluginCommonService, *basedata.Data) {
	t.Helper()

	db, err := basedata.NewDB(false, &basedata.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	data := &basedata.Data{DB: db, Cache: memory.NewCache()}
	if err := db.Sync(new(entity.Config)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Insert(&entity.Config{Key: constant.PluginStatus, Value: `{}`}); err != nil {
		t.Fatal(err)
	}

	repo := configRepo.NewConfigRepo(data)
	service := &PluginCommonService{
		configService:      configService.NewConfigService(repo),
		pluginConfigRepo:   pluginConfigRepo,
		pluginConfigValues: make(map[string]string),
	}
	return service, data
}

const pluginCommonConfigTestSlug = "plugin-common-config-refresh-test"

var (
	pluginCommonConfigTestOnce     sync.Once
	pluginCommonConfigTestInstance = &pluginCommonTestConfigPlugin{}
)

func registerPluginCommonConfigTestPlugin(t *testing.T) {
	t.Helper()
	pluginCommonConfigTestOnce.Do(func() {
		plugin.Register(pluginCommonConfigTestInstance)
	})
}

type pluginCommonTestConfigPlugin struct {
	mu    sync.Mutex
	value string
}

func (p *pluginCommonTestConfigPlugin) Info() plugin.Info {
	return plugin.Info{SlugName: pluginCommonConfigTestSlug}
}

func (p *pluginCommonTestConfigPlugin) ConfigFields() []plugin.ConfigField { return nil }

func (p *pluginCommonTestConfigPlugin) ConfigReceiver(config []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = string(config)
	return nil
}

func (p *pluginCommonTestConfigPlugin) configValue() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.value
}

type pluginCommonTestConfigRepo struct {
	mu        sync.RWMutex
	configs   map[string]string
	readError error
	afterRead func()
}

func (r *pluginCommonTestConfigRepo) SavePluginConfig(
	_ context.Context, slugName, value string,
) error {
	r.set(slugName, value)
	return nil
}

func (r *pluginCommonTestConfigRepo) GetPluginConfigAll(
	context.Context,
) ([]*entity.PluginConfig, error) {
	r.mu.RLock()
	configs := make([]*entity.PluginConfig, 0, len(r.configs))
	for slugName, value := range r.configs {
		configs = append(configs, &entity.PluginConfig{PluginSlugName: slugName, Value: value})
	}
	r.mu.RUnlock()
	if r.afterRead != nil {
		r.afterRead()
	}
	return configs, r.readError
}

func (r *pluginCommonTestConfigRepo) set(slugName, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.configs == nil {
		r.configs = make(map[string]string)
	}
	r.configs[slugName] = value
}

func TestPluginStatusSnapshotTTL(t *testing.T) {
	service, repo := newPluginStatusSnapshotTestService(t)
	ctx := context.Background()
	setStoredPluginStatus(t, repo, `{"first":true,"second":true}`)

	// Startup loads the entire snapshot before any normal plugin check.
	service.refreshPluginStatus(ctx)
	setStoredPluginStatus(t, repo, `{"first":false,"third":true}`)
	if !plugin.StatusManager.IsEnabled("first") || !plugin.StatusManager.IsEnabled("second") {
		t.Fatal("checks before expiry did not use the startup snapshot")
	}
	if got := repo.reads.Load(); got != 1 {
		t.Fatalf("database reads before expiry = %d, want 1", got)
	}

	expirePluginStatusSnapshot(service)
	if plugin.StatusManager.IsEnabled("first") || plugin.StatusManager.IsEnabled("second") ||
		!plugin.StatusManager.IsEnabled("third") {
		t.Fatal("first check after expiry did not replace the whole snapshot")
	}
	if got := repo.reads.Load(); got != 2 {
		t.Fatalf("database reads after expiry = %d, want 2", got)
	}
}

func TestPluginStatusSnapshotCoalescesExpiredChecks(t *testing.T) {
	_, repo := newPluginStatusSnapshotTestService(t)
	setStoredPluginStatus(t, repo, `{"enabled":true}`)
	readStarted, releaseRead := make(chan struct{}), make(chan struct{})
	repo.afterRead = func() {
		close(readStarted)
		<-releaseRead
	}

	var checks sync.WaitGroup
	for range 32 {
		checks.Go(func() {
			if !plugin.StatusManager.IsEnabled("enabled") {
				t.Error("expired check did not see the refreshed status")
			}
		})
	}
	<-readStarted
	close(releaseRead)
	checks.Wait()
	if got := repo.reads.Load(); got != 1 {
		t.Fatalf("database reads for concurrent expired checks = %d, want 1", got)
	}
}

func TestPluginStatusSnapshotKeepsLastGoodValueOnFailure(t *testing.T) {
	for _, failure := range []string{"database", "invalid JSON"} {
		t.Run(failure, func(t *testing.T) {
			service, repo := newPluginStatusSnapshotTestService(t)
			if err := service.UpdatePluginStatus(context.Background(), "enabled", true); err != nil {
				t.Fatal(err)
			}
			if failure == "database" {
				repo.failRead.Store(true)
			} else {
				setStoredPluginStatus(t, repo, `{`)
			}
			expirePluginStatusSnapshot(service)
			reads := repo.reads.Load()
			for range 10 {
				if !plugin.StatusManager.IsEnabled("enabled") {
					t.Fatal("failed refresh overwrote a successful local update")
				}
			}
			if got := repo.reads.Load() - reads; got != 1 {
				t.Fatalf("database reads during failed refresh = %d, want 1", got)
			}

			repo.failRead.Store(false)
			setStoredPluginStatus(t, repo, `{"enabled":false}`)
			expirePluginStatusSnapshot(service)
			if plugin.StatusManager.IsEnabled("enabled") {
				t.Fatal("refresh did not recover after the failure")
			}
		})
	}
}

func TestPluginStatusSnapshotSerializesRefreshWithLocalUpdate(t *testing.T) {
	service, repo := newPluginStatusSnapshotTestService(t)
	setStoredPluginStatus(t, repo, `{"enabled":false}`)
	readStarted, releaseRead := make(chan struct{}), make(chan struct{})
	var firstRead sync.Once
	repo.afterRead = func() {
		firstRead.Do(func() {
			close(readStarted)
			<-releaseRead
		})
	}
	refreshDone := make(chan struct{})
	go func() {
		service.refreshPluginStatus(context.Background())
		close(refreshDone)
	}()
	<-readStarted
	updateStarted, updateDone := make(chan struct{}), make(chan error, 1)
	go func() {
		close(updateStarted)
		updateDone <- service.UpdatePluginStatus(context.Background(), "enabled", true)
	}()
	<-updateStarted
	close(releaseRead)
	<-refreshDone
	if err := <-updateDone; err != nil {
		t.Fatal(err)
	}
	reads := repo.reads.Load()
	if !plugin.StatusManager.IsEnabled("enabled") {
		t.Fatal("older refresh overwrote a successful local update")
	}
	if got := repo.reads.Load(); got != reads {
		t.Fatal("successful local update did not renew the snapshot TTL")
	}
}

func newPluginStatusSnapshotTestService(t *testing.T) (*PluginCommonService, *pluginStatusSnapshotTestRepo) {
	t.Helper()
	oldStatus, err := plugin.StatusManager.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	service, data := newPluginCommonTestService(t, &pluginCommonTestConfigRepo{})
	repo := &pluginStatusSnapshotTestRepo{ConfigRepo: configRepo.NewConfigRepo(data)}
	service.configService = configService.NewConfigService(repo)
	plugin.StatusManager.SetRefreshFunc(func() { service.refreshPluginStatus(context.Background()) })
	t.Cleanup(func() {
		plugin.StatusManager.SetRefreshFunc(nil)
		if err := plugin.StatusManager.UnmarshalJSON(oldStatus); err != nil {
			t.Error(err)
		}
		if err := data.DB.Close(); err != nil {
			t.Error(err)
		}
	})
	return service, repo
}

func expirePluginStatusSnapshot(service *PluginCommonService) {
	service.pluginStatusMu.Lock()
	service.pluginStatusExpires = time.Time{}
	service.pluginStatusMu.Unlock()
}

func setStoredPluginStatus(t *testing.T, repo *pluginStatusSnapshotTestRepo, value string) {
	t.Helper()
	if err := repo.UpdateConfig(context.Background(), constant.PluginStatus, value); err != nil {
		t.Fatal(err)
	}
}

type pluginStatusSnapshotTestRepo struct {
	configService.ConfigRepo
	reads     atomic.Int64
	failRead  atomic.Bool
	afterRead func()
}

func (r *pluginStatusSnapshotTestRepo) GetConfigByKey(context.Context, string) (*entity.Config, error) {
	return nil, fmt.Errorf("status refresh must bypass the config cache")
}

func (r *pluginStatusSnapshotTestRepo) GetConfigByKeyFromDB(ctx context.Context, key string) (*entity.Config, error) {
	r.reads.Add(1)
	if r.failRead.Load() {
		return nil, fmt.Errorf("database unavailable")
	}
	value, err := r.ConfigRepo.GetConfigByKeyFromDB(ctx, key)
	if r.afterRead != nil {
		r.afterRead()
	}
	return value, err
}

func TestPluginConfigRefreshKeepsLocalUpdateOnFailure(t *testing.T) {
	registerPluginCommonConfigTestPlugin(t)
	repo := &pluginCommonTestConfigRepo{}
	service, data := newPluginCommonTestService(t, repo)
	t.Cleanup(func() {
		if err := data.DB.Close(); err != nil {
			t.Error(err)
		}
	})
	req := &schema.UpdatePluginConfigReq{
		PluginSlugName: pluginCommonConfigTestSlug,
		ConfigFields:   map[string]any{"value": "local"},
	}
	if err := service.UpdatePluginConfig(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	repo.readError = fmt.Errorf("database unavailable")
	if err := service.refreshPluginConfigs(context.Background(), false, false); err == nil {
		t.Fatal("expected refresh to fail")
	}
	if got := pluginCommonConfigTestInstance.configValue(); got != `{"value":"local"}` {
		t.Fatalf("failed refresh replaced local config with %s", got)
	}
	if got := service.pluginConfigValues[pluginCommonConfigTestSlug]; got != `{"value":"local"}` {
		t.Fatalf("failed refresh replaced applied config snapshot with %s", got)
	}
}

func TestPluginConfigRefreshSerializesWithLocalUpdate(t *testing.T) {
	registerPluginCommonConfigTestPlugin(t)
	repo := &pluginCommonTestConfigRepo{configs: map[string]string{
		pluginCommonConfigTestSlug: `{"value":"old"}`,
	}}
	service, data := newPluginCommonTestService(t, repo)
	t.Cleanup(func() {
		if err := data.DB.Close(); err != nil {
			t.Error(err)
		}
	})
	readStarted, releaseRead := make(chan struct{}), make(chan struct{})
	repo.afterRead = func() {
		close(readStarted)
		<-releaseRead
	}
	refreshDone := make(chan error, 1)
	go func() {
		refreshDone <- service.refreshPluginConfigs(context.Background(), false, false)
	}()
	<-readStarted
	updateStarted, updateDone := make(chan struct{}), make(chan error, 1)
	go func() {
		close(updateStarted)
		updateDone <- service.UpdatePluginConfig(context.Background(), &schema.UpdatePluginConfigReq{
			PluginSlugName: pluginCommonConfigTestSlug,
			ConfigFields:   map[string]any{"value": "local"},
		})
	}()
	<-updateStarted
	close(releaseRead)
	if err := <-refreshDone; err != nil {
		t.Fatal(err)
	}
	if err := <-updateDone; err != nil {
		t.Fatal(err)
	}
	if got := pluginCommonConfigTestInstance.configValue(); got != `{"value":"local"}` {
		t.Fatalf("older refresh replaced successful local update with %s", got)
	}
	if got := service.pluginConfigValues[pluginCommonConfigTestSlug]; got != `{"value":"local"}` {
		t.Fatalf("older refresh replaced applied config snapshot with %s", got)
	}
}

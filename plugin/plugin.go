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
	"encoding/json"
	"sync"

	"github.com/segmentfault/pacman/cache"
	"github.com/segmentfault/pacman/i18n"
	"xorm.io/xorm"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/gin-gonic/gin"
)

// Data is defined here to avoid circular dependency with internal/base/data
type Data struct {
	DB    *xorm.Engine
	Cache cache.Cache
}

// GinContext is a wrapper of gin.Context
// We export it to make it easy to use in plugins
type GinContext = gin.Context

// StatusManager is a manager that manages the status of plugins
// Init Plugins:
// json.Unmarshal([]byte(`{"plugin1": true, "plugin2": false}`), &plugin.StatusManager)
// Dump Status:
// json.Marshal(plugin.StatusManager)
var StatusManager = statusManager{
	status: make(map[string]bool),
}

// Register registers a plugin
func Register(p Base) {
	registerBase(p)

	if _, ok := p.(Config); ok {
		registerConfig(p.(Config))
	}

	if _, ok := p.(UserConfig); ok {
		registerUserConfig(p.(UserConfig))
	}

	if _, ok := p.(Connector); ok {
		registerConnector(p.(Connector))
	}

	if _, ok := p.(Parser); ok {
		registerParser(p.(Parser))
	}

	if _, ok := p.(Filter); ok {
		registerFilter(p.(Filter))
	}

	if _, ok := p.(Storage); ok {
		registerStorage(p.(Storage))
	}

	if _, ok := p.(Cache); ok {
		registerCache(p.(Cache))
	}

	if _, ok := p.(UserCenter); ok {
		registerUserCenter(p.(UserCenter))
	}

	if _, ok := p.(Agent); ok {
		registerAgent(p.(Agent))
	}

	if _, ok := p.(Search); ok {
		registerSearch(p.(Search))
	}

	if _, ok := p.(Notification); ok {
		registerNotification(p.(Notification))
	}

	if _, ok := p.(Reviewer); ok {
		registerReviewer(p.(Reviewer))
	}

	if _, ok := p.(Captcha); ok {
		registerCaptcha(p.(Captcha))
	}

	if _, ok := p.(Embed); ok {
		registerEmbed(p.(Embed))
	}

	if _, ok := p.(Render); ok {
		registerRender(p.(Render))
	}

	if _, ok := p.(CDN); ok {
		registerCDN(p.(CDN))
	}

	if _, ok := p.(Importer); ok {
		registerImporter(p.(Importer))
	}

	if _, ok := p.(KVStorage); ok {
		registerKVStorage(p.(KVStorage))
	}
}

type Stack[T Base] struct {
	plugins []T
}

type RegisterFn[T Base] func(p T)
type Caller[T Base] func(p T) error
type CallFn[T Base] func(fn Caller[T]) error

// MakePlugin creates a plugin caller and register stack manager
// The parameter super presents if the plugin can be disabled.
// It returns a register function and a caller function
// The register function is used to register a plugin, it will be called in the plugin's init function
// The caller function is used to call all registered plugins
func MakePlugin[T Base](super bool) (CallFn[T], RegisterFn[T]) {
	stack := Stack[T]{}

	call := func(fn Caller[T]) error {
		for _, p := range stack.plugins {
			// If the plugin is disabled, skip it
			if !super && !StatusManager.IsEnabled(p.Info().SlugName) {
				continue
			}

			if err := fn(p); err != nil {
				return err
			}
		}
		return nil
	}

	register := func(p T) {
		for _, plugin := range stack.plugins {
			if plugin.Info().SlugName == p.Info().SlugName {
				panic("plugin " + p.Info().SlugName + " is already registered")
			}
		}
		stack.plugins = append(stack.plugins, p)
	}

	return call, register
}

type statusManager struct {
	lock   sync.Mutex
	status map[string]bool
}

func (m *statusManager) Enable(name string, enabled bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !enabled {
		m.status[name] = enabled
		return
	}
	m.status[name] = enabled

	for _, slugName := range coordinatedCaptchaPlugins(name) {
		m.status[slugName] = false
	}

	for _, slugName := range coordinatedCDNPlugins(name) {
		m.status[slugName] = false
	}
}

func (m *statusManager) IsEnabled(name string) bool {
	if status, ok := m.status[name]; ok {
		return status
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface.
func (m *statusManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.status)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *statusManager) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.status)
}

// Translate translates the key to the current language of the context
func Translate(ctx *GinContext, key string) string {
	return translator.Tr(handler.GetLang(ctx), key)
}

// TranslateWithData translates the key to the language with data
func TranslateWithData(lang i18n.Language, key string, data any) string {
	return translator.TrWithData(lang, key, data)
}

// TranslateFn presents a generator of translated string.
// We use it to delegate the translation work outside the plugin.
type TranslateFn func(ctx *GinContext) string

// Translator contains a function that translates the key to the current language of the context
type Translator struct {
	Fn TranslateFn
}

// MakeTranslator generates a translator from the key
func MakeTranslator(key string) Translator {
	t := func(ctx *GinContext) string {
		return Translate(ctx, key)
	}
	return Translator{Fn: t}
}

// Translate translates the key to the current language of the context
func (t Translator) Translate(ctx *GinContext) string {
	if t.Fn == nil {
		return ""
	}
	return t.Fn(ctx)
}

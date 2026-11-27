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

package migrations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func providerHosts(t *testing.T, value string) map[string]string {
	t.Helper()
	providers := []map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(value), &providers))
	hosts := make(map[string]string)
	for _, p := range providers {
		name, _ := p["name"].(string)
		host, _ := p["default_api_host"].(string)
		hosts[name] = host
	}
	return hosts
}

func TestFixGeminiProviderList(t *testing.T) {
	legacy := []map[string]any{
		{"name": "openai", "display_name": "OpenAI", "default_api_host": "https://api.openai.com"},
		{"name": "gemini", "display_name": "Gemini", "default_api_host": legacyGeminiAPIHost},
		{"name": "anthropic", "display_name": "Anthropic", "default_api_host": "https://api.anthropic.com"},
	}
	assert.True(t, fixGeminiProviderList(legacy))
	data, err := json.Marshal(legacy)
	require.NoError(t, err)
	hosts := providerHosts(t, string(data))
	assert.Equal(t, geminiAPIHost, hosts["gemini"])
	assert.Equal(t, "https://api.openai.com", hosts["openai"])
	assert.Equal(t, "https://api.anthropic.com", hosts["anthropic"])
}

func TestFixGeminiProviderListKeepsCustomizedHost(t *testing.T) {
	custom := []map[string]any{
		{"name": "gemini", "display_name": "Gemini", "default_api_host": "https://gemini.example.com/proxy"},
	}
	assert.False(t, fixGeminiProviderList(custom))
	assert.Equal(t, "https://gemini.example.com/proxy", custom[0]["default_api_host"])
}

func TestFixGeminiProviderListSkipsUpToDate(t *testing.T) {
	upToDate := []map[string]any{
		{"name": "gemini", "display_name": "Gemini", "default_api_host": geminiAPIHost},
	}
	assert.False(t, fixGeminiProviderList(upToDate))
}

func TestFixGeminiSiteAIContent(t *testing.T) {
	legacyContent := `{"enabled":true,"chosen_provider":"gemini","ai_providers":[` +
		`{"provider":"openai","api_host":"https://api.openai.com"},` +
		`{"provider":"gemini","api_host":"https://generativelanguage.googleapis.com","api_key":"k*"}],` +
		`"prompt_config":{"zh_cn":"z","en_us":"e"}}`
	updated, changed, err := fixGeminiSiteAIContent(legacyContent)
	require.NoError(t, err)
	assert.True(t, changed)

	parsed := struct {
		ChosenProvider string `json:"chosen_provider"`
		AIProviders    []struct {
			Provider string `json:"provider"`
			APIHost  string `json:"api_host"`
		} `json:"ai_providers"`
	}{}
	require.NoError(t, json.Unmarshal([]byte(updated), &parsed))
	assert.Equal(t, "gemini", parsed.ChosenProvider)
	require.Len(t, parsed.AIProviders, 2)
	assert.Equal(t, "https://api.openai.com", parsed.AIProviders[0].APIHost)
	assert.Equal(t, geminiAPIHost, parsed.AIProviders[1].APIHost)
}

func TestFixGeminiSiteAIContentKeepsCustomizedHost(t *testing.T) {
	customContent := `{"enabled":true,"chosen_provider":"gemini","ai_providers":[` +
		`{"provider":"gemini","api_host":"https://gemini.example.com/proxy"}]}`
	_, changed, err := fixGeminiSiteAIContent(customContent)
	require.NoError(t, err)
	assert.False(t, changed)
}

func TestFixGeminiSiteAIContentInvalidJSON(t *testing.T) {
	_, changed, err := fixGeminiSiteAIContent("not json")
	assert.Error(t, err)
	assert.False(t, changed)
}

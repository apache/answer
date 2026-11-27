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

package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// isHostOrSubdomain reports whether host equals base or lives under it,
// e.g. ws-example.cn-beijing.maas.aliyuncs.com under maas.aliyuncs.com.
// Both arguments must already be lowercase hostnames without a port.
func isHostOrSubdomain(host, base string) bool {
	return host == base || strings.HasSuffix(host, "."+base)
}

// thinkingParamForHost returns the request body patch that enables thinking
// for a provider host, or nil when the host has no documented
// OpenAI-compatible thinking parameter. The flag is gateway-specific rather
// than a neutral OpenAI extension:
//   - DashScope/Qwen compatibility mode documents top-level "enable_thinking"
//     (https://help.aliyun.com/en/model-studio/qwen-api-via-dashscope). It
//     applies to the dashscope.aliyuncs.com and dashscope-intl.aliyuncs.com
//     gateways as well as the regional Model Studio endpoints served as
//     https://{workspace}.{region}.maas.aliyuncs.com/compatible-mode/v1.
//   - DeepSeek documents an object parameter, thinking: {"type": "enabled"}
//     (https://api-docs.deepseek.com/guides/thinking_mode/).
//   - Gemini's OpenAI-compatible endpoint accepts neither.
//
// Hosts are matched on the parsed URL hostname (scheme-less values are parsed
// as https), never on substrings, so a gateway name embedded in a path, query
// or look-alike domain cannot trigger injection. Unknown hosts receive no
// provider-specific field instead of a global guess.
func thinkingParamForHost(apiHost string) map[string]any {
	raw := strings.TrimSpace(apiHost)
	if raw == "" {
		return nil
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil
	}
	switch {
	case host == "api.deepseek.com":
		return map[string]any{"thinking": map[string]any{"type": "enabled"}}
	case isHostOrSubdomain(host, "dashscope.aliyuncs.com"),
		isHostOrSubdomain(host, "dashscope-intl.aliyuncs.com"),
		isHostOrSubdomain(host, "maas.aliyuncs.com"):
		return map[string]any{"enable_thinking": true}
	default:
		return nil
	}
}

// thinkingTransport merges the thinking parameter into chat completion request
// bodies before they leave the process. The openai SDK has no generic extra
// body hook, so an http.Client with this transport is attached to the client
// config when the provider enables thinking mode and its host documents a
// thinking parameter.
type thinkingTransport struct {
	base  http.RoundTripper
	param map[string]any
}

func (t *thinkingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.param) > 0 && req.Body != nil && req.ContentLength != 0 &&
		strings.HasSuffix(req.URL.Path, "/chat/completions") {
		b, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err == nil {
			var payload map[string]any
			if json.Unmarshal(b, &payload) == nil && payload != nil {
				for k, v := range t.param {
					payload[k] = v
				}
				nb, mErr := json.Marshal(payload)
				if mErr == nil {
					req.Body = io.NopCloser(bytes.NewReader(nb))
					req.ContentLength = int64(len(nb))
					req.Header.Set("Content-Length", strconv.Itoa(len(nb)))
				} else {
					req.Body = io.NopCloser(bytes.NewReader(b))
				}
			} else {
				req.Body = io.NopCloser(bytes.NewReader(b))
			}
		}
	}
	if t.base == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.base.RoundTrip(req)
}

func newThinkingHTTPClient(param map[string]any) *http.Client {
	return &http.Client{Transport: &thinkingTransport{param: param}}
}

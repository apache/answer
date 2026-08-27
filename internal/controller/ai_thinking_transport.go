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
	"strconv"
	"strings"
)

// ThinkingParamKey is the OpenAI-compatible request flag understood by
// reasoning-capable model gateways (DeepSeek, Qwen/DashScope, SenseNova, ...).
const ThinkingParamKey = "enable_thinking"

// thinkingTransport merges the thinking flag into chat completion request
// bodies before they leave the process. The openai SDK has no generic extra
// body hook, so an http.Client with this transport is attached to the client
// config when the provider enables thinking mode.
type thinkingTransport struct{ base http.RoundTripper }

func (t *thinkingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.ContentLength != 0 &&
		strings.HasSuffix(req.URL.Path, "/chat/completions") {
		b, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err == nil {
			var payload map[string]any
			if json.Unmarshal(b, &payload) == nil && payload != nil {
				payload[ThinkingParamKey] = true
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

func newThinkingHTTPClient() *http.Client {
	return &http.Client{Transport: &thinkingTransport{}}
}

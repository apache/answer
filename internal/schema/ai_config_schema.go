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

package schema

import "strings"

// NormalizeAPIHost normalizes an OpenAI-compatible API base host:
// trims whitespace/slashes and appends "/v1" unless already versioned,
// so that hosts with or without the "/v1" suffix resolve identically.
func NormalizeAPIHost(host string) string {
	h := strings.TrimSpace(host)
	h = strings.TrimRight(h, "/")
	if h == "" {
		return ""
	}
	if strings.HasSuffix(h, "/v1") || strings.Contains(h, "/v1beta") {
		return h
	}
	return h + "/v1"
}

// GetAIProviderResp get AI providers response
type GetAIProviderResp struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	DefaultAPIHost string `json:"default_api_host"`
}

// GetAIModelsResp get AI model response
type GetAIModelsResp struct {
	Object string `json:"object"`
	Data   []struct {
		Id      string `json:"id"`
		Object  string `json:"object"`
		Created int    `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

type GetAIModelsReq struct {
	APIHost string `json:"api_host"`
	APIKey  string `json:"api_key"`
}

// GetAIModelResp get AI model response
type GetAIModelResp struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	OwnedBy string `json:"owned_by"`
}

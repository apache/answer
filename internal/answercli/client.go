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

package answercli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrInsecureHTTP = errors.New("refusing to send a personal access token over insecure HTTP")

type APIError struct {
	HTTPStatus     int
	Reason         string
	Message        string
	Data           json.RawMessage
	OutcomeUnknown bool
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Reason
}

type apiResponse struct {
	Code   int             `json:"code"`
	Reason string          `json:"reason"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

type Client struct {
	server string
	token  string
	http   *http.Client
}

func NewClient(profile Profile, httpClient *http.Client) (*Client, error) {
	if err := ValidateServerURL(profile.Server, profile.AllowInsecureHTTP); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{server: strings.TrimRight(profile.Server, "/"), token: profile.Token, http: httpClient}, nil
}

func ValidateServerURL(server string, allowInsecure bool) error {
	parsed, err := url.Parse(server)
	if err != nil || parsed.Hostname() == "" {
		return fmt.Errorf("invalid Answer server URL")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme != "http" {
		return fmt.Errorf("unsupported Answer server URL scheme %q", parsed.Scheme)
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if host == "localhost" || (ip != nil && ip.IsLoopback()) || allowInsecure {
		return nil
	}
	return ErrInsecureHTTP
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	var bodyContent []byte
	if body != nil {
		content, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyContent = content
	}
	endpoint := c.server + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	attempts := 1
	if method == http.MethodGet {
		attempts = 3
	}
	for attempt := 0; attempt < attempts; attempt++ {
		result, retry, err := c.doOnce(ctx, method, endpoint, bodyContent)
		if err == nil || !retry || attempt == attempts-1 {
			return result, err
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}
	return nil, nil
}

func (c *Client) doOnce(ctx context.Context, method, endpoint string, bodyContent []byte) (json.RawMessage, bool, error) {
	var requestBody io.Reader
	if bodyContent != nil {
		requestBody = bytes.NewReader(bodyContent)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return nil, false, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/json")
	if bodyContent != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, method == http.MethodGet, &APIError{Message: err.Error(), OutcomeUnknown: method != http.MethodGet}
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, method == http.MethodGet, err
	}
	result := &apiResponse{}
	if err := json.Unmarshal(content, result); err != nil {
		return nil, false, fmt.Errorf("decode Answer response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retry := method == http.MethodGet && (response.StatusCode == http.StatusBadGateway ||
			response.StatusCode == http.StatusServiceUnavailable || response.StatusCode == http.StatusGatewayTimeout)
		return nil, retry, &APIError{HTTPStatus: response.StatusCode, Reason: result.Reason, Message: result.Msg, Data: result.Data}
	}
	if len(result.Data) == 0 {
		return json.RawMessage("null"), false, nil
	}
	return result.Data, false, nil
}

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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientRetriesReadsButNeverWrites(t *testing.T) {
	readTransport := &sequenceTransport{failures: 2}
	client, err := NewClient(Profile{Server: "https://answer.example.com", Token: "answer_pat_secret"}, &http.Client{Transport: readTransport})
	require.NoError(t, err)
	_, err = client.Do(t.Context(), http.MethodGet, "/answer/api/v1/question/info", nil, nil)
	require.NoError(t, err)
	require.Equal(t, 3, readTransport.calls)

	writeTransport := &sequenceTransport{failures: 2}
	client, err = NewClient(Profile{Server: "https://answer.example.com", Token: "answer_pat_secret"}, &http.Client{Transport: writeTransport})
	require.NoError(t, err)
	_, err = client.Do(t.Context(), http.MethodPost, "/answer/api/v1/question", nil, map[string]string{"title": "question"})
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.True(t, apiErr.OutcomeUnknown)
	require.Equal(t, 1, writeTransport.calls)
}

type sequenceTransport struct {
	calls    int
	failures int
}

func (t *sequenceTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls++
	if t.calls <= t.failures {
		return nil, errors.New("temporary network failure")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"code":200,"data":{}}`)),
	}, nil
}

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

package search_parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUserID_UserMe(t *testing.T) {
	sp := &SearchParser{}
	ctx := context.Background()

	t.Run("parse user:me with logged in user", func(t *testing.T) {
		q := "user:me hello world"
		userID := sp.parseUserID(ctx, &q, "user-123")
		assert.Equal(t, "user-123", userID)
		assert.Equal(t, "hello world", q)
	})

	t.Run("parse user:me when signed out", func(t *testing.T) {
		q := "user:me hello world"
		userID := sp.parseUserID(ctx, &q, "")
		assert.Equal(t, "", userID)
		assert.Equal(t, "hello world", q)
	})

	t.Run("username starting with me should not be parsed as user:me", func(t *testing.T) {
		q := "user:melissa service"
		// without userCommon returning a user, it shouldn't match user:me
		userID := sp.parseUserID(ctx, &q, "user-123")
		assert.NotEqual(t, "user-123", userID)
	})
}

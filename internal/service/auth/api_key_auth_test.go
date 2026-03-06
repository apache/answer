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

package auth

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockAPIKeyRepo struct {
	key   *entity.APIKey
	exist bool
	err   error
}

func (m *mockAPIKeyRepo) GetAPIKeyList(_ context.Context) ([]*entity.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyRepo) GetAPIKey(_ context.Context, _ string) (*entity.APIKey, bool, error) {
	return m.key, m.exist, m.err
}
func (m *mockAPIKeyRepo) UpdateAPIKey(_ context.Context, _ entity.APIKey) error { return nil }
func (m *mockAPIKeyRepo) AddAPIKey(_ context.Context, _ entity.APIKey) error    { return nil }
func (m *mockAPIKeyRepo) DeleteAPIKey(_ context.Context, _ int) error           { return nil }

// --- Tests ---

func TestGetAPIKeyInfo_ValidReadWriteKey(t *testing.T) {
	svc := &AuthService{
		apiKeyRepo: &mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_test", Scope: "read-write", UserID: "100"},
			exist: true,
		},
	}

	info, err := svc.GetAPIKeyInfo(context.Background(), false, "sk_test")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "100", info.UserID)
	assert.Equal(t, "read-write", info.Scope)
}

func TestGetAPIKeyInfo_ReadOnlyKeyGetRequest(t *testing.T) {
	svc := &AuthService{
		apiKeyRepo: &mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_ro", Scope: "read-only", UserID: "200"},
			exist: true,
		},
	}

	info, err := svc.GetAPIKeyInfo(context.Background(), true, "sk_ro")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "200", info.UserID)
}

func TestGetAPIKeyInfo_ReadOnlyKeyNonGetRequest(t *testing.T) {
	svc := &AuthService{
		apiKeyRepo: &mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_ro", Scope: "read-only", UserID: "200"},
			exist: true,
		},
	}

	info, err := svc.GetAPIKeyInfo(context.Background(), false, "sk_ro")
	assert.NoError(t, err)
	assert.Nil(t, info)
}

func TestGetAPIKeyInfo_KeyNotFound(t *testing.T) {
	svc := &AuthService{
		apiKeyRepo: &mockAPIKeyRepo{exist: false},
	}

	info, err := svc.GetAPIKeyInfo(context.Background(), true, "sk_invalid")
	assert.NoError(t, err)
	assert.Nil(t, info)
}

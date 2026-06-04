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
	"context"
	"encoding/json"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestBackfillRequireEmailVerification(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "adds true for missing key",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true}`,
			expected: true,
		},
		{
			name:     "converts null to true",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":null}`,
			expected: true,
		},
		{
			name:     "preserves false",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":false}`,
			expected: false,
		},
		{
			name:     "preserves true",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":true}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := backfillRequireEmailVerification(tt.content)
			require.NoError(t, err)

			var result map[string]bool
			require.NoError(t, json.Unmarshal([]byte(content), &result))
			assert.Equal(t, tt.expected, result["require_email_verification"])
		})
	}
}

func TestAddRequireEmailVerificationAbsentLoginRow(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(entity.SiteInfo)))

	require.NoError(t, addRequireEmailVerification(context.TODO(), x))
}

func TestAddRequireEmailVerificationUpdatesLoginRow(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(entity.SiteInfo)))

	_, err = x.Insert(&entity.SiteInfo{
		Type:    constant.SiteTypeLogin,
		Content: `{"allow_new_registrations":true}`,
		Status:  1,
	})
	require.NoError(t, err)

	require.NoError(t, addRequireEmailVerification(context.TODO(), x))

	login := &entity.SiteInfo{}
	exist, err := x.Where("type = ?", constant.SiteTypeLogin).Get(login)
	require.NoError(t, err)
	require.True(t, exist)

	var result struct {
		RequireEmailVerification bool `json:"require_email_verification"`
	}
	require.NoError(t, json.Unmarshal([]byte(login.Content), &result))
	assert.True(t, result.RequireEmailVerification)
}

func TestSplitLegalMenuKeepsRequireEmailVerificationDefaultTrue(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(entity.SiteInfo)))

	_, err = x.Insert(&entity.SiteInfo{
		Type: constant.SiteTypeLegal,
		Content: `{
			"terms_of_service_original_text":"tos",
			"terms_of_service_parsed_text":"tos",
			"privacy_policy_original_text":"privacy",
			"privacy_policy_parsed_text":"privacy",
			"external_content_display":"always_display"
		}`,
		Status: 1,
	})
	require.NoError(t, err)
	_, err = x.Insert(&entity.SiteInfo{
		Type: constant.SiteTypeLogin,
		Content: `{
			"allow_new_registrations":true,
			"allow_email_registrations":true,
			"allow_password_login":true,
			"login_required":false,
			"allow_email_domains":[]
		}`,
		Status: 1,
	})
	require.NoError(t, err)
	_, err = x.Insert(&entity.SiteInfo{
		Type: constant.SiteTypeGeneral,
		Content: `{
			"name":"site",
			"short_description":"short",
			"description":"description",
			"site_url":"https://example.com",
			"contact_email":"admin@example.com",
			"check_update":true
		}`,
		Status: 1,
	})
	require.NoError(t, err)

	require.NoError(t, splitLegalMenu(context.TODO(), x))

	login := &entity.SiteInfo{}
	exist, err := x.Where("type = ?", constant.SiteTypeLogin).Get(login)
	require.NoError(t, err)
	require.True(t, exist)

	var result struct {
		RequireEmailVerification bool `json:"require_email_verification"`
	}
	require.NoError(t, json.Unmarshal([]byte(login.Content), &result))
	assert.True(t, result.RequireEmailVerification)
}

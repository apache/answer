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

package siteinfo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSiteInfoService_SaveSiteLoginRequireEmailVerification(t *testing.T) {
	tests := []struct {
		name            string
		requireEmail    bool
		expectedRequire bool
	}{
		{
			name:            "explicit true persists true",
			requireEmail:    true,
			expectedRequire: true,
		},
		{
			name:            "explicit false persists false",
			requireEmail:    false,
			expectedRequire: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mock.NewMockSiteInfoRepo(ctl)
			var savedContent string
			repo.EXPECT().SaveByType(gomock.Any(), constant.SiteTypeLogin, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, data *entity.SiteInfo) error {
					savedContent = data.Content
					return nil
				})

			service := &SiteInfoService{
				siteInfoRepo: repo,
			}
			req := &schema.SiteLoginReq{
				AllowNewRegistrations:    true,
				AllowEmailRegistrations:  true,
				AllowPasswordLogin:       true,
				AllowEmailDomains:        []string{},
				RequireEmailVerification: &tt.requireEmail,
			}
			require.NoError(t, service.SaveSiteLogin(context.TODO(), req))
			assert.NotContains(t, savedContent, `"require_email_verification":null`)

			saved := &schema.SiteLoginResp{}
			require.NoError(t, json.Unmarshal([]byte(savedContent), saved))
			assert.Equal(t, tt.expectedRequire, saved.RequireEmailVerification)
		})
	}
}

func TestSiteInfoService_SaveSiteLoginRequiresEmailVerificationValue(t *testing.T) {
	service := &SiteInfoService{}
	req := &schema.SiteLoginReq{
		AllowNewRegistrations:   true,
		AllowEmailRegistrations: true,
		AllowPasswordLogin:      true,
		AllowEmailDomains:       []string{},
	}

	require.Error(t, service.SaveSiteLogin(context.TODO(), req))
}

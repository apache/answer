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
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSiteInfoService_SaveSiteLoginRequireEmailVerification(t *testing.T) {
	tests := []struct {
		name            string
		currentContent  string
		requestPayload  string
		expectGet       bool
		expectedRequire bool
	}{
		{
			name:            "omitted preserves normalized default",
			currentContent:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true}`,
			requestPayload:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[]}`,
			expectGet:       true,
			expectedRequire: true,
		},
		{
			name:            "omitted preserves current false",
			currentContent:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":false}`,
			requestPayload:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[]}`,
			expectGet:       true,
			expectedRequire: false,
		},
		{
			name:            "null normalizes true",
			requestPayload:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[],"require_email_verification":null}`,
			expectedRequire: true,
		},
		{
			name:            "explicit false persists false",
			requestPayload:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[],"require_email_verification":false}`,
			expectedRequire: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mock.NewMockSiteInfoRepo(ctl)
			if tt.expectGet {
				repo.EXPECT().GetByType(gomock.Any(), constant.SiteTypeLogin).
					Return(&entity.SiteInfo{Content: tt.currentContent}, true, nil)
			}

			var savedContent string
			repo.EXPECT().SaveByType(gomock.Any(), constant.SiteTypeLogin, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, data *entity.SiteInfo) error {
					savedContent = data.Content
					return nil
				})

			req := &schema.SiteLoginReq{}
			require.NoError(t, json.Unmarshal([]byte(tt.requestPayload), req))

			service := &SiteInfoService{
				siteInfoRepo:          repo,
				siteInfoCommonService: siteinfo_common.NewSiteInfoCommonService(repo),
			}
			require.NoError(t, service.SaveSiteLogin(context.TODO(), req))
			assert.NotContains(t, savedContent, `"require_email_verification":null`)

			saved := &schema.SiteLoginResp{}
			require.NoError(t, json.Unmarshal([]byte(savedContent), saved))
			assert.Equal(t, tt.expectedRequire, saved.RequireEmailVerification)
		})
	}
}

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

package siteinfo_common

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	mockSiteInfoRepo *mock.MockSiteInfoRepo
)

func mockInit(ctl *gomock.Controller) {
	mockSiteInfoRepo = mock.NewMockSiteInfoRepo(ctl)
	mockSiteInfoRepo.EXPECT().GetByType(gomock.Any(), constant.SiteTypeGeneral).
		Return(&entity.SiteInfo{Content: `{"name":"name"}`}, true, nil)
}

func TestSiteInfoCommonService_GetSiteGeneral(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockInit(ctl)
	siteInfoCommonService := NewSiteInfoCommonService(mockSiteInfoRepo)
	resp, err := siteInfoCommonService.GetSiteGeneral(context.TODO())
	require.NoError(t, err)
	assert.Equal(t, "name", resp.Name)
}

func TestSiteInfoCommonService_GetSiteLoginRequireEmailVerification(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "missing key defaults true",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true}`,
			expected: true,
		},
		{
			name:     "null defaults true",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":null}`,
			expected: true,
		},
		{
			name:     "explicit false is preserved",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":false}`,
			expected: false,
		},
		{
			name:     "explicit true is preserved",
			content:  `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"require_email_verification":true}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			repo := mock.NewMockSiteInfoRepo(ctl)
			repo.EXPECT().GetByType(gomock.Any(), constant.SiteTypeLogin).
				Return(&entity.SiteInfo{Content: tt.content}, true, nil)

			siteInfoCommonService := NewSiteInfoCommonService(repo)
			resp, err := siteInfoCommonService.GetSiteLogin(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, resp.RequireEmailVerification)
		})
	}
}

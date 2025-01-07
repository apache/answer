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
	assert.NoError(t, err)
	assert.Equal(t, resp.Name, "name")
}

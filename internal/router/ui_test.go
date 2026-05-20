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

package router

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/answer/internal/service/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUIRouter_FaviconWithNilBranding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSiteInfoService := mock.NewMockSiteInfoCommonService(ctrl)

	// Simulate a database error
	mockSiteInfoService.EXPECT().
		GetSiteBranding(gomock.Any()).
		Return(nil, errors.New("database connection failed"))

	router := &UIRouter{
		siteInfoService: mockSiteInfoService,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.Register(r, "")

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

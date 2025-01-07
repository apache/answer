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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/site_info"
	"github.com/stretchr/testify/assert"
)

func Test_siteInfoRepo_SaveByType(t *testing.T) {
	siteInfoRepo := site_info.NewSiteInfo(testDataSource)

	data := &entity.SiteInfo{Content: "site_info", Type: "test"}

	err := siteInfoRepo.SaveByType(context.TODO(), data.Type, data)
	assert.NoError(t, err)

	got, exist, err := siteInfoRepo.GetByType(context.TODO(), data.Type)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, data.Content, got.Content)

	data.Content = "new site_info"
	err = siteInfoRepo.SaveByType(context.TODO(), data.Type, data)
	assert.NoError(t, err)

	got, exist, err = siteInfoRepo.GetByType(context.TODO(), data.Type)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, data.Content, got.Content)
}

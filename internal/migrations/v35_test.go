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
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestRepairAdvancedSiteInfoAddsMissingSettings(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(entity.SiteInfo)))

	_, err = x.Insert(&entity.SiteInfo{
		Type:    constant.SiteTypeWrite,
		Content: `{"max_image_size":5}`,
		Status:  1,
	})
	require.NoError(t, err)

	var repairMigration Migration
	for _, m := range GetMigrations() {
		if m.Version() == "v2.0.4" {
			repairMigration = m
			break
		}
	}
	require.NotNil(t, repairMigration)
	require.NoError(t, repairMigration.Migrate(context.Background(), x))

	advanced := &entity.SiteInfo{}
	exists, err := x.Where("type = ?", constant.SiteTypeAdvanced).Get(advanced)
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, `{
		"max_image_size": 5,
		"max_attachment_size": 0,
		"max_image_megapixel": 0,
		"authorized_image_extensions": null,
		"authorized_attachment_extensions": null
	}`, advanced.Content)
}

func TestRepairAdvancedSiteInfoPreservesExistingSettings(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(entity.SiteInfo)))

	const existingContent = `{"max_image_size":99}`
	_, err = x.Insert(
		&entity.SiteInfo{
			Type:    constant.SiteTypeWrite,
			Content: `{invalid`,
			Status:  1,
		},
		&entity.SiteInfo{
			Type:    constant.SiteTypeAdvanced,
			Content: existingContent,
			Status:  1,
		},
	)
	require.NoError(t, err)

	require.NoError(t, repairAdvancedSiteInfo(context.Background(), x))

	advanced := &entity.SiteInfo{}
	exists, err := x.Where("type = ?", constant.SiteTypeAdvanced).Get(advanced)
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, existingContent, advanced.Content)
}

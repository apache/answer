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

package forum_section

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/require"
)

type testRepo struct {
	sections []*entity.ForumSection
}

func (r *testRepo) List(context.Context) ([]*entity.ForumSection, error) {
	return r.sections, nil
}

func (r *testRepo) GetByID(_ context.Context, id int64) (*entity.ForumSection, bool, error) {
	for _, section := range r.sections {
		if section.ID == id {
			return section, true, nil
		}
	}
	return &entity.ForumSection{}, false, nil
}

func (r *testRepo) GetBySlug(_ context.Context, slug string) (*entity.ForumSection, bool, error) {
	for _, section := range r.sections {
		if section.Slug == slug {
			return section, true, nil
		}
	}
	return &entity.ForumSection{}, false, nil
}

func testSections() []*entity.ForumSection {
	return []*entity.ForumSection{
		{ID: 1, Slug: "site-management", Name: "站务管理", Status: 1},
		{ID: 101, ParentID: 1, Slug: "site-announcements", Name: "站务公告", AdminOnly: true, Status: 1},
		{ID: 102, ParentID: 1, Slug: "part-time-jobs", Name: "兼职信息中心", Status: 1},
	}
}

func TestListTreeAndResolveParentSection(t *testing.T) {
	service := NewForumSectionService(&testRepo{sections: testSections()})

	tree, err := service.ListTree(context.Background())
	require.NoError(t, err)
	require.Len(t, tree, 1)
	require.Len(t, tree[0].Children, 2)

	ids, exists, err := service.ResolveSectionIDs(context.Background(), "site-management")
	require.NoError(t, err)
	require.True(t, exists)
	require.ElementsMatch(t, []int64{101, 102}, ids)
}

func TestValidatePostingSectionPermissions(t *testing.T) {
	service := NewForumSectionService(&testRepo{sections: testSections()})
	ctx := context.Background()

	_, err := service.ValidatePostingSection(ctx, 101, false)
	require.Error(t, err)

	_, err = service.ValidatePostingSection(ctx, 101, true)
	require.NoError(t, err)

	_, err = service.ValidatePostingSection(ctx, 102, false)
	require.NoError(t, err)

	_, err = service.ValidatePostingSection(ctx, 1, true)
	require.Error(t, err)
}

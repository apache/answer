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
	"strings"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	forumsectionservice "github.com/apache/answer/internal/service/forum_section"
	"github.com/segmentfault/pacman/errors"
)

type repo struct{ data *data.Data }

func NewForumSectionRepo(dataSource *data.Data) forumsectionservice.Repo {
	return &repo{data: dataSource}
}

func (r *repo) List(ctx context.Context) ([]*entity.ForumSection, error) {
	list := make([]*entity.ForumSection, 0)
	err := r.data.DB.Context(ctx).Where("status = ?", entity.ForumSectionStatusAvailable).
		Asc("sort", "id").Find(&list)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return list, nil
}

func (r *repo) GetByID(ctx context.Context, id int64) (*entity.ForumSection, bool, error) {
	section := &entity.ForumSection{ID: id}
	exists, err := r.data.DB.Context(ctx).Get(section)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return section, exists, nil
}

func (r *repo) GetBySlug(ctx context.Context, slug string) (*entity.ForumSection, bool, error) {
	section := &entity.ForumSection{}
	exists, err := r.data.DB.Context(ctx).Where("slug = ?", strings.ToLower(strings.TrimSpace(slug))).Get(section)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return section, exists, nil
}

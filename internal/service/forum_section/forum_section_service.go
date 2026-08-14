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

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/errors"
)

type Repo interface {
	List(ctx context.Context) ([]*entity.ForumSection, error)
	GetByID(ctx context.Context, id int64) (*entity.ForumSection, bool, error)
	GetBySlug(ctx context.Context, slug string) (*entity.ForumSection, bool, error)
}

type Service struct {
	repo Repo
}

func NewForumSectionService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTree(ctx context.Context) ([]*schema.ForumSectionResp, error) {
	sections, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	parents := make([]*schema.ForumSectionResp, 0)
	parentMap := make(map[int64]*schema.ForumSectionResp)
	for _, section := range sections {
		item := toResp(section)
		if section.ParentID == 0 {
			parents = append(parents, item)
			parentMap[section.ID] = item
		}
	}
	for _, section := range sections {
		if section.ParentID == 0 {
			continue
		}
		if parent := parentMap[section.ParentID]; parent != nil {
			parent.Children = append(parent.Children, toResp(section))
		}
	}
	return parents, nil
}

func (s *Service) ValidatePostingSection(ctx context.Context, id int64, isAdmin bool) (*entity.ForumSection, error) {
	section, exists, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists || section.Status != entity.ForumSectionStatusAvailable || section.ParentID == 0 {
		return nil, errors.BadRequest(reason.ForumSectionInvalid)
	}
	if section.AdminOnly && !isAdmin {
		return nil, errors.Forbidden(reason.ForumSectionAdminOnly)
	}
	return section, nil
}

func (s *Service) ResolveSectionIDs(ctx context.Context, slug string) ([]int64, bool, error) {
	section, exists, err := s.repo.GetBySlug(ctx, slug)
	if err != nil || !exists {
		return nil, exists, err
	}
	if section.ParentID != 0 {
		return []int64{section.ID}, true, nil
	}
	sections, err := s.repo.List(ctx)
	if err != nil {
		return nil, false, err
	}
	ids := make([]int64, 0)
	for _, child := range sections {
		if child.ParentID == section.ID {
			ids = append(ids, child.ID)
		}
	}
	return ids, true, nil
}

func toResp(section *entity.ForumSection) *schema.ForumSectionResp {
	return &schema.ForumSectionResp{
		ID: section.ID, ParentID: section.ParentID, Slug: section.Slug,
		Name: section.Name, AdminOnly: section.AdminOnly, Children: make([]*schema.ForumSectionResp, 0),
	}
}

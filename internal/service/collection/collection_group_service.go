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

package collection

import (
	"context"

	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
)

// CollectionGroupRepo collectionGroup repository
type CollectionGroupRepo interface {
	AddCollectionGroup(ctx context.Context, collectionGroup *entity.CollectionGroup) (err error)
	AddCollectionDefaultGroup(ctx context.Context, userID string) (collectionGroup *entity.CollectionGroup, err error)
	CreateDefaultGroupIfNotExist(ctx context.Context, userID string) (collectionGroup *entity.CollectionGroup, err error)
	UpdateCollectionGroup(ctx context.Context, collectionGroup *entity.CollectionGroup, cols []string) (err error)
	GetCollectionGroup(ctx context.Context, id string) (collectionGroup *entity.CollectionGroup, exist bool, err error)
	GetCollectionGroupPage(ctx context.Context, page, pageSize int, collectionGroup *entity.CollectionGroup) (collectionGroupList []*entity.CollectionGroup, total int64, err error)
	GetDefaultID(ctx context.Context, userID string) (collectionGroup *entity.CollectionGroup, has bool, err error)
}

// CollectionGroupService user service
type CollectionGroupService struct {
	collectionGroupRepo CollectionGroupRepo
}

func NewCollectionGroupService(collectionGroupRepo CollectionGroupRepo) *CollectionGroupService {
	return &CollectionGroupService{
		collectionGroupRepo: collectionGroupRepo,
	}
}

// AddCollectionGroup add collection group
func (cs *CollectionGroupService) AddCollectionGroup(ctx context.Context, req *schema.AddCollectionGroupReq) (err error) {
	collectionGroup := &entity.CollectionGroup{}
	_ = copier.Copy(collectionGroup, req)
	return cs.collectionGroupRepo.AddCollectionGroup(ctx, collectionGroup)
}

// UpdateCollectionGroup update collection group
func (cs *CollectionGroupService) UpdateCollectionGroup(ctx context.Context, req *schema.UpdateCollectionGroupReq, cols []string) (err error) {
	collectionGroup := &entity.CollectionGroup{}
	_ = copier.Copy(collectionGroup, req)
	return cs.collectionGroupRepo.UpdateCollectionGroup(ctx, collectionGroup, cols)
}

// GetCollectionGroup get collection group one
func (cs *CollectionGroupService) GetCollectionGroup(ctx context.Context, id string) (resp *schema.GetCollectionGroupResp, err error) {
	collectionGroup, exist, err := cs.collectionGroupRepo.GetCollectionGroup(ctx, id)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.UnknownError)
	}

	resp = &schema.GetCollectionGroupResp{}
	_ = copier.Copy(resp, collectionGroup)
	return resp, nil
}

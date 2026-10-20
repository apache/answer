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

package follow

import (
	"context"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_common"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	"github.com/apache/answer/pkg/obj"
	"github.com/segmentfault/pacman/errors"
)

type FollowRepo interface {
	Follow(ctx context.Context, objectId, userId string) error
	FollowCancel(ctx context.Context, objectId, userId string) error
}

type ObjectInfoService interface {
	GetInfo(ctx context.Context, objectID string) (*schema.SimpleObjectInfo, error)
}

type FollowService struct {
	tagRepo           tagcommon.TagCommonRepo
	followRepo        FollowRepo
	followCommonRepo  activity_common.FollowRepo
	objectInfoService ObjectInfoService
}

func NewFollowService(
	followRepo FollowRepo,
	followCommonRepo activity_common.FollowRepo,
	tagRepo tagcommon.TagCommonRepo,
	objectInfoService ObjectInfoService,
) *FollowService {
	return &FollowService{
		followRepo:        followRepo,
		followCommonRepo:  followCommonRepo,
		tagRepo:           tagRepo,
		objectInfoService: objectInfoService,
	}
}

// Follow or cancel follow object
func (fs *FollowService) Follow(ctx context.Context, dto *schema.FollowDTO) (resp schema.FollowResp, err error) {
	if dto.IsCancel {
		err = fs.followRepo.FollowCancel(ctx, dto.ObjectID, dto.UserID)
	} else {
		if err := fs.checkVisibility(ctx, dto); err != nil {
			return resp, err
		}
		err = fs.followRepo.Follow(ctx, dto.ObjectID, dto.UserID)
	}
	if err != nil {
		return resp, err
	}
	follows, err := fs.followCommonRepo.GetFollowAmount(ctx, dto.ObjectID)
	if err != nil {
		return resp, err
	}

	resp.Follows = follows
	resp.IsFollowed = !dto.IsCancel
	return resp, nil
}

func (fs *FollowService) checkVisibility(ctx context.Context, dto *schema.FollowDTO) error {
	objectType, err := obj.GetObjectTypeStrByObjectID(dto.ObjectID)
	if err != nil {
		return err
	}
	if objectType != constant.QuestionObjectType && objectType != constant.AnswerObjectType && objectType != constant.CommentObjectType {
		return nil
	}
	objectInfo, err := fs.objectInfoService.GetInfo(ctx, dto.ObjectID)
	if err != nil {
		if serviceErr, ok := err.(*errors.Error); ok && serviceErr.Reason == reason.ObjectNotFound {
			return errors.NotFound(reason.QuestionNotFound)
		}
		return err
	}
	return objectInfo.CheckVisibility(dto.UserID, dto.IsAdminModerator)
}

// UpdateFollowTags update user follow tags
func (fs *FollowService) UpdateFollowTags(ctx context.Context, req *schema.UpdateFollowTagsReq) (err error) {
	objIDs, err := fs.followCommonRepo.GetFollowIDs(ctx, req.UserID, entity.Tag{}.TableName())
	if err != nil {
		return
	}
	oldFollowTagList, err := fs.tagRepo.GetTagListByIDs(ctx, objIDs)
	if err != nil {
		return err
	}
	oldTagMapping := make(map[string]bool)
	for _, tag := range oldFollowTagList {
		oldTagMapping[tag.SlugName] = true
	}

	newTagMapping := make(map[string]bool)
	for _, tag := range req.SlugNameList {
		newTagMapping[tag] = true
	}

	// cancel follow
	for _, tag := range oldFollowTagList {
		if !newTagMapping[tag.SlugName] {
			err := fs.followRepo.FollowCancel(ctx, tag.ID, req.UserID)
			if err != nil {
				return err
			}
		}
	}

	// new follow
	for _, tagSlugName := range req.SlugNameList {
		if !oldTagMapping[tagSlugName] {
			tagInfo, exist, err := fs.tagRepo.GetTagBySlugName(ctx, tagSlugName)
			if err != nil {
				return err
			}
			if !exist {
				continue
			}
			err = fs.followRepo.Follow(ctx, tagInfo.ID, req.UserID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

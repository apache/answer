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

package revision_common

import (
	"context"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/service/revision"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/jinzhu/copier"
)

// RevisionService user service
type RevisionService struct {
	revisionRepo revision.RevisionRepo
	userRepo     usercommon.UserRepo
}

func NewRevisionService(revisionRepo revision.RevisionRepo,
	userRepo usercommon.UserRepo,
) *RevisionService {
	return &RevisionService{
		revisionRepo: revisionRepo,
		userRepo:     userRepo,
	}
}

func (rs *RevisionService) GetUnreviewedRevisionCount(ctx context.Context, req *schema.RevisionSearch) (count int64, err error) {
	if len(req.GetCanReviewObjectTypes()) == 0 {
		return 0, nil
	}
	return rs.revisionRepo.CountUnreviewedRevision(ctx, req.GetCanReviewObjectTypes())
}

// AddRevision add revision
//
// autoUpdateRevisionID bool : if autoUpdateRevisionID is true , the object.revision_id will be updated,
// if not need auto update object.revision_id, it must be false.
// example: user can edit the object, but need audit, the revision_id will be updated when admin approved
func (rs *RevisionService) AddRevision(ctx context.Context, req *schema.AddRevisionDTO, autoUpdateRevisionID bool) (
	revisionID string, err error) {
	req.ObjectID = uid.DeShortID(req.ObjectID)
	rev := &entity.Revision{}
	_ = copier.Copy(rev, req)
	err = rs.revisionRepo.AddRevision(ctx, rev, autoUpdateRevisionID)
	if err != nil {
		return "", err
	}
	return rev.ID, nil
}

// GetRevision get revision
func (rs *RevisionService) GetRevision(ctx context.Context, revisionID string) (
	revision *entity.Revision, err error) {
	revisionInfo, exist, err := rs.revisionRepo.GetRevisionByID(ctx, revisionID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	return revisionInfo, nil
}

// ExistUnreviewedByObjectID
func (rs *RevisionService) ExistUnreviewedByObjectID(ctx context.Context, objectID string) (revision *entity.Revision, exist bool, err error) {
	objectID = uid.DeShortID(objectID)
	revision, exist, err = rs.revisionRepo.ExistUnreviewedByObjectID(ctx, objectID)
	return revision, exist, err
}

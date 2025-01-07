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

package review

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/review"
	"github.com/segmentfault/pacman/errors"
)

// reviewRepo review repository
type reviewRepo struct {
	data *data.Data
}

// NewReviewRepo new repository
func NewReviewRepo(data *data.Data) review.ReviewRepo {
	return &reviewRepo{
		data: data,
	}
}

// AddReview add review
func (cr *reviewRepo) AddReview(ctx context.Context, review *entity.Review) (err error) {
	_, err = cr.data.DB.Context(ctx).Insert(review)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateReviewStatus update review status
func (cr *reviewRepo) UpdateReviewStatus(ctx context.Context, reviewID int, reviewerUserID string, status int) (err error) {
	_, err = cr.data.DB.Context(ctx).ID(reviewID).Update(&entity.Review{
		ReviewerUserID: reviewerUserID, Status: status})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReview get review one
func (cr *reviewRepo) GetReview(ctx context.Context, reviewID int) (
	review *entity.Review, exist bool, err error) {
	review = &entity.Review{}
	exist, err = cr.data.DB.Context(ctx).ID(reviewID).Get(review)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReviewByObject get review by object
func (cr *reviewRepo) GetReviewByObject(ctx context.Context, objectID string) (review *entity.Review, exist bool, err error) {
	review = &entity.Review{}
	exist, err = cr.data.DB.Context(ctx).Desc("id").Where("object_id = ?", objectID).Get(review)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReviewCount get review count
func (cr *reviewRepo) GetReviewCount(ctx context.Context, status int) (count int64, err error) {
	count, err = cr.data.DB.Context(ctx).Count(&entity.Review{Status: status})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReviewPage get review page
func (cr *reviewRepo) GetReviewPage(ctx context.Context, page, pageSize int, cond *entity.Review) (
	reviewList []*entity.Review, total int64, err error) {
	session := cr.data.DB.Context(ctx).Asc("created_at")
	reviewList = make([]*entity.Review, 0)
	total, err = pager.Help(page, pageSize, &reviewList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

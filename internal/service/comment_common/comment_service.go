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

package comment_common

import (
	"context"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/errors"
)

// CommentCommonRepo comment repository
type CommentCommonRepo interface {
	GetComment(ctx context.Context, commentID string) (comment *entity.Comment, exist bool, err error)
	GetCommentWithoutStatus(ctx context.Context, commentID string) (comment *entity.Comment, exist bool, err error)
	GetCommentCount(ctx context.Context) (count int64, err error)
	RemoveAllUserComment(ctx context.Context, userID string) (err error)
}

// CommentCommonService user service
type CommentCommonService struct {
	commentRepo CommentCommonRepo
}

// NewCommentCommonService new comment service
func NewCommentCommonService(
	commentRepo CommentCommonRepo) *CommentCommonService {
	return &CommentCommonService{
		commentRepo: commentRepo,
	}
}

// GetComment get comment one
func (cs *CommentCommonService) GetComment(ctx context.Context, commentID string) (resp *schema.GetCommentResp, err error) {
	comment, exist, err := cs.commentRepo.GetComment(ctx, commentID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.CommentNotFound)
	}

	resp = &schema.GetCommentResp{}
	resp.SetFromComment(comment)
	return resp, nil
}

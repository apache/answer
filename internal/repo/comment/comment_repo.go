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

package comment

import (
	"context"

	"github.com/segmentfault/pacman/log"

	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/base/pager"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/service/comment"
	"github.com/apache/incubator-answer/internal/service/comment_common"
	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// commentRepo comment repository
type commentRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewCommentRepo new repository
func NewCommentRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) comment.CommentRepo {
	return &commentRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// NewCommentCommonRepo new repository
func NewCommentCommonRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) comment_common.CommentCommonRepo {
	return &commentRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddComment add comment
func (cr *commentRepo) AddComment(ctx context.Context, comment *entity.Comment) (err error) {
	comment.ID, err = cr.uniqueIDRepo.GenUniqueIDStr(ctx, comment.TableName())
	if err != nil {
		return err
	}
	_, err = cr.data.DB.Context(ctx).Insert(comment)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveComment delete comment
func (cr *commentRepo) RemoveComment(ctx context.Context, commentID string) (err error) {
	session := cr.data.DB.Context(ctx).ID(commentID)
	_, err = session.Update(&entity.Comment{Status: entity.CommentStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateCommentContent update comment
func (cr *commentRepo) UpdateCommentContent(
	ctx context.Context, commentID string, originalText string, parsedText string) (err error) {
	_, err = cr.data.DB.Context(ctx).ID(commentID).Update(&entity.Comment{
		OriginalText: originalText,
		ParsedText:   parsedText,
	})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetComment get comment one
func (cr *commentRepo) GetComment(ctx context.Context, commentID string) (
	comment *entity.Comment, exist bool, err error) {
	comment = &entity.Comment{}
	exist, err = cr.data.DB.Context(ctx).Where("status = ?", entity.CommentStatusAvailable).ID(commentID).Get(comment)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCommentWithoutStatus get comment one without status
func (cr *commentRepo) GetCommentWithoutStatus(ctx context.Context, commentID string) (
	comment *entity.Comment, exist bool, err error) {
	comment = &entity.Comment{}
	exist, err = cr.data.DB.Context(ctx).ID(commentID).Get(comment)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (cr *commentRepo) GetCommentCount(ctx context.Context) (count int64, err error) {
	list := make([]*entity.Comment, 0)
	count, err = cr.data.DB.Context(ctx).Where("status = ?", entity.CommentStatusAvailable).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCommentPage get comment page
func (cr *commentRepo) GetCommentPage(ctx context.Context, commentQuery *comment.CommentQuery) (
	commentList []*entity.Comment, total int64, err error,
) {
	commentList = make([]*entity.Comment, 0)

	session := cr.data.DB.Context(ctx)
	session.OrderBy(commentQuery.GetOrderBy())
	session.Where("status = ?", entity.CommentStatusAvailable)

	cond := &entity.Comment{ObjectID: commentQuery.ObjectID, UserID: commentQuery.UserID}
	total, err = pager.Help(commentQuery.Page, commentQuery.PageSize, &commentList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveAllUserComment remove all user comment
func (cr *commentRepo) RemoveAllUserComment(ctx context.Context, userID string) (err error) {
	session := cr.data.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.CommentStatusDeleted)
	affected, err := session.Update(&entity.Comment{Status: entity.CommentStatusDeleted})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	log.Infof("delete user comment, userID: %s, affected: %d", userID, affected)
	return
}

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

package answercommon

import (
	"context"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
)

type AnswerRepo interface {
	AddAnswer(ctx context.Context, answer *entity.Answer) (err error)
	RemoveAnswer(ctx context.Context, id string) (err error)
	RecoverAnswer(ctx context.Context, answerID string) (err error)
	UpdateAnswer(ctx context.Context, answer *entity.Answer, cols []string) (err error)
	GetAnswer(ctx context.Context, id string) (answer *entity.Answer, exist bool, err error)
	GetAnswerList(ctx context.Context, answer *entity.Answer) (answerList []*entity.Answer, err error)
	GetAnswerPage(ctx context.Context, page, pageSize int, answer *entity.Answer) (answerList []*entity.Answer, total int64, err error)
	UpdateAcceptedStatus(ctx context.Context, acceptedAnswerID string, questionID string) error
	GetByID(ctx context.Context, answerID string) (*entity.Answer, bool, error)
	GetByIDs(ctx context.Context, answerIDs ...string) ([]*entity.Answer, error)
	GetCountByQuestionID(ctx context.Context, questionID string) (int64, error)
	GetCountByUserID(ctx context.Context, userID string) (int64, error)
	GetIDsByUserIDAndQuestionID(ctx context.Context, userID string, questionID string) ([]string, error)
	SearchList(ctx context.Context, search *entity.AnswerSearch) ([]*entity.Answer, int64, error)
	GetPersonalAnswerPage(ctx context.Context, cond *entity.PersonalAnswerPageQueryCond) (
		resp []*entity.Answer, total int64, err error)
	AdminSearchList(ctx context.Context, search *schema.AdminAnswerPageReq) ([]*entity.Answer, int64, error)
	UpdateAnswerStatus(ctx context.Context, answerID string, status int) (err error)
	GetAnswerCount(ctx context.Context) (count int64, err error)
	RemoveAllUserAnswer(ctx context.Context, userID string) (err error)
	SumVotesByQuestionID(ctx context.Context, questionID string) (float64, error)
	DeletePermanentlyAnswers(ctx context.Context) (err error)
}

// AnswerCommon user service
type AnswerCommon struct {
	answerRepo AnswerRepo
}

func NewAnswerCommon(answerRepo AnswerRepo) *AnswerCommon {
	return &AnswerCommon{
		answerRepo: answerRepo,
	}
}

func (as *AnswerCommon) SearchAnswerIDs(ctx context.Context, userID, questionID string) ([]string, error) {
	ids, err := as.answerRepo.GetIDsByUserIDAndQuestionID(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (as *AnswerCommon) AdminSearchList(ctx context.Context, req *schema.AdminAnswerPageReq) (
	resp []*entity.Answer, count int64, err error) {
	resp, count, err = as.answerRepo.AdminSearchList(ctx, req)
	if handler.GetEnableShortID(ctx) {
		for _, item := range resp {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return resp, count, err
}

func (as *AnswerCommon) Search(ctx context.Context, search *entity.AnswerSearch) ([]*entity.Answer, int64, error) {
	list, count, err := as.answerRepo.SearchList(ctx, search)
	if err != nil {
		return list, count, err
	}
	return list, count, err
}

func (as *AnswerCommon) PersonalAnswerPage(ctx context.Context,
	cond *entity.PersonalAnswerPageQueryCond) ([]*entity.Answer, int64, error) {
	return as.answerRepo.GetPersonalAnswerPage(ctx, cond)
}

func (as *AnswerCommon) ShowFormat(ctx context.Context, data *entity.Answer) *schema.AnswerInfo {
	info := schema.AnswerInfo{}
	info.ID = data.ID
	info.QuestionID = data.QuestionID
	info.Content = data.OriginalText
	info.HTML = data.ParsedText
	info.Accepted = data.Accepted
	info.VoteCount = data.VoteCount
	info.CreateTime = data.CreatedAt.Unix()
	info.UpdateTime = data.UpdatedAt.Unix()
	if data.UpdatedAt.Unix() < 1 {
		info.UpdateTime = 0
	}
	info.UserID = data.UserID
	info.UpdateUserID = data.LastEditUserID
	info.Status = data.Status
	info.MemberActions = make([]*schema.PermissionMemberAction, 0)
	return &info
}

func (as *AnswerCommon) AdminShowFormat(ctx context.Context, data *entity.Answer) *schema.AdminAnswerInfo {
	info := schema.AdminAnswerInfo{}
	info.ID = data.ID
	info.QuestionID = data.QuestionID
	info.Accepted = data.Accepted
	info.VoteCount = data.VoteCount
	info.CreateTime = data.CreatedAt.Unix()
	info.UpdateTime = data.UpdatedAt.Unix()
	if data.UpdatedAt.Unix() < 1 {
		info.UpdateTime = 0
	}
	info.UserID = data.UserID
	info.UpdateUserID = data.LastEditUserID
	info.Description = htmltext.FetchExcerpt(data.ParsedText, "...", 240)
	return &info
}

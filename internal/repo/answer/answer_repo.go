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

package answer

import (
	"context"
	"time"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/pager"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	answercommon "github.com/apache/incubator-answer/internal/service/answer_common"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/internal/service/unique"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/apache/incubator-answer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// answerRepo answer repository
type answerRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
	userRankRepo rank.UserRankRepo
	activityRepo activity_common.ActivityRepo
}

// NewAnswerRepo new repository
func NewAnswerRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
	userRankRepo rank.UserRankRepo,
	activityRepo activity_common.ActivityRepo,
) answercommon.AnswerRepo {
	return &answerRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
		userRankRepo: userRankRepo,
		activityRepo: activityRepo,
	}
}

// AddAnswer add answer
func (ar *answerRepo) AddAnswer(ctx context.Context, answer *entity.Answer) (err error) {
	answer.QuestionID = uid.DeShortID(answer.QuestionID)
	ID, err := ar.uniqueIDRepo.GenUniqueIDStr(ctx, answer.TableName())
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	answer.ID = ID
	_, err = ar.data.DB.Context(ctx).Insert(answer)

	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		answer.ID = uid.EnShortID(answer.ID)
		answer.QuestionID = uid.EnShortID(answer.QuestionID)
	}
	_ = ar.updateSearch(ctx, answer.ID)
	return nil
}

// RemoveAnswer delete answer
func (ar *answerRepo) RemoveAnswer(ctx context.Context, answerID string) (err error) {
	answerID = uid.DeShortID(answerID)
	_, err = ar.data.DB.Context(ctx).ID(answerID).Cols("status").Update(&entity.Answer{
		Status: entity.AnswerStatusDeleted,
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = ar.updateSearch(ctx, answerID)
	return nil
}

// RecoverAnswer recover answer
func (ar *answerRepo) RecoverAnswer(ctx context.Context, answerID string) (err error) {
	answerID = uid.DeShortID(answerID)
	_, err = ar.data.DB.Context(ctx).ID(answerID).Cols("status").Update(&entity.Answer{
		Status: entity.AnswerStatusAvailable,
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = ar.updateSearch(ctx, answerID)
	return nil
}

// RemoveAllUserAnswer remove all user answer
func (ar *answerRepo) RemoveAllUserAnswer(ctx context.Context, userID string) (err error) {
	// find all answer id that need to be deleted
	answerIDs := make([]string, 0)
	session := ar.data.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.AnswerStatusDeleted)
	err = session.Select("id").Table("answer").Find(&answerIDs)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if len(answerIDs) == 0 {
		return nil
	}

	log.Infof("find %d answers need to be deleted for user %s", len(answerIDs), userID)

	// delete all question
	session = ar.data.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.AnswerStatusDeleted)
	_, err = session.Cols("status", "updated_at").Update(&entity.Answer{
		UpdatedAt: time.Now(),
		Status:    entity.AnswerStatusDeleted,
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// update search content
	for _, id := range answerIDs {
		_ = ar.updateSearch(ctx, id)
	}
	return nil
}

// UpdateAnswer update answer
func (ar *answerRepo) UpdateAnswer(ctx context.Context, answer *entity.Answer, cols []string) (err error) {
	answer.ID = uid.DeShortID(answer.ID)
	answer.QuestionID = uid.DeShortID(answer.QuestionID)
	_, err = ar.data.DB.Context(ctx).ID(answer.ID).Cols(cols...).Update(answer)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = ar.updateSearch(ctx, answer.ID)
	return err
}

func (ar *answerRepo) UpdateAnswerStatus(ctx context.Context, answerID string, status int) (err error) {
	answerID = uid.DeShortID(answerID)
	_, err = ar.data.DB.Context(ctx).ID(answerID).Cols("status").Update(&entity.Answer{Status: status})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = ar.updateSearch(ctx, answerID)
	return
}

// GetAnswer get answer one
func (ar *answerRepo) GetAnswer(ctx context.Context, id string) (
	answer *entity.Answer, exist bool, err error,
) {
	id = uid.DeShortID(id)
	answer = &entity.Answer{}
	exist, err = ar.data.DB.Context(ctx).ID(id).Get(answer)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		answer.ID = uid.EnShortID(answer.ID)
		answer.QuestionID = uid.EnShortID(answer.QuestionID)
	}
	return
}

// GetAnswerCount count answer
func (ar *answerRepo) GetAnswerCount(ctx context.Context) (count int64, err error) {
	var resp = new(entity.Answer)
	count, err = ar.data.DB.Context(ctx).Where("status = ?", entity.AnswerStatusAvailable).Count(resp)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetAnswerList get answer list all
func (ar *answerRepo) GetAnswerList(ctx context.Context, answer *entity.Answer) (answerList []*entity.Answer, err error) {
	answerList = make([]*entity.Answer, 0)
	answer.ID = uid.DeShortID(answer.ID)
	answer.QuestionID = uid.DeShortID(answer.QuestionID)
	err = ar.data.DB.Context(ctx).Find(answerList, answer)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range answerList {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return
}

// GetAnswerPage get answer page
func (ar *answerRepo) GetAnswerPage(ctx context.Context, page, pageSize int, answer *entity.Answer) (answerList []*entity.Answer, total int64, err error) {
	answer.ID = uid.DeShortID(answer.ID)
	answer.QuestionID = uid.DeShortID(answer.QuestionID)
	answerList = make([]*entity.Answer, 0)
	total, err = pager.Help(page, pageSize, answerList, answer, ar.data.DB.Context(ctx))
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range answerList {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return
}

// UpdateAcceptedStatus update all accepted status of this question's answers
func (ar *answerRepo) UpdateAcceptedStatus(ctx context.Context, acceptedAnswerID string, questionID string) error {
	acceptedAnswerID = uid.DeShortID(acceptedAnswerID)
	questionID = uid.DeShortID(questionID)

	// update all this question's answer accepted status to false
	_, err := ar.data.DB.Context(ctx).Where("question_id = ?", questionID).Cols("adopted").Update(&entity.Answer{
		Accepted: schema.AnswerAcceptedFailed,
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// if acceptedAnswerID is not empty, update accepted status to true
	if len(acceptedAnswerID) > 0 && acceptedAnswerID != "0" {
		_, err = ar.data.DB.Context(ctx).Where("id = ?", acceptedAnswerID).Cols("adopted").Update(&entity.Answer{
			Accepted: schema.AnswerAcceptedEnable,
		})
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
	}
	_ = ar.updateSearch(ctx, acceptedAnswerID)
	return nil
}

// GetByID
func (ar *answerRepo) GetByID(ctx context.Context, answerID string) (*entity.Answer, bool, error) {
	var resp entity.Answer
	answerID = uid.DeShortID(answerID)
	has, err := ar.data.DB.Context(ctx).ID(answerID).Get(&resp)
	if err != nil {
		return &resp, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		resp.ID = uid.EnShortID(resp.ID)
		resp.QuestionID = uid.EnShortID(resp.QuestionID)
	}
	return &resp, has, nil
}

func (ar *answerRepo) GetCountByQuestionID(ctx context.Context, questionID string) (int64, error) {
	questionID = uid.DeShortID(questionID)
	var resp = new(entity.Answer)
	count, err := ar.data.DB.Context(ctx).Where("question_id =? and  status = ?", questionID, entity.AnswerStatusAvailable).Count(resp)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (ar *answerRepo) GetCountByUserID(ctx context.Context, userID string) (int64, error) {
	var resp = new(entity.Answer)
	count, err := ar.data.DB.Context(ctx).Where(" user_id = ?  and  status = ?", userID, entity.AnswerStatusAvailable).Count(resp)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (ar *answerRepo) GetIDsByUserIDAndQuestionID(ctx context.Context, userID string, questionID string) ([]string, error) {
	questionID = uid.DeShortID(questionID)
	var ids []string
	resp := make([]string, 0)
	err := ar.data.DB.Context(ctx).Table(entity.Answer{}.TableName()).Where("question_id =? and  user_id = ? and status = ?", questionID, userID, entity.AnswerStatusAvailable).OrderBy("created_at ASC").Cols("id").Find(&ids)
	if err != nil {
		return resp, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, id := range ids {
			resp = append(resp, uid.EnShortID(id))
		}
	} else {
		resp = ids
	}
	return resp, nil
}

// SearchList
func (ar *answerRepo) SearchList(ctx context.Context, search *entity.AnswerSearch) ([]*entity.Answer, int64, error) {
	if search.QuestionID != "" {
		search.QuestionID = uid.DeShortID(search.QuestionID)
	}
	search.ID = uid.DeShortID(search.ID)
	var count int64
	var err error
	rows := make([]*entity.Answer, 0)
	if search.Page > 0 {
		search.Page = search.Page - 1
	} else {
		search.Page = 0
	}
	if search.PageSize == 0 {
		search.PageSize = constant.DefaultPageSize
	}
	offset := search.Page * search.PageSize
	session := ar.data.DB.Context(ctx)

	if search.QuestionID != "" {
		session = session.And("question_id = ?", search.QuestionID)
	}
	if len(search.UserID) > 0 {
		session = session.And("user_id = ?", search.UserID)
	}
	switch search.Order {
	case entity.AnswerSearchOrderByTime:
		session = session.OrderBy("created_at desc")
	case entity.AnswerSearchOrderByTimeAsc:
		session = session.OrderBy("created_at asc")
	case entity.AnswerSearchOrderByVote:
		session = session.OrderBy("vote_count desc")
	default:
		session = session.OrderBy("adopted desc,vote_count desc,created_at asc")
	}
	if !search.IncludeDeleted {
		if search.LoginUserID == "" {
			session = session.And("status = ? ", entity.AnswerStatusAvailable)
		} else {
			session = session.And("status = ? OR user_id = ?", entity.AnswerStatusAvailable, search.LoginUserID)
		}
	}

	session = session.Limit(search.PageSize, offset)
	count, err = session.FindAndCount(&rows)
	if err != nil {
		return rows, count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range rows {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return rows, count, nil
}

// GetPersonalAnswerPage personal answer page
func (ar *answerRepo) GetPersonalAnswerPage(ctx context.Context, req *entity.PersonalAnswerPageQueryCond) (
	resp []*entity.Answer, total int64, err error) {
	cond := &entity.Answer{
		UserID: req.UserID,
	}
	session := ar.data.DB.Context(ctx)
	switch req.Order {
	case entity.AnswerSearchOrderByTime:
		session = session.OrderBy("created_at desc")
	case entity.AnswerSearchOrderByTimeAsc:
		session = session.OrderBy("created_at asc")
	case entity.AnswerSearchOrderByVote:
		session = session.OrderBy("vote_count desc")
	default:
		session = session.OrderBy("adopted desc,vote_count desc,created_at asc")
	}
	if req.ShowPending {
		session = session.And("status != ?", entity.AnswerStatusDeleted)
	} else {
		session = session.And("status = ?", entity.AnswerStatusAvailable)
	}
	resp = make([]*entity.Answer, 0)
	total, err = pager.Help(req.Page, req.PageSize, &resp, cond, session)
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range resp {
			item.ID = uid.EnShortID(item.ID)
			item.QuestionID = uid.EnShortID(item.QuestionID)
		}
	}
	return resp, total, nil
}

func (ar *answerRepo) AdminSearchList(ctx context.Context, req *schema.AdminAnswerPageReq) (
	resp []*entity.Answer, total int64, err error) {
	cond := &entity.Answer{}
	session := ar.data.DB.Context(ctx)
	if len(req.QuestionID) == 0 && len(req.AnswerID) == 0 {
		session.Join("INNER", "question", "answer.question_id = question.id")
		if len(req.QuestionTitle) > 0 {
			session.Where("question.title like ?", "%"+req.QuestionTitle+"%")
		}
	}
	if len(req.AnswerID) > 0 {
		cond.ID = req.AnswerID
	}
	if len(req.QuestionID) > 0 {
		session.Where("answer.question_id = ?", req.QuestionID)
	}
	if req.Status > 0 {
		cond.Status = req.Status
	}
	session.Desc("answer.created_at")

	resp = make([]*entity.Answer, 0)
	total, err = pager.Help(req.Page, req.PageSize, &resp, cond, session)
	if err != nil {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return resp, total, nil
}

// updateSearch update search, if search plugin not enable, do nothing
func (ar *answerRepo) updateSearch(ctx context.Context, answerID string) (err error) {
	answerID = uid.DeShortID(answerID)
	// check search plugin
	var (
		s plugin.Search
	)
	_ = plugin.CallSearch(func(search plugin.Search) error {
		s = search
		return nil
	})
	if s == nil {
		return
	}
	answer, exist, err := ar.GetAnswer(ctx, answerID)
	if !exist {
		return
	}
	if err != nil {
		return err
	}

	// get question
	var (
		question = new(entity.Question)
	)
	exist, err = ar.data.DB.Context(ctx).Where("id = ?", answer.QuestionID).Get(&question)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return
	}

	// get tags
	var (
		tagListList = make([]*entity.TagRel, 0)
		tags        = make([]string, 0)
	)
	st := ar.data.DB.Context(ctx).Where("object_id = ?", uid.DeShortID(question.ID))
	st.Where("status = ?", entity.TagRelStatusAvailable)
	err = st.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	for _, tag := range tagListList {
		tags = append(tags, tag.TagID)
	}

	content := &plugin.SearchContent{
		ObjectID:    answerID,
		Title:       question.Title,
		Type:        constant.AnswerObjectType,
		Content:     answer.OriginalText,
		Answers:     0,
		Status:      plugin.SearchContentStatus(answer.Status),
		Tags:        tags,
		QuestionID:  answer.QuestionID,
		UserID:      answer.UserID,
		Views:       int64(question.ViewCount),
		Created:     answer.CreatedAt.Unix(),
		Active:      answer.UpdatedAt.Unix(),
		Score:       int64(answer.VoteCount),
		HasAccepted: answer.Accepted == schema.AnswerAcceptedEnable,
	}
	err = s.UpdateContent(ctx, content)
	return
}

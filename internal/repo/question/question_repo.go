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

package question

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/unique"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// questionRepo question repository
type questionRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewQuestionRepo new repository
func NewQuestionRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
) questioncommon.QuestionRepo {
	return &questionRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddQuestion add question
func (qr *questionRepo) AddQuestion(ctx context.Context, question *entity.Question) (err error) {
	question.ID, err = qr.uniqueIDRepo.GenUniqueIDStr(ctx, question.TableName())
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_, err = qr.data.DB.Context(ctx).Insert(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	return
}

// RemoveQuestion delete question
func (qr *questionRepo) RemoveQuestion(ctx context.Context, id string) (err error) {
	id = uid.DeShortID(id)
	_, err = qr.data.DB.Context(ctx).Where("id =?", id).Delete(&entity.Question{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateQuestion update question
func (qr *questionRepo) UpdateQuestion(ctx context.Context, question *entity.Question, Cols []string) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols(Cols...).Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return
}

func (qr *questionRepo) UpdatePvCount(ctx context.Context, questionID string) (err error) {
	questionID = uid.DeShortID(questionID)
	question := &entity.Question{}
	_, err = qr.data.DB.Context(ctx).Where("id =?", questionID).Incr("view_count", 1).Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateAnswerCount(ctx context.Context, questionID string, num int) (err error) {
	questionID = uid.DeShortID(questionID)
	question := &entity.Question{}
	question.AnswerCount = num
	_, err = qr.data.DB.Context(ctx).Where("id =?", questionID).Cols("answer_count").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateCollectionCount(ctx context.Context, questionID string) (count int64, err error) {
	questionID = uid.DeShortID(questionID)
	_, err = qr.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		count, err = session.Count(&entity.Collection{ObjectID: questionID})
		if err != nil {
			return nil, err
		}

		question := &entity.Question{CollectionCount: int(count)}
		_, err = session.ID(questionID).MustCols("collection_count").Update(question)
		if err != nil {
			return nil, err
		}
		return
	})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (qr *questionRepo) UpdateQuestionStatus(ctx context.Context, questionID string, status int) (err error) {
	questionID = uid.DeShortID(questionID)
	_, err = qr.data.DB.Context(ctx).ID(questionID).Cols("status").Update(&entity.Question{Status: status})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, questionID)
	return nil
}

func (qr *questionRepo) UpdateQuestionStatusWithOutUpdateTime(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("status").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) DeletePermanentlyQuestions(ctx context.Context) (err error) {
	// get all deleted question ids
	ids := make([]string, 0)
	err = qr.data.DB.Context(ctx).Select("id").Table(new(entity.Question).TableName()).
		Where("status = ?", entity.QuestionStatusDeleted).Find(&ids)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if len(ids) == 0 {
		return nil
	}

	// delete all revisions permanently
	_, err = qr.data.DB.Context(ctx).In("object_id", ids).Delete(&entity.Revision{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	_, err = qr.data.DB.Context(ctx).Where("status = ?", entity.QuestionStatusDeleted).Delete(&entity.Question{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (qr *questionRepo) RecoverQuestion(ctx context.Context, questionID string) (err error) {
	questionID = uid.DeShortID(questionID)
	_, err = qr.data.DB.Context(ctx).ID(questionID).Cols("status").Update(&entity.Question{Status: entity.QuestionStatusAvailable})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, questionID)
	return nil
}

func (qr *questionRepo) UpdateQuestionOperation(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("pin", "show").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (qr *questionRepo) UpdateAccepted(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("accepted_answer_id").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return nil
}

func (qr *questionRepo) UpdateLastAnswer(ctx context.Context, question *entity.Question) (err error) {
	question.ID = uid.DeShortID(question.ID)
	_, err = qr.data.DB.Context(ctx).Where("id =?", question.ID).Cols("last_answer_id").Update(question)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	_ = qr.UpdateSearch(ctx, question.ID)
	return nil
}

// GetQuestion get question one
func (qr *questionRepo) GetQuestion(ctx context.Context, id string) (
	question *entity.Question, exist bool, err error,
) {
	id = uid.DeShortID(id)
	question = &entity.Question{}
	question.ID = id
	exist, err = qr.data.DB.Context(ctx).Where("id = ?", id).Get(question)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		question.ID = uid.EnShortID(question.ID)
	}
	return
}

// GetQuestionsByTitle get question list by title
func (qr *questionRepo) GetQuestionsByTitle(ctx context.Context, title string, pageSize int) (
	questionList []*entity.Question, err error) {
	questionList = make([]*entity.Question, 0)
	session := qr.data.DB.Context(ctx)
	session.Where("status != ?", entity.QuestionStatusDeleted)
	session.Where("title like ?", "%"+title+"%")
	session.Limit(pageSize)
	err = session.Find(&questionList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return
}

func (qr *questionRepo) FindByID(ctx context.Context, id []string) (questionList []*entity.Question, err error) {
	for key, itemID := range id {
		id[key] = uid.DeShortID(itemID)
	}
	questionList = make([]*entity.Question, 0)
	err = qr.data.DB.Context(ctx).Table("question").In("id", id).Find(&questionList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return
}

// GetQuestionList get question list all
func (qr *questionRepo) GetQuestionList(ctx context.Context, question *entity.Question) (questionList []*entity.Question, err error) {
	question.ID = uid.DeShortID(question.ID)
	questionList = make([]*entity.Question, 0)
	err = qr.data.DB.Context(ctx).Find(&questionList, question)
	if err != nil {
		return questionList, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	for _, item := range questionList {
		item.ID = uid.DeShortID(item.ID)
	}
	return
}

func (qr *questionRepo) GetQuestionCount(ctx context.Context) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.Where(builder.Lt{"status": entity.QuestionStatusDeleted})
	count, err = session.Count(&entity.Question{Show: entity.QuestionShow})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (qr *questionRepo) GetUnansweredQuestionCount(ctx context.Context) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.Where(builder.Lt{"status": entity.QuestionStatusDeleted}).
		And(builder.Eq{"answer_count": 0})
	count, err = session.Count(&entity.Question{Show: entity.QuestionShow})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (qr *questionRepo) GetResolvedQuestionCount(ctx context.Context) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.Where(builder.Lt{"status": entity.QuestionStatusDeleted}).
		And(builder.Neq{"answer_count": 0}).
		And(builder.Neq{"accepted_answer_id": 0})
	count, err = session.Count(&entity.Question{Show: entity.QuestionShow})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (qr *questionRepo) GetUserQuestionCount(ctx context.Context, userID string, show int) (count int64, err error) {
	session := qr.data.DB.Context(ctx)
	session.Where(builder.Lt{"status": entity.QuestionStatusDeleted})
	count, err = session.Count(&entity.Question{UserID: userID, Show: show})
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (qr *questionRepo) SitemapQuestions(ctx context.Context, page, pageSize int) (
	questionIDList []*schema.SiteMapQuestionInfo, err error) {
	page = page - 1
	questionIDList = make([]*schema.SiteMapQuestionInfo, 0)

	// try to get sitemap data from cache
	cacheKey := fmt.Sprintf(constant.SiteMapQuestionCacheKeyPrefix, page)
	cacheData, exist, err := qr.data.Cache.GetString(ctx, cacheKey)
	if err == nil && exist {
		_ = json.Unmarshal([]byte(cacheData), &questionIDList)
		return questionIDList, nil
	}

	// get sitemap data from db
	rows := make([]*entity.Question, 0)
	session := qr.data.DB.Context(ctx)
	session.Select("id,title,created_at,post_update_time")
	session.Where("`show` = ?", entity.QuestionShow)
	session.Where("status = ? OR status = ?", entity.QuestionStatusAvailable, entity.QuestionStatusClosed)
	session.Limit(pageSize, page*pageSize)
	session.Asc("created_at")
	err = session.Find(&rows)
	if err != nil {
		return questionIDList, err
	}

	// warp data
	for _, question := range rows {
		item := &schema.SiteMapQuestionInfo{ID: question.ID}
		if handler.GetEnableShortID(ctx) {
			item.ID = uid.EnShortID(question.ID)
		}
		item.Title = htmltext.UrlTitle(question.Title)
		if question.PostUpdateTime.IsZero() {
			item.UpdateTime = question.CreatedAt.Format(time.RFC3339)
		} else {
			item.UpdateTime = question.PostUpdateTime.Format(time.RFC3339)
		}
		questionIDList = append(questionIDList, item)
	}

	// set sitemap data to cache
	cacheDataByte, _ := json.Marshal(questionIDList)
	if err := qr.data.Cache.SetString(ctx, cacheKey, string(cacheDataByte), constant.SiteMapQuestionCacheTime); err != nil {
		log.Error(err)
	}
	return questionIDList, nil
}

// GetQuestionPage query question page
func (qr *questionRepo) GetQuestionPage(ctx context.Context, page, pageSize int,
	tagIDs []string, userID, orderCond string, inDays int, showHidden, showPending bool) (
	questionList []*entity.Question, total int64, err error) {
	questionList = make([]*entity.Question, 0)
	session := qr.data.DB.Context(ctx)
	status := []int{entity.QuestionStatusAvailable}
	if orderCond != "unanswered" {
		status = append(status, entity.QuestionStatusClosed)
	}
	if showPending {
		status = append(status, entity.QuestionStatusPending)
	}
	session.Select("question.*")
	session.In("question.status", status)
	if len(tagIDs) > 0 {
		session.Join("LEFT", "tag_rel", "question.id = tag_rel.object_id")
		session.In("tag_rel.tag_id", tagIDs)
		session.And("tag_rel.status = ?", entity.TagRelStatusAvailable)
	}
	if len(userID) > 0 {
		session.And("question.user_id = ?", userID)
		if !showHidden {
			session.And("question.show = ?", entity.QuestionShow)
		}
	} else {
		session.And("question.show = ?", entity.QuestionShow)
	}
	if inDays > 0 {
		session.And("question.created_at > ?", time.Now().AddDate(0, 0, -inDays))
	}

	switch orderCond {
	case "newest":
		session.OrderBy("question.pin desc,question.created_at DESC")
	case "active":
		if inDays == 0 {
			session.And("question.created_at > ?", time.Now().AddDate(0, 0, -180))
		}
		session.And("question.post_update_time > ?", time.Now().AddDate(0, 0, -90))
		session.OrderBy("question.pin desc,question.post_update_time DESC, question.updated_at DESC")
	case "hot":
		session.OrderBy("question.pin desc,question.hot_score DESC")
	case "score":
		session.OrderBy("question.pin desc,question.vote_count DESC, question.view_count DESC")
	case "unanswered":
		session.Where("question.answer_count = 0")
		session.OrderBy("question.pin desc,question.created_at DESC")
	case "frequent":
		session.OrderBy("question.pin DESC, question.linked_count DESC, question.updated_at DESC")
	}

	session.GroupBy("question.id")
	total, err = pager.Help(page, pageSize, &questionList, &entity.Question{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return questionList, total, err
}

// GetRecommendQuestionPageByTags get recommend question page by tags
func (qr *questionRepo) GetRecommendQuestionPageByTags(ctx context.Context, userID string, tagIDs, followedQuestionIDs []string, page, pageSize int) (
	questionList []*entity.Question, total int64, err error) {
	questionList = make([]*entity.Question, 0)
	orderBySQL := "question.pin DESC, question.created_at DESC"

	// Please Make sure every question has at least one tag
	selectSQL := entity.Question{}.TableName() + ".*"
	if len(followedQuestionIDs) > 0 {
		idStr := "'" + strings.Join(followedQuestionIDs, "','") + "'"
		selectSQL += fmt.Sprintf(", CASE WHEN question.id IN (%s) THEN 0 ELSE 1 END AS order_priority", idStr)
		orderBySQL = "order_priority, " + orderBySQL
	}
	session := qr.data.DB.Context(ctx).Select(selectSQL)

	if len(tagIDs) > 0 {
		session.Where("question.user_id != ?", userID).
			And("question.id NOT IN (SELECT question_id FROM answer WHERE user_id = ?)", userID).
			Join("INNER", "tag_rel", "question.id = tag_rel.object_id").
			And("tag_rel.status = ?", entity.TagRelStatusAvailable).
			Join("INNER", "tag", "tag.id = tag_rel.tag_id").
			In("tag.id", tagIDs)
	} else if len(followedQuestionIDs) == 0 {
		return questionList, 0, nil
	}

	if len(followedQuestionIDs) > 0 {
		if len(tagIDs) > 0 {
			// if tags provided, show followed questions and tag questions
			session.Or(builder.In("question.id", followedQuestionIDs))
		} else {
			// if no tags, only show followed questions
			session.Where(builder.In("question.id", followedQuestionIDs))
		}
	}

	session.
		And("question.show = ? and question.status = ?", entity.QuestionShow, entity.QuestionStatusAvailable).
		Distinct("question.id").
		OrderBy(orderBySQL)

	total, err = pager.Help(page, pageSize, &questionList, &entity.Question{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}

	return questionList, total, err
}

func (qr *questionRepo) AdminQuestionPage(ctx context.Context, search *schema.AdminQuestionPageReq) ([]*entity.Question, int64, error) {
	var (
		count   int64
		err     error
		session = qr.data.DB.Context(ctx).Table("question")
	)

	session.Where(builder.Eq{
		"status": search.Status,
	})

	rows := make([]*entity.Question, 0)
	if search.Page > 0 {
		search.Page = search.Page - 1
	} else {
		search.Page = 0
	}
	if search.PageSize == 0 {
		search.PageSize = constant.DefaultPageSize
	}

	// search by question title like or question id
	if len(search.Query) > 0 {
		// check id search
		var (
			idSearch = false
			id       = ""
		)

		if strings.Contains(search.Query, "question:") {
			idSearch = true
			id = strings.TrimSpace(strings.TrimPrefix(search.Query, "question:"))
			id = uid.DeShortID(id)
			for _, r := range id {
				if !unicode.IsDigit(r) {
					idSearch = false
					break
				}
			}
		}

		if idSearch {
			session.And(builder.Eq{
				"id": id,
			})
		} else {
			session.And(builder.Like{
				"title", search.Query,
			})
		}
	}

	offset := search.Page * search.PageSize

	session.OrderBy("created_at desc").
		Limit(search.PageSize, offset)
	count, err = session.FindAndCount(&rows)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return rows, count, err
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range rows {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return rows, count, nil
}

// UpdateSearch update search, if search plugin not enable, do nothing
func (qr *questionRepo) UpdateSearch(ctx context.Context, questionID string) (err error) {
	// check search plugin
	var s plugin.Search
	_ = plugin.CallSearch(func(search plugin.Search) error {
		s = search
		return nil
	})
	if s == nil {
		return
	}
	questionID = uid.DeShortID(questionID)
	question, exist, err := qr.GetQuestion(ctx, questionID)
	if !exist {
		return
	}
	if err != nil {
		return err
	}

	// get tags
	var (
		tagListList = make([]*entity.TagRel, 0)
		tags        = make([]string, 0)
	)
	session := qr.data.DB.Context(ctx).Where("object_id = ?", questionID)
	session.Where("status = ?", entity.TagRelStatusAvailable)
	err = session.Find(&tagListList)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	for _, tag := range tagListList {
		tags = append(tags, tag.TagID)
	}
	content := &plugin.SearchContent{
		ObjectID:    questionID,
		Title:       question.Title,
		Type:        constant.QuestionObjectType,
		Content:     question.OriginalText,
		Answers:     int64(question.AnswerCount),
		Status:      plugin.SearchContentStatus(question.Status),
		Tags:        tags,
		QuestionID:  questionID,
		UserID:      question.UserID,
		Views:       int64(question.ViewCount),
		Created:     question.CreatedAt.Unix(),
		Active:      question.UpdatedAt.Unix(),
		Score:       int64(question.VoteCount),
		HasAccepted: question.AcceptedAnswerID != "" && question.AcceptedAnswerID != "0",
	}
	err = s.UpdateContent(ctx, content)
	return
}

func (qr *questionRepo) RemoveAllUserQuestion(ctx context.Context, userID string) (err error) {
	// get all question id that need to be deleted
	questionIDs := make([]string, 0)
	session := qr.data.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.QuestionStatusDeleted)
	err = session.Select("id").Table("question").Find(&questionIDs)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if len(questionIDs) == 0 {
		return nil
	}

	log.Infof("find %d questions need to be deleted for user %s", len(questionIDs), userID)

	// delete all question
	session = qr.data.DB.Context(ctx).Where("user_id = ?", userID)
	session.Where("status != ?", entity.QuestionStatusDeleted)
	_, err = session.Cols("status", "updated_at").Update(&entity.Question{
		UpdatedAt: time.Now(),
		Status:    entity.QuestionStatusDeleted,
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// update search content
	for _, id := range questionIDs {
		_ = qr.UpdateSearch(ctx, id)
	}
	return nil
}

// LinkQuestion batch insert question link
func (qr *questionRepo) LinkQuestion(ctx context.Context, link ...*entity.QuestionLink) (err error) {
	// Batch retrieve all links
	var links []*entity.QuestionLink
	for _, l := range link {
		l.FromQuestionID = uid.DeShortID(l.FromQuestionID)
		l.ToQuestionID = uid.DeShortID(l.ToQuestionID)
		l.FromAnswerID = uid.DeShortID(l.FromAnswerID)
		l.ToAnswerID = uid.DeShortID(l.ToAnswerID)
		links = append(links, l)
	}
	// Retrieve existing records from the database
	var existLinks []*entity.QuestionLink
	session := qr.data.DB.Context(ctx)
	for _, link := range links {
		session = session.Or(builder.Eq{
			"from_question_id": link.FromQuestionID,
			"to_question_id":   link.ToQuestionID,
			"from_answer_id":   link.FromAnswerID,
			"to_answer_id":     link.ToAnswerID,
		})
	}
	err = session.Find(&existLinks)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// Optimize separation of records that need to be updated or inserted using a map
	existMap := make(map[string]*entity.QuestionLink)
	for _, el := range existLinks {
		key := fmt.Sprintf("%s:%s:%s:%s", el.FromQuestionID, el.ToQuestionID, el.FromAnswerID, el.ToAnswerID)
		existMap[key] = el
	}

	var updateLinks []*entity.QuestionLink
	var insertLinks []*entity.QuestionLink
	for _, link := range links {
		key := fmt.Sprintf("%s:%s:%s:%s", link.FromQuestionID, link.ToQuestionID, link.FromAnswerID, link.ToAnswerID)
		if el, exist := existMap[key]; exist {
			if el.Status == entity.QuestionLinkStatusDeleted {
				el.Status = entity.QuestionLinkStatusAvailable
				el.UpdatedAt = time.Now()
				updateLinks = append(updateLinks, el)
			}
		} else {
			link.Status = entity.QuestionLinkStatusAvailable
			link.CreatedAt = time.Now()
			link.UpdatedAt = time.Now()
			insertLinks = append(insertLinks, link)
		}
	}

	// Batch update
	if len(updateLinks) > 0 {
		for _, link := range updateLinks {
			_, err = qr.data.DB.Context(ctx).ID(link.ID).Cols("status").Update(&entity.QuestionLink{Status: entity.QuestionLinkStatusAvailable})
			if err != nil {
				return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
			}
		}
	}

	// Batch insert
	if len(insertLinks) > 0 {
		_, err = qr.data.DB.Context(ctx).Insert(insertLinks)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
	}

	return
}

// UpdateQuestionLinkCount update question link count
func (qr *questionRepo) UpdateQuestionLinkCount(ctx context.Context, questionID string) (err error) {
	// count the number of links
	count, err := qr.data.DB.Context(ctx).
		Where("to_question_id = ?", questionID).
		Where("status = ?", entity.QuestionLinkStatusAvailable).
		Count(&entity.QuestionLink{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// update the number of links
	_, err = qr.data.DB.Context(ctx).ID(questionID).
		Cols("linked_count").Update(&entity.Question{LinkedCount: int(count)})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetLinkedQuestionIDs get linked question ids
func (qr *questionRepo) GetLinkedQuestionIDs(ctx context.Context, questionID string, status int) (
	questionIDs []string, err error) {
	questionIDs = make([]string, 0)
	err = qr.data.DB.Context(ctx).
		Select("to_question_id").
		Table(new(entity.QuestionLink).TableName()).
		Where("from_question_id = ?", questionID).
		Where("status = ?", status).
		Find(&questionIDs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return questionIDs, nil
}

// RecoverQuestionLink batch recover question link
func (qr *questionRepo) RecoverQuestionLink(ctx context.Context, links ...*entity.QuestionLink) (err error) {
	return qr.UpdateQuestionLinkStatus(ctx, entity.QuestionLinkStatusAvailable, links...)
}

// RemoveQuestionLink batch remove question link
func (qr *questionRepo) RemoveQuestionLink(ctx context.Context, links ...*entity.QuestionLink) (err error) {
	return qr.UpdateQuestionLinkStatus(ctx, entity.QuestionLinkStatusDeleted, links...)
}

// UpdateQuestionLinkStatus update question link status
func (qr *questionRepo) UpdateQuestionLinkStatus(ctx context.Context, status int, links ...*entity.QuestionLink) (err error) {
	if len(links) == 0 {
		return nil
	}

	session := qr.data.DB.Context(ctx).Cols("status")
	for _, link := range links {
		eq := builder.Eq{}
		if link.FromQuestionID != "" {
			eq["from_question_id"] = uid.DeShortID(link.FromQuestionID)
		}
		if link.FromAnswerID != "" {
			eq["from_answer_id"] = uid.DeShortID(link.FromAnswerID)
		}
		if link.ToQuestionID != "" {
			eq["to_question_id"] = uid.DeShortID(link.ToQuestionID)
		}
		if link.ToAnswerID != "" {
			eq["to_answer_id"] = uid.DeShortID(link.ToAnswerID)
		}
		session = session.Or(eq)
	}
	_, err = session.Update(&entity.QuestionLink{Status: status})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetQuestionLink get linked question to questionID
func (qr *questionRepo) GetQuestionLink(ctx context.Context, page, pageSize int, questionID string, orderCond string, inDays int) (questionList []*entity.Question, total int64, err error) {
	questionList = make([]*entity.Question, 0)
	questionID = uid.DeShortID(questionID)
	questionStatus := []int{entity.QuestionStatusAvailable, entity.QuestionStatusClosed, entity.QuestionStatusPending}
	if questionID == "0" {
		return nil, 0, errors.InternalServer(reason.DatabaseError).WithError(
			fmt.Errorf("questionID is empty"),
		).WithStack()
	}

	session := qr.data.DB.Context(ctx).
		Table("question_link").
		Join("INNER", "question", "question_link.from_question_id = question.id").
		Where("question_link.to_question_id = ? AND question.show = ?", questionID, entity.QuestionShow).
		Distinct("question.id").
		Where("question_link.status = ?", entity.QuestionLinkStatusAvailable).
		Select("question.*").
		In("question.status", questionStatus)

	switch orderCond {
	case "newest":
		session.OrderBy("question.pin desc,question.created_at DESC")
	case "active":
		if inDays == 0 {
			session.And("question.created_at > ?", time.Now().AddDate(0, 0, -180))
		}
		session.And("question.post_update_time > ?", time.Now().AddDate(0, 0, -90))
		session.OrderBy("question.pin desc,question.post_update_time DESC, question.updated_at DESC")
	case "hot":
		session.OrderBy("question.pin desc,question.hot_score DESC")
	case "score":
		session.OrderBy("question.pin desc,question.vote_count DESC, question.view_count DESC")
	case "unanswered":
		session.Where("question.answer_count = 0")
		session.OrderBy("question.pin desc,question.created_at DESC")
	case "frequent":
		session.OrderBy("question.pin DESC, question.linked_count DESC, question.updated_at DESC")
	}

	if page > 0 && pageSize > 0 {
		session.Limit(pageSize, (page-1)*pageSize)
	}

	total, err = pager.Help(page, pageSize, &questionList, &entity.Question{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range questionList {
			item.ID = uid.EnShortID(item.ID)
		}
	}
	return
}

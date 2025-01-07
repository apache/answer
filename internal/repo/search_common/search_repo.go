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

package search_common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tagcommon "github.com/apache/answer/internal/service/tag_common"
	"github.com/apache/answer/plugin"

	"github.com/apache/answer/pkg/htmltext"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/search_common"
	"github.com/apache/answer/internal/service/unique"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/obj"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

var (
	qFields = []string{
		"`question`.`id`",
		"`question`.`id` as `question_id`",
		"`title`",
		"`parsed_text`",
		"`question`.`created_at` as `created_at`",
		"`user_id`",
		"`vote_count`",
		"`answer_count`",
		"CASE WHEN `accepted_answer_id` > 0 THEN 2 ELSE 0 END as `accepted`",
		"`question`.`status` as `status`",
		"`post_update_time`",
	}
	aFields = []string{
		"`answer`.`id` as `id`",
		"`question_id`",
		"`question`.`title` as `title`",
		"`answer`.`parsed_text` as `parsed_text`",
		"`answer`.`created_at` as `created_at`",
		"`answer`.`user_id` as `user_id`",
		"`answer`.`vote_count` as `vote_count`",
		"0 as `answer_count`",
		"`adopted` as `accepted`",
		"`answer`.`status` as `status`",
		"`answer`.`created_at` as `post_update_time`",
	}
)

// searchRepo tag repository
type searchRepo struct {
	data         *data.Data
	userCommon   *usercommon.UserCommon
	uniqueIDRepo unique.UniqueIDRepo
	tagCommon    *tagcommon.TagCommonService
}

// NewSearchRepo new repository
func NewSearchRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
	userCommon *usercommon.UserCommon,
	tagCommon *tagcommon.TagCommonService,
) search_common.SearchRepo {
	return &searchRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
		userCommon:   userCommon,
		tagCommon:    tagCommon,
	}
}

// SearchContents search question and answer data
func (sr *searchRepo) SearchContents(ctx context.Context, words []string, tagIDs [][]string, userID string, votes int, page, pageSize int, order string) (resp []*schema.SearchResult, total int64, err error) {
	words = filterWords(words)

	var (
		b     *builder.Builder
		ub    *builder.Builder
		qfs   = qFields
		afs   = aFields
		argsQ = []interface{}{}
		argsA = []interface{}{}
	)

	if order == "relevance" {
		if len(words) > 0 {
			qfs, argsQ = addRelevanceField([]string{"title", "original_text"}, words, qfs)
			afs, argsA = addRelevanceField([]string{"`answer`.`original_text`"}, words, afs)
		} else {
			order = "newest"
		}
	}

	b = builder.MySQL().Select(qfs...).From("`question`")
	ub = builder.MySQL().Select(afs...).From("`answer`").
		LeftJoin("`question`", "`question`.id = `answer`.question_id")

	b.Where(builder.Lt{"`question`.`status`": entity.QuestionStatusDeleted}).
		And(builder.Eq{"`question`.`show`": entity.QuestionShow})
	ub.Where(builder.Lt{"`question`.`status`": entity.QuestionStatusDeleted}).
		And(builder.Lt{"`answer`.`status`": entity.AnswerStatusDeleted}).
		And(builder.Eq{"`question`.`show`": entity.QuestionShow})

	argsQ = append(argsQ, entity.QuestionStatusDeleted, entity.QuestionShow)
	argsA = append(argsA, entity.QuestionStatusDeleted, entity.AnswerStatusDeleted, entity.QuestionShow)

	likeConQ := builder.NewCond()
	likeConA := builder.NewCond()
	for _, word := range words {
		likeConQ = likeConQ.Or(builder.Like{"title", word}).
			Or(builder.Like{"original_text", word})
		argsQ = append(argsQ, "%"+word+"%")
		argsQ = append(argsQ, "%"+word+"%")

		likeConA = likeConA.Or(builder.Like{"`answer`.original_text", word})
		argsA = append(argsA, "%"+word+"%")
	}

	b.Where(likeConQ)
	ub.Where(likeConA)

	// check tag
	for ti, tagID := range tagIDs {
		ast := "tag_rel" + strconv.Itoa(ti)
		b.Join("INNER", "tag_rel as "+ast, "question.id = "+ast+".object_id").
			And(builder.Eq{
				ast + ".status": entity.TagRelStatusAvailable,
			}).
			And(builder.In(ast+".tag_id", tagID))
		ub.Join("INNER", "tag_rel as "+ast, "question_id = "+ast+".object_id").
			And(builder.Eq{
				ast + ".status": entity.TagRelStatusAvailable,
			}).
			And(builder.In(ast+".tag_id", tagID))
		argsQ = append(argsQ, entity.TagRelStatusAvailable)
		argsA = append(argsA, entity.TagRelStatusAvailable)
		for _, t := range tagID {
			argsQ = append(argsQ, t)
			argsA = append(argsA, t)
		}
	}

	// check user
	if userID != "" {
		b.Where(builder.Eq{"question.user_id": userID})
		ub.Where(builder.Eq{"answer.user_id": userID})
		argsQ = append(argsQ, userID)
		argsA = append(argsA, userID)
	}

	// check vote
	if votes == 0 {
		b.Where(builder.Eq{"question.vote_count": votes})
		ub.Where(builder.Eq{"answer.vote_count": votes})
		argsQ = append(argsQ, votes)
		argsA = append(argsA, votes)
	} else if votes > 0 {
		b.Where(builder.Gte{"question.vote_count": votes})
		ub.Where(builder.Gte{"answer.vote_count": votes})
		argsQ = append(argsQ, votes)
		argsA = append(argsA, votes)
	}

	//b = b.Union("all", ub)
	ubSQL, _, err := ub.ToSQL()
	if err != nil {
		return
	}
	bSQL, _, err := b.ToSQL()
	if err != nil {
		return
	}
	sql := fmt.Sprintf("(%s UNION ALL %s)", bSQL, ubSQL)

	countSQL, _, err := builder.MySQL().Select("count(*) total").From(sql, "c").ToSQL()
	if err != nil {
		return
	}

	startNum := (page - 1) * pageSize
	querySQL, _, err := builder.MySQL().Select("*").From(sql, "t").OrderBy(sr.parseOrder(ctx, order)).Limit(pageSize, startNum).ToSQL()
	if err != nil {
		return
	}

	queryArgs := []interface{}{}
	countArgs := []interface{}{}

	queryArgs = append(queryArgs, querySQL)
	queryArgs = append(queryArgs, argsQ...)
	queryArgs = append(queryArgs, argsA...)

	countArgs = append(countArgs, countSQL)
	countArgs = append(countArgs, argsQ...)
	countArgs = append(countArgs, argsA...)

	res, err := sr.data.DB.Context(ctx).Query(queryArgs...)
	if err != nil {
		return
	}

	tr, err := sr.data.DB.Context(ctx).Query(countArgs...)
	if len(tr) != 0 {
		total = converter.StringToInt64(string(tr[0]["total"]))
	}
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	} else {
		resp, err = sr.parseResult(ctx, res, words)
		return
	}
}

// SearchQuestions search question data
func (sr *searchRepo) SearchQuestions(ctx context.Context, words []string, tagIDs [][]string, notAccepted bool, views, answers int, page, pageSize int, order string) (resp []*schema.SearchResult, total int64, err error) {
	words = filterWords(words)
	var (
		qfs  = qFields
		args = []interface{}{}
	)
	if order == "relevance" {
		if len(words) > 0 {
			qfs, args = addRelevanceField([]string{"title", "original_text"}, words, qfs)
		} else {
			order = "newest"
		}
	}

	b := builder.MySQL().Select(qfs...).From("question")

	b.Where(builder.Lt{"`question`.`status`": entity.QuestionStatusDeleted}).And(builder.Eq{"`question`.`show`": entity.QuestionShow})
	args = append(args, entity.QuestionStatusDeleted, entity.QuestionShow)

	likeConQ := builder.NewCond()
	for _, word := range words {
		likeConQ = likeConQ.Or(builder.Like{"title", word}).
			Or(builder.Like{"original_text", word})
		args = append(args, "%"+word+"%")
		args = append(args, "%"+word+"%")
	}
	b.Where(likeConQ)

	// check tag
	for ti, tagID := range tagIDs {
		ast := "tag_rel" + strconv.Itoa(ti)
		b.Join("INNER", "tag_rel as "+ast, "question.id = "+ast+".object_id").
			And(builder.Eq{
				ast + ".status": entity.TagRelStatusAvailable,
			}).
			And(builder.In(ast+".tag_id", tagID))
		args = append(args, entity.TagRelStatusAvailable)
		for _, t := range tagID {
			args = append(args, t)
		}
	}

	// check need filter has not accepted
	if notAccepted {
		b.And(builder.Eq{"accepted_answer_id": 0})
		args = append(args, 0)
	}

	// check views
	if views > -1 {
		b.And(builder.Gte{"view_count": views})
		args = append(args, views)
	}

	// check answers
	if answers == 0 {
		b.And(builder.Eq{"answer_count": answers})
		args = append(args, answers)
	} else if answers > 0 {
		b.And(builder.Gte{"answer_count": answers})
		args = append(args, answers)
	}

	if answers == 0 {
		b.And(builder.Eq{"answer_count": 0})
		args = append(args, 0)
	} else if answers > 0 {
		b.And(builder.Gte{"answer_count": answers})
		args = append(args, answers)
	}

	queryArgs := []interface{}{}
	countArgs := []interface{}{}

	countSQL, _, err := builder.MySQL().Select("count(*) total").From(b, "c").ToSQL()
	if err != nil {
		return
	}

	startNum := (page - 1) * pageSize
	querySQL, _, err := b.OrderBy(sr.parseOrder(ctx, order)).Limit(pageSize, startNum).ToSQL()
	if err != nil {
		return
	}
	queryArgs = append(queryArgs, querySQL)
	queryArgs = append(queryArgs, args...)

	countArgs = append(countArgs, countSQL)
	countArgs = append(countArgs, args...)

	res, err := sr.data.DB.Context(ctx).Query(queryArgs...)
	if err != nil {
		return
	}

	tr, err := sr.data.DB.Context(ctx).Query(countArgs...)
	if err != nil {
		return
	}

	if len(tr) != 0 {
		total = converter.StringToInt64(string(tr[0]["total"]))
	}
	resp, err = sr.parseResult(ctx, res, words)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SearchAnswers search answer data
func (sr *searchRepo) SearchAnswers(ctx context.Context, words []string, tagIDs [][]string, accepted bool, questionID string, page, pageSize int, order string) (resp []*schema.SearchResult, total int64, err error) {
	words = filterWords(words)

	var (
		afs  = aFields
		args = []interface{}{}
	)
	if order == "relevance" {
		if len(words) > 0 {
			afs, args = addRelevanceField([]string{"`answer`.`original_text`"}, words, afs)
		} else {
			order = "newest"
		}
	}

	b := builder.MySQL().Select(afs...).From("`answer`").
		LeftJoin("`question`", "`question`.id = `answer`.question_id")

	b.Where(builder.Lt{"`question`.`status`": entity.QuestionStatusDeleted}).
		And(builder.Lt{"`answer`.`status`": entity.AnswerStatusDeleted}).And(builder.Eq{"`question`.`show`": entity.QuestionShow})
	args = append(args, entity.QuestionStatusDeleted, entity.AnswerStatusDeleted, entity.QuestionShow)

	likeConA := builder.NewCond()
	for _, word := range words {
		likeConA = likeConA.Or(builder.Like{"`answer`.original_text", word})
		args = append(args, "%"+word+"%")
	}

	b.Where(likeConA)

	// check tag
	for ti, tagID := range tagIDs {
		ast := "tag_rel" + strconv.Itoa(ti)
		b.Join("INNER", "tag_rel as "+ast, "question_id = "+ast+".object_id").
			And(builder.Eq{
				ast + ".status": entity.TagRelStatusAvailable,
			}).
			And(builder.In(ast+".tag_id", tagID))
		args = append(args, entity.TagRelStatusAvailable)
		for _, t := range tagID {
			args = append(args, t)
		}
	}

	// check limit accepted
	if accepted {
		b.Where(builder.Eq{"adopted": schema.AnswerAcceptedEnable})
		args = append(args, schema.AnswerAcceptedEnable)
	}

	// check question id
	if questionID != "" {
		b.Where(builder.Eq{"question_id": questionID})
		args = append(args, questionID)
	}

	queryArgs := []interface{}{}
	countArgs := []interface{}{}

	countSQL, _, err := builder.MySQL().Select("count(*) total").From(b, "c").ToSQL()
	if err != nil {
		return
	}

	startNum := (page - 1) * pageSize
	querySQL, _, err := b.OrderBy(sr.parseOrder(ctx, order)).Limit(pageSize, startNum).ToSQL()
	if err != nil {
		return
	}

	queryArgs = append(queryArgs, querySQL)
	queryArgs = append(queryArgs, args...)

	countArgs = append(countArgs, countSQL)
	countArgs = append(countArgs, args...)

	res, err := sr.data.DB.Context(ctx).Query(queryArgs...)
	if err != nil {
		return
	}

	tr, err := sr.data.DB.Context(ctx).Query(countArgs...)
	if err != nil {
		return
	}

	total = converter.StringToInt64(string(tr[0]["total"]))
	resp, err = sr.parseResult(ctx, res, words)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (sr *searchRepo) parseOrder(ctx context.Context, order string) (res string) {
	switch order {
	case "newest":
		res = "created_at desc"
	case "active":
		res = "post_update_time desc"
	case "score":
		res = "vote_count desc"
	case "relevance":
		res = "relevance desc"
	default:
		res = "created_at desc"
	}
	return
}

// ParseSearchPluginResult parse search plugin result
func (sr *searchRepo) ParseSearchPluginResult(ctx context.Context, sres []plugin.SearchResult, words []string) (resp []*schema.SearchResult, err error) {
	var (
		qres []map[string][]byte
		res  = make([]map[string][]byte, 0)
		b    *builder.Builder
	)
	for _, r := range sres {
		switch r.Type {
		case "question":
			b = builder.MySQL().Select(qFields...).From("question").Where(builder.Eq{"id": r.ID}).
				And(builder.Lt{"`status`": entity.QuestionStatusDeleted})
		case "answer":
			b = builder.MySQL().Select(aFields...).From("answer").LeftJoin("`question`", "`question`.`id` = `answer`.`question_id`").
				Where(builder.Eq{"`answer`.`id`": r.ID}).
				And(builder.Lt{"`question`.`status`": entity.QuestionStatusDeleted}).
				And(builder.Lt{"`answer`.`status`": entity.AnswerStatusDeleted}).And(builder.Eq{"`question`.`show`": entity.QuestionShow})
		}
		qres, err = sr.data.DB.Context(ctx).Query(b)
		if err != nil || len(qres) == 0 {
			continue
		}
		res = append(res, qres[0])
	}
	return sr.parseResult(ctx, res, words)
}

// parseResult parse search result, return the data structure
func (sr *searchRepo) parseResult(ctx context.Context, res []map[string][]byte, words []string) (resp []*schema.SearchResult, err error) {
	questionIDs := make([]string, 0)
	userIDs := make([]string, 0)
	resultList := make([]*schema.SearchResult, 0)
	for _, r := range res {
		questionIDs = append(questionIDs, string(r["question_id"]))
		userIDs = append(userIDs, string(r["user_id"]))
		tp, _ := time.ParseInLocation("2006-01-02 15:04:05", string(r["created_at"]), time.Local)

		var ID = string(r["id"])
		var QuestionID = string(r["question_id"])
		if handler.GetEnableShortID(ctx) {
			ID = uid.EnShortID(ID)
			QuestionID = uid.EnShortID(QuestionID)
		}

		object := &schema.SearchObject{
			ID:              ID,
			QuestionID:      QuestionID,
			Title:           string(r["title"]),
			UrlTitle:        htmltext.UrlTitle(string(r["title"])),
			Excerpt:         htmltext.FetchMatchedExcerpt(string(r["parsed_text"]), words, "...", 100),
			CreatedAtParsed: tp.Unix(),
			UserInfo: &schema.SearchObjectUser{
				ID: string(r["user_id"]),
			},
			Tags:        make([]*schema.TagResp, 0),
			VoteCount:   converter.StringToInt(string(r["vote_count"])),
			Accepted:    string(r["accepted"]) == "2",
			AnswerCount: converter.StringToInt(string(r["answer_count"])),
		}

		objectKey, err := obj.GetObjectTypeStrByObjectID(string(r["id"]))
		if err != nil {
			continue
		}

		switch objectKey {
		case "question":
			for k, v := range entity.AdminQuestionSearchStatus {
				if v == converter.StringToInt(string(r["status"])) {
					object.StatusStr = k
					break
				}
			}
		case "answer":
			for k, v := range entity.AdminAnswerSearchStatus {
				if v == converter.StringToInt(string(r["status"])) {
					object.StatusStr = k
					break
				}
			}
		}

		resultList = append(resultList, &schema.SearchResult{
			ObjectType: objectKey,
			Object:     object,
		})
	}

	tagsMap, err := sr.tagCommon.BatchGetObjectTag(ctx, questionIDs)
	if err != nil {
		return nil, err
	}
	userInfoMap, err := sr.userCommon.BatchUserBasicInfoByID(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range resultList {
		tags, ok := tagsMap[item.Object.QuestionID]
		if ok {
			item.Object.Tags = tags
		}
		if userInfo := userInfoMap[item.Object.UserInfo.ID]; userInfo != nil {
			item.Object.UserInfo.Username = userInfo.Username
			item.Object.UserInfo.DisplayName = userInfo.DisplayName
			item.Object.UserInfo.Rank = userInfo.Rank
			item.Object.UserInfo.Status = userInfo.Status
		}
	}
	return resultList, nil
}

func addRelevanceField(searchFields, words, fields []string) (res []string, args []interface{}) {
	relevanceRes := []string{}
	args = []interface{}{}

	for _, searchField := range searchFields {
		var (
			relevance    = "(LENGTH(" + searchField + ") - LENGTH(%s))"
			replacement  = "REPLACE(%s, ?, '')"
			replaceField = searchField
			replaced     string
			argsField    = []interface{}{}
		)

		res = fields
		for i, word := range words {
			if i == 0 {
				argsField = append(argsField, word)
				replaced = fmt.Sprintf(replacement, replaceField)
			} else {
				argsField = append(argsField, word)
				replaced = fmt.Sprintf(replacement, replaced)
			}
		}
		args = append(args, argsField...)

		relevance = fmt.Sprintf(relevance, replaced)
		relevanceRes = append(relevanceRes, relevance)
	}

	res = append(res, "("+strings.Join(relevanceRes, " + ")+") as relevance")
	return
}

func filterWords(words []string) (res []string) {
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			res = append(res, word)
		}
	}
	return
}

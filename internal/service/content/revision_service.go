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

package content

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_queue"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/notice_queue"
	"github.com/apache/answer/internal/service/object_info"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/report_common"
	"github.com/apache/answer/internal/service/review"
	"github.com/apache/answer/internal/service/revision"
	"github.com/apache/answer/internal/service/tag_common"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/obj"
	"github.com/apache/answer/pkg/uid"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// RevisionService user service
type RevisionService struct {
	revisionRepo             revision.RevisionRepo
	userCommon               *usercommon.UserCommon
	questionCommon           *questioncommon.QuestionCommon
	answerService            *AnswerService
	objectInfoService        *object_info.ObjService
	questionRepo             questioncommon.QuestionRepo
	answerRepo               answercommon.AnswerRepo
	tagRepo                  tag_common.TagRepo
	tagCommon                *tagcommon.TagCommonService
	notificationQueueService notice_queue.NotificationQueueService
	activityQueueService     activity_queue.ActivityQueueService
	reportRepo               report_common.ReportRepo
	reviewService            *review.ReviewService
	reviewActivity           activity.ReviewActivityRepo
}

func NewRevisionService(
	revisionRepo revision.RevisionRepo,
	userCommon *usercommon.UserCommon,
	questionCommon *questioncommon.QuestionCommon,
	answerService *AnswerService,
	objectInfoService *object_info.ObjService,
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	tagRepo tag_common.TagRepo,
	tagCommon *tagcommon.TagCommonService,
	notificationQueueService notice_queue.NotificationQueueService,
	activityQueueService activity_queue.ActivityQueueService,
	reportRepo report_common.ReportRepo,
	reviewService *review.ReviewService,
	reviewActivity activity.ReviewActivityRepo,
) *RevisionService {
	return &RevisionService{
		revisionRepo:             revisionRepo,
		userCommon:               userCommon,
		questionCommon:           questionCommon,
		answerService:            answerService,
		objectInfoService:        objectInfoService,
		questionRepo:             questionRepo,
		answerRepo:               answerRepo,
		tagRepo:                  tagRepo,
		tagCommon:                tagCommon,
		notificationQueueService: notificationQueueService,
		activityQueueService:     activityQueueService,
		reportRepo:               reportRepo,
		reviewService:            reviewService,
		reviewActivity:           reviewActivity,
	}
}

func (rs *RevisionService) RevisionAudit(ctx context.Context, req *schema.RevisionAuditReq) (err error) {
	revisioninfo, exist, err := rs.revisionRepo.GetRevisionByID(ctx, req.ID)
	if err != nil {
		return
	}
	if !exist {
		return
	}
	if revisioninfo.Status != entity.RevisionUnreviewedStatus {
		return
	}
	if req.Operation == schema.RevisionAuditReject {
		err = rs.revisionRepo.UpdateStatus(ctx, req.ID, entity.RevisionReviewRejectStatus, req.UserID)
		return
	}
	if req.Operation == schema.RevisionAuditApprove {
		objectType, objectTypeerr := obj.GetObjectTypeStrByObjectID(revisioninfo.ObjectID)
		if objectTypeerr != nil {
			return objectTypeerr
		}
		revisionitem := &schema.GetRevisionResp{}
		_ = copier.Copy(revisionitem, revisioninfo)
		rs.parseItem(ctx, revisionitem)
		var saveErr error
		switch objectType {
		case constant.QuestionObjectType:
			if !req.CanReviewQuestion {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditQuestion(ctx, revisionitem)
			}
		case constant.AnswerObjectType:
			if !req.CanReviewAnswer {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditAnswer(ctx, revisionitem)
			}
		case constant.TagObjectType:
			if !req.CanReviewTag {
				saveErr = errors.BadRequest(reason.RevisionNoPermission)
			} else {
				saveErr = rs.revisionAuditTag(ctx, revisionitem)
			}
		}
		if saveErr != nil {
			return saveErr
		}
		err = rs.revisionRepo.UpdateStatus(ctx, req.ID, entity.RevisionReviewPassStatus, req.UserID)
		if err != nil {
			return err
		}
		err = rs.reviewActivity.Review(ctx, &schema.PassReviewActivity{
			UserID:           revisioninfo.UserID,
			TriggerUserID:    req.UserID,
			ObjectID:         revisioninfo.ObjectID,
			OriginalObjectID: "0",
			RevisionID:       revisioninfo.ID,
		})
		if err != nil {
			log.Errorf("add review activity failed: %v", err)
		}

		msg := &schema.NotificationMsg{
			TriggerUserID:  req.UserID,
			ReceiverUserID: revisioninfo.UserID,
			Type:           schema.NotificationTypeAchievement,
			ObjectID:       revisioninfo.ObjectID,
			ObjectType:     objectType,
		}
		rs.notificationQueueService.Send(ctx, msg)
		return
	}

	return nil
}

func (rs *RevisionService) revisionAuditQuestion(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	questioninfo, ok := revisionitem.ContentParsed.(*schema.QuestionInfoResp)
	if ok {
		var PostUpdateTime time.Time
		dbquestion, exist, dberr := rs.questionRepo.GetQuestion(ctx, questioninfo.ID)
		if dberr != nil || !exist {
			return
		}

		PostUpdateTime = time.Unix(questioninfo.UpdateTime, 0)
		if dbquestion.PostUpdateTime.Unix() > PostUpdateTime.Unix() {
			PostUpdateTime = dbquestion.PostUpdateTime
		}
		question := &entity.Question{}
		question.ID = questioninfo.ID
		question.Title = questioninfo.Title
		question.OriginalText = questioninfo.Content
		question.ParsedText = questioninfo.HTML
		question.UpdatedAt = time.Unix(questioninfo.UpdateTime, 0)
		question.PostUpdateTime = PostUpdateTime
		question.LastEditUserID = revisionitem.UserID
		saveerr := rs.questionRepo.UpdateQuestion(ctx, question, []string{"title", "original_text", "parsed_text", "updated_at", "post_update_time", "last_edit_user_id"})
		if saveerr != nil {
			return saveerr
		}
		objectTagTags := make([]*schema.TagItem, 0)
		for _, tag := range questioninfo.Tags {
			item := &schema.TagItem{}
			item.SlugName = tag.SlugName
			objectTagTags = append(objectTagTags, item)
		}
		objectTagData := schema.TagChange{}
		objectTagData.ObjectID = question.ID
		objectTagData.Tags = objectTagTags
		saveerr = rs.tagCommon.ObjectChangeTag(ctx, &objectTagData)
		if saveerr != nil {
			return saveerr
		}
		rs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         revisionitem.ObjectID,
			ActivityTypeKey:  constant.ActQuestionEdited,
			RevisionID:       revisionitem.ID,
			OriginalObjectID: revisionitem.ObjectID,
		})
	}
	return nil
}

func (rs *RevisionService) revisionAuditAnswer(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	answerinfo, ok := revisionitem.ContentParsed.(*schema.AnswerInfo)
	if ok {

		var PostUpdateTime time.Time
		dbquestion, exist, dberr := rs.questionRepo.GetQuestion(ctx, answerinfo.QuestionID)
		if dberr != nil || !exist {
			return
		}

		PostUpdateTime = time.Unix(answerinfo.UpdateTime, 0)
		if dbquestion.PostUpdateTime.Unix() > PostUpdateTime.Unix() {
			PostUpdateTime = dbquestion.PostUpdateTime
		}

		insertData := new(entity.Answer)
		insertData.ID = answerinfo.ID
		insertData.OriginalText = answerinfo.Content
		insertData.ParsedText = answerinfo.HTML
		insertData.UpdatedAt = time.Unix(answerinfo.UpdateTime, 0)
		insertData.LastEditUserID = revisionitem.UserID
		saveerr := rs.answerRepo.UpdateAnswer(ctx, insertData, []string{"original_text", "parsed_text", "updated_at", "last_edit_user_id"})
		if saveerr != nil {
			return saveerr
		}
		saveerr = rs.questionCommon.UpdatePostSetTime(ctx, answerinfo.QuestionID, PostUpdateTime)
		if saveerr != nil {
			return saveerr
		}
		questionInfo, exist, err := rs.questionRepo.GetQuestion(ctx, answerinfo.QuestionID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.QuestionNotFound)
		}
		msg := &schema.NotificationMsg{
			TriggerUserID:  revisionitem.UserID,
			ReceiverUserID: questionInfo.UserID,
			Type:           schema.NotificationTypeInbox,
			ObjectID:       answerinfo.ID,
		}
		msg.ObjectType = constant.AnswerObjectType
		msg.NotificationAction = constant.NotificationUpdateAnswer
		rs.notificationQueueService.Send(ctx, msg)

		rs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         insertData.ID,
			OriginalObjectID: insertData.ID,
			ActivityTypeKey:  constant.ActAnswerEdited,
			RevisionID:       revisionitem.ID,
		})
	}
	return nil
}

func (rs *RevisionService) revisionAuditTag(ctx context.Context, revisionitem *schema.GetRevisionResp) (err error) {
	taginfo, ok := revisionitem.ContentParsed.(*schema.GetTagResp)
	if ok {
		tag := &entity.Tag{}
		tag.ID = taginfo.TagID
		tag.OriginalText = taginfo.OriginalText
		tag.ParsedText = taginfo.ParsedText
		saveerr := rs.tagRepo.UpdateTag(ctx, tag)
		if saveerr != nil {
			return saveerr
		}

		tagInfo, exist, err := rs.tagCommon.GetTagByID(ctx, taginfo.TagID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.TagNotFound)
		}
		if tagInfo.MainTagID == 0 && len(tagInfo.SlugName) > 0 {
			log.Debugf("tag %s update slug_name", tagInfo.SlugName)
			tagList, err := rs.tagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(tagInfo.ID)})
			if err != nil {
				return err
			}
			updateTagSlugNames := make([]string, 0)
			for _, tag := range tagList {
				updateTagSlugNames = append(updateTagSlugNames, tag.SlugName)
			}
			err = rs.tagRepo.UpdateTagSynonym(ctx, updateTagSlugNames, converter.StringToInt64(tagInfo.ID), tagInfo.MainTagSlugName)
			if err != nil {
				return err
			}
		}

		rs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           revisionitem.UserID,
			ObjectID:         taginfo.TagID,
			OriginalObjectID: taginfo.TagID,
			ActivityTypeKey:  constant.ActTagEdited,
			RevisionID:       revisionitem.ID,
		})
	}
	return nil
}

// GetUnreviewedRevisionPage get unreviewed list
func (rs *RevisionService) GetUnreviewedRevisionPage(ctx context.Context, req *schema.RevisionSearch) (
	resp *pager.PageModel, err error) {
	revisionResp := make([]*schema.GetUnreviewedRevisionResp, 0)
	if len(req.GetCanReviewObjectTypes()) == 0 {
		return pager.NewPageModel(0, revisionResp), nil
	}
	revisionPage, total, err := rs.revisionRepo.GetUnreviewedRevisionPage(
		ctx, req.Page, 1, req.GetCanReviewObjectTypes())
	if err != nil {
		return nil, err
	}
	for _, rev := range revisionPage {
		item := &schema.GetUnreviewedRevisionResp{}
		_, ok := constant.ObjectTypeNumberMapping[rev.ObjectType]
		if !ok {
			continue
		}
		item.Type = constant.ObjectTypeNumberMapping[rev.ObjectType]
		info, err := rs.objectInfoService.GetUnreviewedRevisionInfo(ctx, rev.ObjectID)
		if err != nil {
			return nil, err
		}
		item.Info = info
		revisionitem := &schema.GetRevisionResp{}
		_ = copier.Copy(revisionitem, rev)
		rs.parseItem(ctx, revisionitem)
		item.UnreviewedInfo = revisionitem

		// get user info
		userInfo, exists, e := rs.userCommon.GetUserBasicInfoByID(ctx, revisionitem.UserID)
		if e != nil {
			return nil, e
		}
		if exists {
			var uinfo schema.UserBasicInfo
			_ = copier.Copy(&uinfo, userInfo)
			item.UnreviewedInfo.UserInfo = uinfo
		}
		item.Info.UrlTitle = htmltext.UrlTitle(item.Info.Title)
		item.UnreviewedInfo.UrlTitle = htmltext.UrlTitle(item.UnreviewedInfo.Title)
		revisionResp = append(revisionResp, item)
	}
	return pager.NewPageModel(total, revisionResp), nil
}

// GetRevisionList get revision list all
func (rs *RevisionService) GetRevisionList(ctx context.Context, req *schema.GetRevisionListReq) (resp []schema.GetRevisionResp, err error) {
	var (
		rev  entity.Revision
		revs []entity.Revision
	)

	resp = []schema.GetRevisionResp{}
	_ = copier.Copy(&rev, req)

	revs, err = rs.revisionRepo.GetRevisionList(ctx, &rev)
	if err != nil {
		return
	}

	for _, r := range revs {
		var (
			uinfo schema.UserBasicInfo
			item  schema.GetRevisionResp
		)

		_ = copier.Copy(&item, r)
		rs.parseItem(ctx, &item)

		// get user info
		userInfo, exists, e := rs.userCommon.GetUserBasicInfoByID(ctx, item.UserID)
		if e != nil {
			return nil, e
		}
		if exists {
			err = copier.Copy(&uinfo, userInfo)
			item.UserInfo = uinfo
		}
		resp = append(resp, item)
	}
	return
}

func (rs *RevisionService) parseItem(ctx context.Context, item *schema.GetRevisionResp) {
	var (
		err          error
		question     entity.QuestionWithTagsRevision
		questionInfo *schema.QuestionInfoResp
		answer       entity.Answer
		answerInfo   *schema.AnswerInfo
		tag          entity.Tag
		tagInfo      *schema.GetTagResp
	)

	shortID := handler.GetEnableShortID(ctx)
	if shortID {
		item.ObjectID = uid.EnShortID(item.ObjectID)
	}
	switch item.ObjectType {
	case constant.ObjectTypeStrMapping["question"]:
		err = json.Unmarshal([]byte(item.Content), &question)
		if err != nil {
			break
		}
		questionInfo = rs.questionCommon.ShowFormatWithTag(ctx, &question)
		if shortID {
			questionInfo.ID = uid.EnShortID(questionInfo.ID)
		}
		item.ContentParsed = questionInfo
	case constant.ObjectTypeStrMapping["answer"]:
		err = json.Unmarshal([]byte(item.Content), &answer)
		if err != nil {
			break
		}
		answerInfo = rs.answerService.ShowFormat(ctx, &answer)
		if shortID {
			answerInfo.ID = uid.EnShortID(answerInfo.ID)
			answerInfo.QuestionID = uid.EnShortID(answerInfo.QuestionID)
		}
		item.ContentParsed = answerInfo
	case constant.ObjectTypeStrMapping["tag"]:
		err = json.Unmarshal([]byte(item.Content), &tag)
		if err != nil {
			break
		}
		tagInfo = &schema.GetTagResp{
			TagID:         tag.ID,
			CreatedAt:     tag.CreatedAt.Unix(),
			UpdatedAt:     tag.UpdatedAt.Unix(),
			SlugName:      tag.SlugName,
			DisplayName:   tag.DisplayName,
			OriginalText:  tag.OriginalText,
			ParsedText:    tag.ParsedText,
			FollowCount:   tag.FollowCount,
			QuestionCount: tag.QuestionCount,
			Recommend:     tag.Recommend,
			Reserved:      tag.Reserved,
		}
		tagInfo.GetExcerpt()
		item.ContentParsed = tagInfo
	}

	if err != nil {
		item.ContentParsed = item.Content
	}
	item.CreatedAtParsed = item.CreatedAt.Unix()
}

// CheckCanUpdateRevision can check revision
func (rs *RevisionService) CheckCanUpdateRevision(ctx context.Context, req *schema.CheckCanQuestionUpdate) (
	resp *schema.ErrTypeData, err error) {
	_, exist, err := rs.revisionRepo.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return nil, nil
	}
	if exist {
		return &schema.ErrTypeToast, errors.BadRequest(reason.RevisionReviewUnderway)
	}
	return nil, nil
}

// GetReviewingType get reviewing type
func (rs *RevisionService) GetReviewingType(ctx context.Context, req *schema.GetReviewingTypeReq) (resp []*schema.GetReviewingTypeResp, err error) {
	resp = make([]*schema.GetReviewingTypeResp, 0)

	// get queue amount
	if req.IsAdmin {
		reviewCount, err := rs.reviewService.GetReviewPendingCount(ctx)
		if err != nil {
			log.Errorf("get report count failed: %v", err)
		} else {
			resp = append(resp, &schema.GetReviewingTypeResp{
				Name:       string(constant.QueuedPost),
				Label:      translator.Tr(handler.GetLangByCtx(ctx), constant.ReviewQueuedPostLabel),
				TodoAmount: reviewCount,
			})
		}
	}

	// get flag amount
	if req.IsAdmin {
		reportCount, err := rs.reportRepo.GetReportCount(ctx)
		if err != nil {
			log.Errorf("get report count failed: %v", err)
		} else {
			resp = append(resp, &schema.GetReviewingTypeResp{
				Name:       string(constant.FlaggedPost),
				Label:      translator.Tr(handler.GetLangByCtx(ctx), constant.ReviewFlaggedPostLabel),
				TodoAmount: reportCount,
			})
		}
	}

	// get suggestion amount
	countUnreviewedRevision, err := rs.revisionRepo.CountUnreviewedRevision(ctx, req.GetCanReviewObjectTypes())
	if err != nil {
		log.Errorf("get unreviewed revision count failed: %v", err)
	} else {
		resp = append(resp, &schema.GetReviewingTypeResp{
			Name:       string(constant.SuggestedPostEdit),
			Label:      translator.Tr(handler.GetLangByCtx(ctx), constant.ReviewSuggestedPostEditLabel),
			TodoAmount: countUnreviewedRevision,
		})
	}
	return resp, nil
}

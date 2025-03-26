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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/apache/answer/internal/service/event_queue"
	"github.com/apache/answer/plugin"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/activity_queue"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	collectioncommon "github.com/apache/answer/internal/service/collection_common"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/export"
	metacommon "github.com/apache/answer/internal/service/meta_common"
	"github.com/apache/answer/internal/service/notice_queue"
	"github.com/apache/answer/internal/service/notification"
	"github.com/apache/answer/internal/service/permission"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/review"
	"github.com/apache/answer/internal/service/revision_common"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/internal/service/tag"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/token"
	"github.com/apache/answer/pkg/uid"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
)

// QuestionRepo question repository

// QuestionService user service
type QuestionService struct {
	activityRepo                     activity_common.ActivityRepo
	questionRepo                     questioncommon.QuestionRepo
	answerRepo                       answercommon.AnswerRepo
	tagCommon                        *tagcommon.TagCommonService
	tagService                       *tag.TagService
	questioncommon                   *questioncommon.QuestionCommon
	userCommon                       *usercommon.UserCommon
	userRepo                         usercommon.UserRepo
	userRoleRelService               *role.UserRoleRelService
	revisionService                  *revision_common.RevisionService
	metaService                      *metacommon.MetaCommonService
	collectionCommon                 *collectioncommon.CollectionCommon
	answerActivityService            *activity.AnswerActivityService
	emailService                     *export.EmailService
	notificationQueueService         notice_queue.NotificationQueueService
	externalNotificationQueueService notice_queue.ExternalNotificationQueueService
	activityQueueService             activity_queue.ActivityQueueService
	siteInfoService                  siteinfo_common.SiteInfoCommonService
	newQuestionNotificationService   *notification.ExternalNotificationService
	reviewService                    *review.ReviewService
	configService                    *config.ConfigService
	eventQueueService                event_queue.EventQueueService
	reviewRepo                       review.ReviewRepo
}

func NewQuestionService(
	activityRepo activity_common.ActivityRepo,
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	tagCommon *tagcommon.TagCommonService,
	tagService *tag.TagService,
	questioncommon *questioncommon.QuestionCommon,
	userCommon *usercommon.UserCommon,
	userRepo usercommon.UserRepo,
	userRoleRelService *role.UserRoleRelService,
	revisionService *revision_common.RevisionService,
	metaService *metacommon.MetaCommonService,
	collectionCommon *collectioncommon.CollectionCommon,
	answerActivityService *activity.AnswerActivityService,
	emailService *export.EmailService,
	notificationQueueService notice_queue.NotificationQueueService,
	externalNotificationQueueService notice_queue.ExternalNotificationQueueService,
	activityQueueService activity_queue.ActivityQueueService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	newQuestionNotificationService *notification.ExternalNotificationService,
	reviewService *review.ReviewService,
	configService *config.ConfigService,
	eventQueueService event_queue.EventQueueService,
	reviewRepo review.ReviewRepo,
) *QuestionService {
	return &QuestionService{
		activityRepo:                     activityRepo,
		questionRepo:                     questionRepo,
		answerRepo:                       answerRepo,
		tagCommon:                        tagCommon,
		tagService:                       tagService,
		questioncommon:                   questioncommon,
		userCommon:                       userCommon,
		userRepo:                         userRepo,
		userRoleRelService:               userRoleRelService,
		revisionService:                  revisionService,
		metaService:                      metaService,
		collectionCommon:                 collectionCommon,
		answerActivityService:            answerActivityService,
		emailService:                     emailService,
		notificationQueueService:         notificationQueueService,
		externalNotificationQueueService: externalNotificationQueueService,
		activityQueueService:             activityQueueService,
		siteInfoService:                  siteInfoService,
		newQuestionNotificationService:   newQuestionNotificationService,
		reviewService:                    reviewService,
		configService:                    configService,
		eventQueueService:                eventQueueService,
		reviewRepo:                       reviewRepo,
	}
}

func (qs *QuestionService) CloseQuestion(ctx context.Context, req *schema.CloseQuestionReq) error {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	cf, err := qs.configService.GetConfigByID(ctx, req.CloseType)
	if err != nil || cf == nil {
		return errors.BadRequest(reason.ReportNotFound)
	}
	if cf.Key == constant.ReasonADuplicate && !checker.IsURL(req.CloseMsg) {
		return errors.BadRequest(reason.InvalidURLError)
	}

	questionInfo.Status = entity.QuestionStatusClosed
	err = qs.questionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, questionInfo.Status)
	if err != nil {
		return err
	}

	closeMeta, _ := json.Marshal(schema.CloseQuestionMeta{
		CloseType: req.CloseType,
		CloseMsg:  req.CloseMsg,
	})
	err = qs.metaService.AddMeta(ctx, req.ID, entity.QuestionCloseReasonKey, string(closeMeta))
	if err != nil {
		return err
	}
	if cf.Key == constant.ReasonADuplicate {
		qs.questioncommon.AddQuestionLinkForCloseReason(ctx, questionInfo, req.CloseMsg)
	}

	qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionClosed,
	})
	return nil
}

// ReopenQuestion reopen question
func (qs *QuestionService) ReopenQuestion(ctx context.Context, req *schema.ReopenQuestionReq) error {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	questionInfo.Status = entity.QuestionStatusAvailable
	err = qs.questionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, questionInfo.Status)
	if err != nil {
		return err
	}
	qs.questioncommon.RemoveQuestionLinkForReopen(ctx, questionInfo)
	qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionReopened,
	})
	return nil
}

func (qs *QuestionService) AddQuestionCheckTags(ctx context.Context, Tags []*entity.Tag) ([]string, error) {
	list := make([]string, 0)
	for _, tag := range Tags {
		if tag.Reserved {
			list = append(list, tag.DisplayName)
		}
	}
	if len(list) > 0 {
		return list, errors.BadRequest(reason.RequestFormatError)
	}
	return []string{}, nil
}
func (qs *QuestionService) CheckAddQuestion(ctx context.Context, req *schema.QuestionAdd) (errorlist any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return errorlist, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, Tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}
	return nil, nil
}

// HasNewTag
func (qs *QuestionService) HasNewTag(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	return qs.tagCommon.HasNewTag(ctx, tags)
}

// AddQuestion add question
func (qs *QuestionService) AddQuestion(ctx context.Context, req *schema.QuestionAdd) (questionInfo any, err error) {
	if len(req.Tags) == 0 {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.TagNotFound),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	tagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}
	if !req.QuestionPermission.CanUseReservedTag {
		taglist, err := qs.AddQuestionCheckTags(ctx, tags)
		errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
			strings.Join(taglist, ","))
		if err != nil {
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RecommendTagEnter)
			return errorlist, err
		}
	}

	question := &entity.Question{}
	now := time.Now()
	question.UserID = req.UserID
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.AcceptedAnswerID = "0"
	question.LastAnswerID = "0"
	question.LastEditUserID = "0"
	//question.PostUpdateTime = nil
	question.Status = entity.QuestionStatusPending
	question.RevisionID = "0"
	question.CreatedAt = now
	question.PostUpdateTime = now
	question.Pin = entity.QuestionUnPin
	question.Show = entity.QuestionShow
	//question.UpdatedAt = nil
	err = qs.questionRepo.AddQuestion(ctx, question)
	if err != nil {
		return
	}
	question.Status = qs.reviewService.AddQuestionReview(ctx, question, req.Tags, req.IP, req.UserAgent)
	if err := qs.questionRepo.UpdateQuestionStatus(ctx, question.ID, question.Status); err != nil {
		return nil, err
	}
	if question.Status == entity.QuestionStatusAvailable {
		question.ParsedText, err = qs.questioncommon.UpdateQuestionLink(ctx, question.ID, "", question.ParsedText, question.OriginalText)
		if err != nil {
			return nil, err
		}
		err = qs.questionRepo.UpdateQuestion(ctx, question, []string{"parsed_text"})
		if err != nil {
			return nil, err
		}
	}
	objectTagData := schema.TagChange{}
	objectTagData.ObjectID = question.ID
	objectTagData.Tags = req.Tags
	objectTagData.UserID = req.UserID
	err = qs.ChangeTag(ctx, &objectTagData)
	if err != nil {
		return
	}
	_ = qs.questionRepo.UpdateSearch(ctx, question.ID)

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := qs.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}

	// user add question count
	userQuestionCount, err := qs.questioncommon.GetUserQuestionCount(ctx, question.UserID)
	if err != nil {
		log.Errorf("get user question count error %v", err)
	} else {
		err = qs.userCommon.UpdateQuestionCount(ctx, question.UserID, userQuestionCount)
		if err != nil {
			log.Errorf("update user question count error %v", err)
		}
	}

	qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           question.UserID,
		ObjectID:         question.ID,
		OriginalObjectID: question.ID,
		ActivityTypeKey:  constant.ActQuestionAsked,
		RevisionID:       revisionID,
	})

	if question.Status == entity.QuestionStatusAvailable {
		qs.externalNotificationQueueService.Send(ctx,
			schema.CreateNewQuestionNotificationMsg(question.ID, question.Title, question.UserID, tags))
	}
	qs.eventQueueService.Send(ctx, schema.NewEvent(constant.EventQuestionCreate, req.UserID).TID(question.ID).
		QID(question.ID, question.UserID))

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// OperationQuestion
func (qs *QuestionService) OperationQuestion(ctx context.Context, req *schema.OperationQuestionReq) (err error) {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}
	// Hidden question cannot be placed at the top
	if questionInfo.Show == entity.QuestionHide && req.Operation == schema.QuestionOperationPin {
		return nil
	}
	// Question cannot be hidden when they are at the top
	if questionInfo.Pin == entity.QuestionPin && req.Operation == schema.QuestionOperationHide {
		return nil
	}

	switch req.Operation {
	case schema.QuestionOperationHide:
		questionInfo.Show = entity.QuestionHide
		err = qs.questionRepo.RemoveQuestionLink(ctx, &entity.QuestionLink{
			FromQuestionID: questionInfo.ID,
		}, &entity.QuestionLink{
			ToQuestionID: questionInfo.ID,
		})
		if err != nil {
			return
		}
		err = qs.tagCommon.HideTagRelListByObjectID(ctx, req.ID)
		if err != nil {
			return err
		}
		err = qs.tagCommon.RefreshTagCountByQuestionID(ctx, req.ID)
		if err != nil {
			return err
		}
	case schema.QuestionOperationShow:
		questionInfo.Show = entity.QuestionShow
		err = qs.questionRepo.RecoverQuestionLink(ctx, &entity.QuestionLink{
			FromQuestionID: questionInfo.ID,
		}, &entity.QuestionLink{
			ToQuestionID: questionInfo.ID,
		})
		if err != nil {
			return
		}
		err = qs.tagCommon.ShowTagRelListByObjectID(ctx, req.ID)
		if err != nil {
			return err
		}
		err = qs.tagCommon.RefreshTagCountByQuestionID(ctx, req.ID)
		if err != nil {
			return err
		}
	case schema.QuestionOperationPin:
		questionInfo.Pin = entity.QuestionPin
	case schema.QuestionOperationUnPin:
		questionInfo.Pin = entity.QuestionUnPin
	}

	err = qs.questionRepo.UpdateQuestionOperation(ctx, questionInfo)
	if err != nil {
		return err
	}

	actMap := make(map[string]constant.ActivityTypeKey)
	actMap[schema.QuestionOperationPin] = constant.ActQuestionPin
	actMap[schema.QuestionOperationUnPin] = constant.ActQuestionUnPin
	actMap[schema.QuestionOperationHide] = constant.ActQuestionHide
	actMap[schema.QuestionOperationShow] = constant.ActQuestionShow
	_, ok := actMap[req.Operation]
	if ok {
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  actMap[req.Operation],
		})
	}

	return nil
}

// RemoveQuestion delete question
func (qs *QuestionService) RemoveQuestion(ctx context.Context, req *schema.RemoveQuestionReq) (err error) {
	questionInfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	//if the status is deleted, return directly
	if questionInfo.Status == entity.QuestionStatusDeleted {
		return nil
	}
	if !has {
		return nil
	}
	if !req.IsAdmin {
		if questionInfo.UserID != req.UserID {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AcceptedAnswerID != "0" {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}
		if questionInfo.AnswerCount > 1 {
			return errors.BadRequest(reason.QuestionCannotDeleted)
		}

		if questionInfo.AnswerCount == 1 {
			answersearch := &entity.AnswerSearch{}
			answersearch.QuestionID = req.ID
			answerList, _, err := qs.questioncommon.AnswerCommon.Search(ctx, answersearch)
			if err != nil {
				return err
			}
			for _, answer := range answerList {
				if answer.VoteCount > 0 {
					return errors.BadRequest(reason.QuestionCannotDeleted)
				}
			}
		}
	}

	questionInfo.Status = entity.QuestionStatusDeleted
	err = qs.questionRepo.UpdateQuestionStatusWithOutUpdateTime(ctx, questionInfo)
	if err != nil {
		return err
	}

	userQuestionCount, err := qs.questioncommon.GetUserQuestionCount(ctx, questionInfo.UserID)
	if err != nil {
		log.Error("user GetUserQuestionCount error", err.Error())
	} else {
		err = qs.userCommon.UpdateQuestionCount(ctx, questionInfo.UserID, userQuestionCount)
		if err != nil {
			log.Error("user IncreaseQuestionCount error", err.Error())
		}
	}

	// If this question has been reviewed, then delete the review.
	reviewInfo, exist, err := qs.reviewRepo.GetReviewByObject(ctx, questionInfo.ID)
	if exist && err == nil {
		err = qs.reviewRepo.UpdateReviewStatus(ctx, reviewInfo.ID, req.UserID, entity.ReviewStatusRejected)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError)
		}
	}

	//tag count
	tagIDs := make([]string, 0)
	Tags, tagerr := qs.tagCommon.GetObjectEntityTag(ctx, req.ID)
	if tagerr != nil {
		log.Error("GetObjectEntityTag error", tagerr)
		return nil
	}
	for _, v := range Tags {
		tagIDs = append(tagIDs, v.ID)
	}
	err = qs.tagCommon.RemoveTagRelListByObjectID(ctx, req.ID)
	if err != nil {
		log.Error("RemoveTagRelListByObjectID error", err.Error())
	}
	err = qs.tagCommon.RefreshTagQuestionCount(ctx, tagIDs)
	if err != nil {
		log.Error("efreshTagQuestionCount error", err.Error())
	}

	// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
	// facing the problem of recovery.
	// err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
	// if err != nil {
	// 	 log.Errorf("user DeleteQuestion rank rollback error %s", err.Error())
	// }
	err = qs.questionRepo.RemoveQuestionLink(ctx, &entity.QuestionLink{
		FromQuestionID: questionInfo.ID,
	}, &entity.QuestionLink{
		ToQuestionID: questionInfo.ID,
	})
	if err != nil {
		return
	}
	qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           questionInfo.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionDeleted,
	})
	qs.eventQueueService.Send(ctx, schema.NewEvent(constant.EventQuestionDelete, req.UserID).TID(questionInfo.ID).
		QID(questionInfo.ID, questionInfo.UserID))
	return nil
}

func (qs *QuestionService) UpdateQuestionCheckTags(ctx context.Context, req *schema.QuestionUpdate) (errorlist []*validator.FormErrorField, err error) {
	dbinfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}

	oldTags, tagerr := qs.tagCommon.GetObjectEntityTag(ctx, req.ID)
	if tagerr != nil {
		log.Error("GetObjectEntityTag error", tagerr)
		return nil, nil
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := qs.tagCommon.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		log.Error("GetTagListByNames error", tagerr)
		return nil, nil
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	return nil, nil
}

func (qs *QuestionService) RecoverQuestion(ctx context.Context, req *schema.QuestionRecoverReq) (err error) {
	questionInfo, exist, err := qs.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	if questionInfo.Status != entity.QuestionStatusDeleted {
		return nil
	}

	err = qs.questionRepo.RecoverQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}

	// update user's question count
	userQuestionCount, err := qs.questioncommon.GetUserQuestionCount(ctx, questionInfo.UserID)
	if err != nil {
		log.Error("user GetUserQuestionCount error", err.Error())
	} else {
		err = qs.userCommon.UpdateQuestionCount(ctx, questionInfo.UserID, userQuestionCount)
		if err != nil {
			log.Error("user IncreaseQuestionCount error", err.Error())
		}
	}

	// update tag's question count
	if err = qs.tagCommon.RecoverTagRelListByObjectID(ctx, questionInfo.ID); err != nil {
		log.Errorf("remove tag rel list by object id error %v", err)
	}

	tagIDs := make([]string, 0)
	tags, err := qs.tagCommon.GetObjectEntityTag(ctx, questionInfo.ID)
	if err != nil {
		return err
	}
	for _, v := range tags {
		tagIDs = append(tagIDs, v.ID)
	}
	if len(tagIDs) > 0 {
		if err = qs.tagCommon.RefreshTagQuestionCount(ctx, tagIDs); err != nil {
			log.Errorf("update tag's question count failed, %v", err)
		}
	}
	err = qs.questionRepo.RecoverQuestionLink(ctx, &entity.QuestionLink{
		FromQuestionID: questionInfo.ID,
	}, &entity.QuestionLink{
		ToQuestionID: questionInfo.ID,
	})
	if err != nil {
		return
	}

	qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         questionInfo.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionUndeleted,
	})
	return nil
}

func (qs *QuestionService) UpdateQuestionInviteUser(ctx context.Context, req *schema.QuestionUpdateInviteUser) (err error) {
	originQuestion, exist, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}

	//verify invite user
	inviteUserInfoList, err := qs.userCommon.BatchGetUserBasicInfoByUserNames(ctx, req.InviteUser)
	if err != nil {
		log.Error("BatchGetUserBasicInfoByUserNames error", err.Error())
	}
	inviteUserIDs := make([]string, 0)
	for _, item := range req.InviteUser {
		_, ok := inviteUserInfoList[item]
		if ok {
			inviteUserIDs = append(inviteUserIDs, inviteUserInfoList[item].ID)
		}
	}
	inviteUserStr := ""
	inviteUserByte, err := json.Marshal(inviteUserIDs)
	if err != nil {
		log.Error("json.Marshal error", err.Error())
		inviteUserStr = "[]"
	} else {
		inviteUserStr = string(inviteUserByte)
	}
	question := &entity.Question{}
	question.ID = uid.DeShortID(req.ID)
	question.InviteUserID = inviteUserStr

	saveerr := qs.questionRepo.UpdateQuestion(ctx, question, []string{"invite_user_id"})
	if saveerr != nil {
		return saveerr
	}
	//send notification
	oldInviteUserIDsStr := originQuestion.InviteUserID
	oldInviteUserIDs := make([]string, 0)
	needSendNotificationUserIDs := make([]string, 0)
	if oldInviteUserIDsStr != "" {
		err = json.Unmarshal([]byte(oldInviteUserIDsStr), &oldInviteUserIDs)
		if err == nil {
			needSendNotificationUserIDs = converter.ArrayNotInArray(oldInviteUserIDs, inviteUserIDs)
		}
	} else {
		needSendNotificationUserIDs = inviteUserIDs
	}
	go qs.notificationInviteUser(ctx, needSendNotificationUserIDs, originQuestion.ID, originQuestion.Title, req.UserID)

	return nil
}

func (qs *QuestionService) notificationInviteUser(
	ctx context.Context, invitedUserIDs []string, questionID, questionTitle, questionUserID string) {
	inviter, exist, err := qs.userCommon.GetUserBasicInfoByID(ctx, questionUserID)
	if err != nil {
		log.Error(err)
		return
	}
	if !exist {
		log.Warnf("user %s not found", questionUserID)
		return
	}

	users, err := qs.userRepo.BatchGetByID(ctx, invitedUserIDs)
	if err != nil {
		log.Error(err)
		return
	}
	invitee := make(map[string]*entity.User, len(users))
	for _, user := range users {
		invitee[user.ID] = user
	}
	for _, userID := range invitedUserIDs {
		msg := &schema.NotificationMsg{
			ReceiverUserID: userID,
			TriggerUserID:  questionUserID,
			Type:           schema.NotificationTypeInbox,
			ObjectID:       questionID,
		}
		msg.ObjectType = constant.QuestionObjectType
		msg.NotificationAction = constant.NotificationInvitedYouToAnswer
		qs.notificationQueueService.Send(ctx, msg)

		receiverUserInfo, ok := invitee[userID]
		if !ok {
			log.Warnf("user %s not found", userID)
			return
		}
		externalNotificationMsg := &schema.ExternalNotificationMsg{
			ReceiverUserID: receiverUserInfo.ID,
			ReceiverEmail:  receiverUserInfo.EMail,
			ReceiverLang:   receiverUserInfo.Language,
		}
		rawData := &schema.NewInviteAnswerTemplateRawData{
			InviterDisplayName: inviter.DisplayName,
			QuestionTitle:      questionTitle,
			QuestionID:         questionID,
			UnsubscribeCode:    token.GenerateToken(),
		}
		externalNotificationMsg.NewInviteAnswerTemplateRawData = rawData
		qs.externalNotificationQueueService.Send(ctx, externalNotificationMsg)
	}
}

// UpdateQuestion update question
func (qs *QuestionService) UpdateQuestion(ctx context.Context, req *schema.QuestionUpdate) (questionInfo any, err error) {
	var canUpdate bool
	questionInfo = &schema.QuestionInfoResp{}

	_, existUnreviewed, err := qs.revisionService.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return
	}
	if existUnreviewed {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return
	}

	dbinfo, has, err := qs.questionRepo.GetQuestion(ctx, req.ID)
	if err != nil {
		return
	}
	if !has {
		return
	}
	if dbinfo.Status == entity.QuestionStatusDeleted {
		err = errors.BadRequest(reason.QuestionCannotUpdate)
		return nil, err
	}

	now := time.Now()
	question := &entity.Question{}
	question.Title = req.Title
	question.OriginalText = req.Content
	question.ParsedText = req.HTML
	question.ID = uid.DeShortID(req.ID)
	question.UpdatedAt = now
	question.PostUpdateTime = now
	question.UserID = dbinfo.UserID
	question.LastEditUserID = req.UserID

	oldTags, tagerr := qs.tagCommon.GetObjectEntityTag(ctx, question.ID)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	tagNameList := make([]string, 0)
	oldtagNameList := make([]string, 0)
	for _, tag := range req.Tags {
		tag.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		tagNameList = append(tagNameList, tag.SlugName)
	}
	for _, tag := range oldTags {
		oldtagNameList = append(oldtagNameList, tag.SlugName)
	}

	isChange := qs.tagCommon.CheckTagsIsChange(ctx, tagNameList, oldtagNameList)

	//If the content is the same, ignore it
	if dbinfo.Title == req.Title && dbinfo.OriginalText == req.Content && !isChange {
		return
	}

	Tags, tagerr := qs.tagCommon.GetTagListByNames(ctx, tagNameList)
	if tagerr != nil {
		return questionInfo, tagerr
	}

	// if user can not use reserved tag, old reserved tag can not be removed and new reserved tag can not be added.
	if !req.CanUseReservedTag {
		CheckOldTag, CheckNewTag, CheckOldTaglist, CheckNewTaglist := qs.CheckChangeReservedTag(ctx, oldTags, Tags)
		if !CheckOldTag {
			errMsg := fmt.Sprintf(`The reserved tag "%s" must be present.`,
				strings.Join(CheckOldTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
		if !CheckNewTag {
			errMsg := fmt.Sprintf(`"%s" can only be used by moderators.`,
				strings.Join(CheckNewTaglist, ","))
			errorlist := make([]*validator.FormErrorField, 0)
			errorlist = append(errorlist, &validator.FormErrorField{
				ErrorField: "tags",
				ErrorMsg:   errMsg,
			})
			err = errors.BadRequest(reason.RequestFormatError).WithMsg(errMsg)
			return errorlist, err
		}
	}
	// Check whether mandatory labels are selected
	recommendExist, err := qs.tagCommon.ExistRecommend(ctx, req.Tags)
	if err != nil {
		return
	}
	if !recommendExist {
		errorlist := make([]*validator.FormErrorField, 0)
		errorlist = append(errorlist, &validator.FormErrorField{
			ErrorField: "tags",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.RecommendTagEnter),
		})
		err = errors.BadRequest(reason.RecommendTagEnter)
		return errorlist, err
	}

	//Administrators and themselves do not need to be audited

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   question.UserID,
		ObjectID: question.ID,
		Title:    question.Title,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview {
		canUpdate = true
	}

	// It's not you or the administrator that needs to be reviewed
	if !canUpdate {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
		revisionDTO.UserID = req.UserID //use revision userid
	} else {
		//Direct modification
		revisionDTO.Status = entity.RevisionReviewPassStatus
		//update question to db
		question.ParsedText, err = qs.questioncommon.UpdateQuestionLink(ctx, question.ID, "", question.ParsedText, question.OriginalText)
		if err != nil {
			return questionInfo, err
		}
		saveerr := qs.questionRepo.UpdateQuestion(ctx, question, []string{"title", "original_text", "parsed_text", "updated_at", "post_update_time", "last_edit_user_id"})
		if saveerr != nil {
			return questionInfo, saveerr
		}
		objectTagData := schema.TagChange{}
		objectTagData.ObjectID = question.ID
		objectTagData.Tags = req.Tags
		objectTagData.UserID = req.UserID
		tagerr := qs.ChangeTag(ctx, &objectTagData)
		if tagerr != nil {
			return questionInfo, tagerr
		}
	}

	questionWithTagsRevision, err := qs.changeQuestionToRevision(ctx, question, Tags)
	if err != nil {
		return nil, err
	}
	infoJSON, _ := json.Marshal(questionWithTagsRevision)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := qs.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return
	}
	if canUpdate {
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         question.ID,
			ActivityTypeKey:  constant.ActQuestionEdited,
			RevisionID:       revisionID,
			OriginalObjectID: question.ID,
		})
		qs.eventQueueService.Send(ctx, schema.NewEvent(constant.EventQuestionUpdate, req.UserID).TID(question.ID).
			QID(question.ID, question.UserID))
	}

	questionInfo, err = qs.GetQuestion(ctx, question.ID, question.UserID, req.QuestionPermission)
	return
}

// GetQuestion get question one
func (qs *QuestionService) GetQuestion(ctx context.Context, questionID, userID string,
	per schema.QuestionPermission) (resp *schema.QuestionInfoResp, err error) {
	question, err := qs.questioncommon.Info(ctx, questionID, userID)
	if err != nil {
		return
	}
	// If the question is deleted or pending, only the administrator and the author can view it
	if (question.Status == entity.QuestionStatusDeleted ||
		question.Status == entity.QuestionStatusPending) && !per.CanReopen && question.UserID != userID {
		return nil, errors.NotFound(reason.QuestionNotFound)
	}
	if question.Status != entity.QuestionStatusClosed {
		per.CanReopen = false
	}
	if question.Status == entity.QuestionStatusClosed {
		per.CanClose = false
	}
	if question.Pin == entity.QuestionPin {
		per.CanPin = false
		per.CanHide = false
	}
	if question.Pin == entity.QuestionUnPin {
		per.CanUnPin = false
	}
	if question.Show == entity.QuestionShow {
		per.CanShow = false
	}
	if question.Show == entity.QuestionHide {
		per.CanHide = false
		per.CanPin = false
	}

	if question.Status == entity.QuestionStatusDeleted {
		operation := &schema.Operation{}
		operation.Msg = translator.Tr(handler.GetLangByCtx(ctx), reason.QuestionAlreadyDeleted)
		operation.Level = schema.OperationLevelDanger
		question.Operation = operation
	}
	if question.Status == entity.QuestionStatusPending {
		operation := &schema.Operation{}
		operation.Msg = translator.Tr(handler.GetLangByCtx(ctx), reason.QuestionUnderReview)
		operation.Level = schema.OperationLevelSecondary
		question.Operation = operation
	}

	question.Description = htmltext.FetchExcerpt(question.HTML, "...", 240)
	question.MemberActions = permission.GetQuestionPermission(ctx, userID, question.UserID, question.Status,
		per.CanEdit, per.CanDelete,
		per.CanClose, per.CanReopen, per.CanPin, per.CanHide, per.CanUnPin, per.CanShow,
		per.CanRecover)
	question.ExtendsActions = permission.GetQuestionExtendsPermission(ctx, per.CanInviteOtherToAnswer)
	return question, nil
}

// GetQuestionAndAddPV get question one
func (qs *QuestionService) GetQuestionAndAddPV(ctx context.Context, questionID, loginUserID string,
	per schema.QuestionPermission) (
	resp *schema.QuestionInfoResp, err error) {
	err = qs.questioncommon.UpdatePv(ctx, questionID)
	if err != nil {
		log.Error(err)
	}
	return qs.GetQuestion(ctx, questionID, loginUserID, per)
}

func (qs *QuestionService) InviteUserInfo(ctx context.Context, questionID string) (inviteList []*schema.UserBasicInfo, err error) {
	return qs.questioncommon.InviteUserInfo(ctx, questionID)
}

func (qs *QuestionService) ChangeTag(ctx context.Context, objectTagData *schema.TagChange) error {
	return qs.tagCommon.ObjectChangeTag(ctx, objectTagData)
}

func (qs *QuestionService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity.Tag) (bool, bool, []string, []string) {
	return qs.tagCommon.CheckChangeReservedTag(ctx, oldobjectTagData, objectTagData)
}

// PersonalQuestionPage get question list by user
func (qs *QuestionService) PersonalQuestionPage(ctx context.Context, req *schema.PersonalQuestionPageReq) (
	pageModel *pager.PageModel, err error) {

	userinfo, exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = req.OrderCond
	search.Page = req.Page
	search.PageSize = req.PageSize
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = req.LoginUserID
	// Only author and administrator can view the pending question
	if req.LoginUserID == userinfo.ID || req.IsAdmin {
		search.ShowPending = true
	}
	questionList, total, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return nil, err
	}
	userQuestionInfoList := make([]*schema.UserQuestionInfo, 0)
	for _, item := range questionList {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		status, ok := entity.AdminQuestionSearchStatusIntToString[item.Status]
		if ok {
			info.Status = status
		}
		userQuestionInfoList = append(userQuestionInfoList, info)
	}
	return pager.NewPageModel(total, userQuestionInfoList), nil
}

func (qs *QuestionService) PersonalAnswerPage(ctx context.Context, req *schema.PersonalAnswerPageReq) (
	pageModel *pager.PageModel, err error) {
	userinfo, exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	cond := &entity.PersonalAnswerPageQueryCond{}
	cond.UserID = userinfo.ID
	cond.Page = req.Page
	cond.PageSize = req.PageSize
	cond.ShowPending = req.IsAdmin || req.LoginUserID == cond.UserID
	if req.OrderCond == "newest" {
		cond.Order = entity.AnswerSearchOrderByTime
	} else {
		cond.Order = entity.AnswerSearchOrderByDefault
	}
	questionIDs := make([]string, 0)
	answerList, total, err := qs.questioncommon.AnswerCommon.PersonalAnswerPage(ctx, cond)
	if err != nil {
		return nil, err
	}

	answerlist := make([]*schema.AnswerInfo, 0)
	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	for _, item := range answerList {
		answerinfo := qs.questioncommon.AnswerCommon.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, uid.DeShortID(item.QuestionID))
	}
	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, req.LoginUserID)
	if err != nil {
		return nil, err
	}

	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		} else {
			continue
		}
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		if item.QuestionInfo.Status == entity.QuestionStatusDeleted {
			info.QuestionInfo.Title = "Deleted question"

		}
		userAnswerlist = append(userAnswerlist, info)
	}

	return pager.NewPageModel(total, userAnswerlist), nil
}

// PersonalCollectionPage get collection list by user
func (qs *QuestionService) PersonalCollectionPage(ctx context.Context, req *schema.PersonalCollectionPageReq) (
	pageModel *pager.PageModel, err error) {
	list := make([]*schema.QuestionInfoResp, 0)
	collectionSearch := &entity.CollectionSearch{}
	collectionSearch.UserID = req.UserID
	collectionSearch.Page = req.Page
	collectionSearch.PageSize = req.PageSize
	collectionList, total, err := qs.collectionCommon.SearchList(ctx, collectionSearch)
	if err != nil {
		return nil, err
	}
	questionIDs := make([]string, 0)
	for _, item := range collectionList {
		questionIDs = append(questionIDs, item.ObjectID)
	}

	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, req.UserID)
	if err != nil {
		return nil, err
	}
	for _, id := range questionIDs {
		if handler.GetEnableShortID(ctx) {
			id = uid.EnShortID(id)
		}
		_, ok := questionMaps[id]
		if ok {
			questionMaps[id].LastAnsweredUserInfo = nil
			questionMaps[id].UpdateUserInfo = nil
			questionMaps[id].Content = ""
			questionMaps[id].HTML = ""
			if questionMaps[id].Status == entity.QuestionStatusDeleted {
				questionMaps[id].Title = "Deleted question"
			}
			list = append(list, questionMaps[id])
		}
	}

	return pager.NewPageModel(total, list), nil
}

func (qs *QuestionService) SearchUserTopList(ctx context.Context, userName string, loginUserID string) ([]*schema.UserQuestionInfo, []*schema.UserAnswerInfo, error) {
	answerlist := make([]*schema.AnswerInfo, 0)

	userAnswerlist := make([]*schema.UserAnswerInfo, 0)
	userQuestionlist := make([]*schema.UserQuestionInfo, 0)

	userinfo, Exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, userName)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	if !Exist {
		return userQuestionlist, userAnswerlist, nil
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "score"
	search.Page = 0
	search.PageSize = 5
	search.UserIDBeSearched = userinfo.ID
	search.LoginUserID = loginUserID
	questionlist, _, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	answersearch := &entity.AnswerSearch{}
	answersearch.UserID = userinfo.ID
	answersearch.PageSize = 5
	answersearch.Order = entity.AnswerSearchOrderByVote
	questionIDs := make([]string, 0)
	answerList, _, err := qs.questioncommon.AnswerCommon.Search(ctx, answersearch)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerList {
		answerinfo := qs.questioncommon.AnswerCommon.ShowFormat(ctx, item)
		answerlist = append(answerlist, answerinfo)
		questionIDs = append(questionIDs, item.QuestionID)
	}
	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, loginUserID)
	if err != nil {
		return userQuestionlist, userAnswerlist, err
	}
	for _, item := range answerlist {
		_, ok := questionMaps[item.QuestionID]
		if ok {
			item.QuestionInfo = questionMaps[item.QuestionID]
		}
	}

	for _, item := range questionlist {
		info := &schema.UserQuestionInfo{}
		_ = copier.Copy(info, item)
		info.UrlTitle = htmltext.UrlTitle(info.Title)
		userQuestionlist = append(userQuestionlist, info)
	}

	for _, item := range answerlist {
		info := &schema.UserAnswerInfo{}
		_ = copier.Copy(info, item)
		info.AnswerID = item.ID
		info.QuestionID = item.QuestionID
		info.QuestionInfo.UrlTitle = htmltext.UrlTitle(info.QuestionInfo.Title)
		userAnswerlist = append(userAnswerlist, info)
	}

	return userQuestionlist, userAnswerlist, nil
}

// GetQuestionsByTitle get questions by title
func (qs *QuestionService) GetQuestionsByTitle(ctx context.Context, title string) (
	resp []*schema.QuestionBaseInfo, err error) {
	resp = make([]*schema.QuestionBaseInfo, 0)
	if len(title) == 0 {
		return resp, nil
	}
	// check search plugin
	var finder plugin.Search
	_ = plugin.CallSearch(func(search plugin.Search) error {
		finder = search
		return nil
	})

	var questions []*entity.Question
	if finder != nil {
		// call search plugin if available
		words := []string{title}
		res, _, err := finder.SearchQuestions(ctx, &plugin.SearchBasicCond{
			Words:    words,
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			return resp, err
		}
		// get question ids from res
		questionIDs := make([]string, 0)
		for _, question := range res {
			questionIDs = append(questionIDs, question.ID)
		}
		questions, err = qs.questionRepo.FindByID(ctx, questionIDs)
	} else {
		questions, err = qs.questionRepo.GetQuestionsByTitle(ctx, title, 10)
	}

	if err != nil {
		return resp, err
	}
	for _, question := range questions {
		item := &schema.QuestionBaseInfo{}
		item.ID = question.ID
		item.Title = question.Title
		item.UrlTitle = htmltext.UrlTitle(question.Title)
		item.ViewCount = question.ViewCount
		item.AnswerCount = question.AnswerCount
		item.CollectionCount = question.CollectionCount
		item.FollowCount = question.FollowCount
		status, ok := entity.AdminQuestionSearchStatusIntToString[question.Status]
		if ok {
			item.Status = status
		}
		if question.AcceptedAnswerID != "0" {
			item.AcceptedAnswer = true
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// SimilarQuestion
func (qs *QuestionService) SimilarQuestion(ctx context.Context, questionID string, loginUserID string) ([]*schema.QuestionPageResp, int64, error) {
	question, err := qs.questioncommon.Info(ctx, questionID, loginUserID)
	if err != nil {
		return nil, 0, nil
	}
	tagNames := make([]string, 0, len(question.Tags))
	for _, tag := range question.Tags {
		tagNames = append(tagNames, tag.SlugName)
	}
	search := &schema.QuestionPageReq{}
	search.OrderCond = "hot"
	search.Page = 0
	search.PageSize = 6
	if len(tagNames) > 0 {
		search.Tag = tagNames[0]
	}
	search.LoginUserID = loginUserID
	similarQuestions, _, err := qs.GetQuestionPage(ctx, search)
	if err != nil {
		return nil, 0, err
	}
	var result []*schema.QuestionPageResp
	for _, v := range similarQuestions {
		if uid.DeShortID(v.ID) != questionID {
			result = append(result, v)
		}
	}
	return result, int64(len(result)), nil
}

// GetQuestionPage query questions page
func (qs *QuestionService) GetQuestionPage(ctx context.Context, req *schema.QuestionPageReq) (
	questions []*schema.QuestionPageResp, total int64, err error) {
	questions = make([]*schema.QuestionPageResp, 0)
	// query by user role
	showHidden := false
	if req.LoginUserID != "" && req.UserIDBeSearched != "" {
		showHidden = req.LoginUserID == req.UserIDBeSearched
		if !showHidden {
			userRole, err := qs.userRoleRelService.GetUserRole(ctx, req.LoginUserID)
			if err != nil {
				return nil, 0, err
			}
			showHidden = userRole == role.RoleAdminID || userRole == role.RoleModeratorID
		}
	}
	// query by tag condition
	var tagIDs = make([]string, 0)
	if len(req.Tag) > 0 {
		tagInfo, exist, err := qs.tagCommon.GetTagBySlugName(ctx, strings.ToLower(req.Tag))
		if err != nil {
			return nil, 0, err
		}
		if exist {
			synTagIds, err := qs.tagCommon.GetTagIDsByMainTagID(ctx, tagInfo.ID)
			if err != nil {
				return nil, 0, err
			}
			tagIDs = append(synTagIds, tagInfo.ID)
		} else {
			return questions, 0, nil
		}
	}

	// query by user condition
	if req.Username != "" {
		userinfo, exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, req.Username)
		if err != nil {
			return nil, 0, err
		}
		if !exist {
			return questions, 0, nil
		}
		req.UserIDBeSearched = userinfo.ID
	}

	if req.OrderCond == schema.QuestionOrderCondHot {
		req.InDays = schema.HotInDays
	}

	questionList, total, err := qs.questionRepo.GetQuestionPage(ctx, req.Page, req.PageSize,
		tagIDs, req.UserIDBeSearched, req.OrderCond, req.InDays, showHidden, req.ShowPending)
	if err != nil {
		return nil, 0, err
	}
	questions, err = qs.questioncommon.FormatQuestionsPage(ctx, questionList, req.LoginUserID, req.OrderCond)
	if err != nil {
		return nil, 0, err
	}
	return questions, total, nil
}

// GetRecommendQuestionPage retrieves recommended question page based on following tags and questions.
func (qs *QuestionService) GetRecommendQuestionPage(ctx context.Context, req *schema.QuestionPageReq) (
	questions []*schema.QuestionPageResp, total int64, err error) {
	followingTagsResp, err := qs.tagService.GetFollowingTags(ctx, req.LoginUserID)
	if err != nil {
		return nil, 0, err
	}
	tagIDs := make([]string, 0, len(followingTagsResp))
	for _, tag := range followingTagsResp {
		tagIDs = append(tagIDs, tag.TagID)
	}

	activityType, err := qs.activityRepo.GetActivityTypeByObjectType(ctx, constant.QuestionObjectType, "follow")
	if err != nil {
		return nil, 0, err
	}
	activities, err := qs.activityRepo.GetUserActivitiesByActivityType(ctx, req.LoginUserID, activityType)
	if err != nil {
		return nil, 0, err
	}

	followedQuestionIDs := make([]string, 0, len(activities))
	for _, activity := range activities {
		if activity.Cancelled == entity.ActivityCancelled {
			continue
		}
		followedQuestionIDs = append(followedQuestionIDs, activity.ObjectID)
	}
	questionList, total, err := qs.questionRepo.GetRecommendQuestionPageByTags(ctx, req.LoginUserID, tagIDs, followedQuestionIDs, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	questions, err = qs.questioncommon.FormatQuestionsPage(ctx, questionList, req.LoginUserID, schema.QuestionOrderCondFrequent)
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

func (qs *QuestionService) AdminSetQuestionStatus(ctx context.Context, req *schema.AdminUpdateQuestionStatusReq) error {
	setStatus, ok := entity.AdminQuestionSearchStatus[req.Status]
	if !ok {
		return errors.BadRequest(reason.RequestFormatError)
	}
	questionInfo, exist, err := qs.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	err = qs.questionRepo.UpdateQuestionStatus(ctx, questionInfo.ID, setStatus)
	if err != nil {
		return err
	}

	msg := &schema.NotificationMsg{}
	if setStatus == entity.QuestionStatusDeleted {
		// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
		// facing the problem of recovery.
		//err = qs.answerActivityService.DeleteQuestion(ctx, questionInfo.ID, questionInfo.CreatedAt, questionInfo.VoteCount)
		//if err != nil {
		//	log.Errorf("admin delete question then rank rollback error %s", err.Error())
		//}
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionDeleted,
		})
		msg.NotificationAction = constant.NotificationYourQuestionWasDeleted
	}
	if setStatus == entity.QuestionStatusAvailable && questionInfo.Status == entity.QuestionStatusClosed {
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionReopened,
		})
	}
	if setStatus == entity.QuestionStatusClosed && questionInfo.Status != entity.QuestionStatusClosed {
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           questionInfo.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionClosed,
		})
		msg.NotificationAction = constant.NotificationYourQuestionIsClosed
	}
	// recover
	if setStatus == entity.QuestionStatusAvailable && questionInfo.Status == entity.QuestionStatusDeleted {
		qs.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			TriggerUserID:    converter.StringToInt64(req.UserID),
			ObjectID:         questionInfo.ID,
			OriginalObjectID: questionInfo.ID,
			ActivityTypeKey:  constant.ActQuestionUndeleted,
		})
	}

	if len(msg.NotificationAction) > 0 {
		msg.ObjectID = questionInfo.ID
		msg.Type = schema.NotificationTypeInbox
		msg.ReceiverUserID = questionInfo.UserID
		msg.TriggerUserID = req.UserID
		msg.ObjectType = constant.QuestionObjectType
		qs.notificationQueueService.Send(ctx, msg)
	}
	return nil
}

func (qs *QuestionService) AdminQuestionPage(
	ctx context.Context, req *schema.AdminQuestionPageReq) (
	resp *pager.PageModel, err error) {

	list := make([]*schema.AdminQuestionInfo, 0)
	questionList, count, err := qs.questionRepo.AdminQuestionPage(ctx, req)
	if err != nil {
		return nil, err
	}

	userIds := make([]string, 0)
	for _, info := range questionList {
		item := &schema.AdminQuestionInfo{}
		_ = copier.Copy(item, info)
		item.CreateTime = info.CreatedAt.Unix()
		item.UpdateTime = info.PostUpdateTime.Unix()
		item.EditTime = info.UpdatedAt.Unix()
		list = append(list, item)
		userIds = append(userIds, info.UserID)
	}
	userInfoMap, err := qs.userCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		if u, ok := userInfoMap[item.UserID]; ok {
			item.UserInfo = u
		}
	}
	return pager.NewPageModel(count, list), nil
}

// AdminAnswerPage search answer list
func (qs *QuestionService) AdminAnswerPage(ctx context.Context, req *schema.AdminAnswerPageReq) (
	resp *pager.PageModel, err error) {
	answerList, count, err := qs.questioncommon.AnswerCommon.AdminSearchList(ctx, req)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]string, 0)
	userIds := make([]string, 0)
	answerResp := make([]*schema.AdminAnswerInfo, 0)
	for _, item := range answerList {
		answerInfo := qs.questioncommon.AnswerCommon.AdminShowFormat(ctx, item)
		answerResp = append(answerResp, answerInfo)
		questionIDs = append(questionIDs, item.QuestionID)
		userIds = append(userIds, item.UserID)
	}
	userInfoMap, err := qs.userCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, err
	}
	questionMaps, err := qs.questioncommon.FindInfoByID(ctx, questionIDs, req.LoginUserID)
	if err != nil {
		return nil, err
	}

	for _, item := range answerResp {
		if q, ok := questionMaps[item.QuestionID]; ok {
			item.QuestionInfo.Title = q.Title
		}
		if u, ok := userInfoMap[item.UserID]; ok {
			item.UserInfo = u
		}
	}
	return pager.NewPageModel(count, answerResp), nil
}

func (qs *QuestionService) changeQuestionToRevision(ctx context.Context, questionInfo *entity.Question, tags []*entity.Tag) (
	questionRevision *entity.QuestionWithTagsRevision, err error) {
	questionRevision = &entity.QuestionWithTagsRevision{}
	questionRevision.Question = *questionInfo

	for _, tag := range tags {
		item := &entity.TagSimpleInfoForRevision{}
		_ = copier.Copy(item, tag)
		questionRevision.Tags = append(questionRevision.Tags, item)
	}
	return questionRevision, nil
}

func (qs *QuestionService) SitemapCron(ctx context.Context) {
	siteSeo, err := qs.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	ctx = context.WithValue(ctx, constant.ShortIDFlag, siteSeo.IsShortLink())
	qs.questioncommon.SitemapCron(ctx)
}

func (qs *QuestionService) GetQuestionLink(ctx context.Context, req *schema.GetQuestionLinkReq) (
	questions []*schema.QuestionPageResp, total int64, err error) {
	if req.OrderCond == schema.QuestionOrderCondHot {
		req.InDays = schema.HotInDays
	}

	questionList, total, err := qs.questionRepo.GetQuestionLink(ctx, req.Page, req.PageSize, req.QuestionID, req.OrderCond, req.InDays)
	if err != nil {
		return nil, 0, err
	}

	questions, err = qs.questioncommon.FormatQuestionsPage(ctx, questionList, req.LoginUserID, req.OrderCond)
	if err != nil {
		return nil, 0, err
	}
	return questions, total, nil
}

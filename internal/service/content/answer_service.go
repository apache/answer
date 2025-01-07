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

	"github.com/apache/answer/internal/service/event_queue"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/activity_queue"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	collectioncommon "github.com/apache/answer/internal/service/collection_common"
	"github.com/apache/answer/internal/service/export"
	"github.com/apache/answer/internal/service/notice_queue"
	"github.com/apache/answer/internal/service/permission"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/review"
	"github.com/apache/answer/internal/service/revision_common"
	"github.com/apache/answer/internal/service/role"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/token"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AnswerService user service
type AnswerService struct {
	answerRepo                       answercommon.AnswerRepo
	questionRepo                     questioncommon.QuestionRepo
	questionCommon                   *questioncommon.QuestionCommon
	answerActivityService            *activity.AnswerActivityService
	userCommon                       *usercommon.UserCommon
	collectionCommon                 *collectioncommon.CollectionCommon
	userRepo                         usercommon.UserRepo
	revisionService                  *revision_common.RevisionService
	AnswerCommon                     *answercommon.AnswerCommon
	voteRepo                         activity_common.VoteRepo
	emailService                     *export.EmailService
	roleService                      *role.UserRoleRelService
	notificationQueueService         notice_queue.NotificationQueueService
	externalNotificationQueueService notice_queue.ExternalNotificationQueueService
	activityQueueService             activity_queue.ActivityQueueService
	reviewService                    *review.ReviewService
	eventQueueService                event_queue.EventQueueService
}

func NewAnswerService(
	answerRepo answercommon.AnswerRepo,
	questionRepo questioncommon.QuestionRepo,
	questionCommon *questioncommon.QuestionCommon,
	userCommon *usercommon.UserCommon,
	collectionCommon *collectioncommon.CollectionCommon,
	userRepo usercommon.UserRepo,
	revisionService *revision_common.RevisionService,
	answerAcceptActivityRepo *activity.AnswerActivityService,
	answerCommon *answercommon.AnswerCommon,
	voteRepo activity_common.VoteRepo,
	emailService *export.EmailService,
	roleService *role.UserRoleRelService,
	notificationQueueService notice_queue.NotificationQueueService,
	externalNotificationQueueService notice_queue.ExternalNotificationQueueService,
	activityQueueService activity_queue.ActivityQueueService,
	reviewService *review.ReviewService,
	eventQueueService event_queue.EventQueueService,
) *AnswerService {
	return &AnswerService{
		answerRepo:                       answerRepo,
		questionRepo:                     questionRepo,
		userCommon:                       userCommon,
		collectionCommon:                 collectionCommon,
		questionCommon:                   questionCommon,
		userRepo:                         userRepo,
		revisionService:                  revisionService,
		answerActivityService:            answerAcceptActivityRepo,
		AnswerCommon:                     answerCommon,
		voteRepo:                         voteRepo,
		emailService:                     emailService,
		roleService:                      roleService,
		notificationQueueService:         notificationQueueService,
		externalNotificationQueueService: externalNotificationQueueService,
		activityQueueService:             activityQueueService,
		reviewService:                    reviewService,
		eventQueueService:                eventQueueService,
	}
}

// RemoveAnswer delete answer
func (as *AnswerService) RemoveAnswer(ctx context.Context, req *schema.RemoveAnswerReq) (err error) {
	answerInfo, exist, err := as.answerRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	// if the status is deleted, return directly
	if answerInfo.Status == entity.AnswerStatusDeleted {
		return nil
	}
	roleID, err := as.roleService.GetUserRole(ctx, req.UserID)
	if err != nil {
		return err
	}
	if roleID != role.RoleAdminID && roleID != role.RoleModeratorID {
		if answerInfo.UserID != req.UserID {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if answerInfo.VoteCount > 0 {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if answerInfo.Accepted == schema.AnswerAcceptedEnable {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		_, exist, err := as.questionRepo.GetQuestion(ctx, answerInfo.QuestionID)
		if err != nil {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}
		if !exist {
			return errors.BadRequest(reason.AnswerCannotDeleted)
		}

	}

	err = as.answerRepo.RemoveAnswer(ctx, req.ID)
	if err != nil {
		return err
	}

	// user add question count
	err = as.questionCommon.UpdateAnswerCount(ctx, answerInfo.QuestionID)
	if err != nil {
		log.Error("IncreaseAnswerCount error", err.Error())
	}
	userAnswerCount, err := as.answerRepo.GetCountByUserID(ctx, answerInfo.UserID)
	if err != nil {
		log.Error("GetCountByUserID error", err.Error())
	}
	err = as.userCommon.UpdateAnswerCount(ctx, answerInfo.UserID, int(userAnswerCount))
	if err != nil {
		log.Error("user IncreaseAnswerCount error", err.Error())
	}
	err = as.questionRepo.RemoveQuestionLink(ctx, &entity.QuestionLink{
		FromQuestionID: answerInfo.QuestionID,
		FromAnswerID:   answerInfo.ID,
	}, &entity.QuestionLink{
		ToQuestionID: answerInfo.QuestionID,
		ToAnswerID:   answerInfo.ID,
	})
	if err != nil {
		log.Error("RemoveQuestionLink error", err.Error())
	}

	// #2372 In order to simplify the process and complexity, as well as to consider if it is in-house,
	// facing the problem of recovery.
	//err = as.answerActivityService.DeleteAnswer(ctx, answerInfo.ID, answerInfo.CreatedAt, answerInfo.VoteCount)
	//if err != nil {
	//	log.Errorf("delete answer activity change failed: %s", err.Error())
	//}
	as.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         answerInfo.ID,
		OriginalObjectID: answerInfo.ID,
		ActivityTypeKey:  constant.ActAnswerDeleted,
	})
	as.eventQueueService.Send(ctx, schema.NewEvent(constant.EventAnswerDelete, req.UserID).TID(answerInfo.ID).
		AID(answerInfo.ID, answerInfo.UserID))
	return
}

// RecoverAnswer recover deleted answer
func (as *AnswerService) RecoverAnswer(ctx context.Context, req *schema.RecoverAnswerReq) (err error) {
	answerInfo, exist, err := as.answerRepo.GetByID(ctx, req.AnswerID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.AnswerNotFound)
	}
	if answerInfo.Status != entity.AnswerStatusDeleted {
		return nil
	}
	if err = as.answerRepo.RecoverAnswer(ctx, req.AnswerID); err != nil {
		return err
	}
	if err = as.questionRepo.RecoverQuestionLink(ctx, &entity.QuestionLink{
		FromQuestionID: answerInfo.QuestionID,
		FromAnswerID:   answerInfo.ID,
	}, &entity.QuestionLink{
		ToQuestionID: answerInfo.QuestionID,
		ToAnswerID:   answerInfo.ID,
	}); err != nil {
		return err
	}

	if err = as.questionCommon.UpdateAnswerCount(ctx, answerInfo.QuestionID); err != nil {
		log.Errorf("update answer count failed: %s", err.Error())
	}
	userAnswerCount, err := as.answerRepo.GetCountByUserID(ctx, answerInfo.UserID)
	if err != nil {
		log.Errorf("get user answer count failed: %s", err.Error())
	} else {
		err = as.userCommon.UpdateAnswerCount(ctx, answerInfo.UserID, int(userAnswerCount))
		if err != nil {
			log.Errorf("update user answer count failed: %s", err.Error())
		}
	}
	as.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         answerInfo.ID,
		OriginalObjectID: answerInfo.ID,
		ActivityTypeKey:  constant.ActAnswerUndeleted,
	})
	return nil
}

func (as *AnswerService) Insert(ctx context.Context, req *schema.AnswerAddReq) (string, error) {
	questionInfo, exist, err := as.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.QuestionNotFound)
	}
	if questionInfo.Status == entity.QuestionStatusClosed || questionInfo.Status == entity.QuestionStatusDeleted {
		err = errors.BadRequest(reason.AnswerCannotAddByClosedQuestion)
		return "", err
	}
	insertData := &entity.Answer{}
	insertData.UserID = req.UserID
	insertData.OriginalText = req.Content
	insertData.ParsedText = req.HTML
	insertData.Accepted = schema.AnswerAcceptedFailed
	insertData.QuestionID = req.QuestionID
	insertData.RevisionID = "0"
	insertData.LastEditUserID = "0"
	insertData.Status = entity.AnswerStatusPending
	//insertData.UpdatedAt = now
	if err = as.answerRepo.AddAnswer(ctx, insertData); err != nil {
		return "", err
	}
	insertData.Status = as.reviewService.AddAnswerReview(ctx, insertData, req.IP, req.UserAgent)
	if err := as.answerRepo.UpdateAnswerStatus(ctx, insertData.ID, insertData.Status); err != nil {
		return "", err
	}
	if insertData.Status == entity.AnswerStatusAvailable {
		insertData.ParsedText, err = as.questionCommon.UpdateQuestionLink(ctx, insertData.QuestionID, insertData.ID, insertData.ParsedText, insertData.OriginalText)
		if err != nil {
			return "", err
		}
		if err = as.answerRepo.UpdateAnswer(ctx, insertData, []string{"parsed_text"}); err != nil {
			return "", err
		}
	}
	err = as.questionCommon.UpdateAnswerCount(ctx, req.QuestionID)
	if err != nil {
		log.Error("IncreaseAnswerCount error", err.Error())
	}
	err = as.questionCommon.UpdateLastAnswer(ctx, req.QuestionID, uid.DeShortID(insertData.ID))
	if err != nil {
		log.Error("UpdateLastAnswer error", err.Error())
	}
	err = as.questionCommon.UpdatePostTime(ctx, req.QuestionID)
	if err != nil {
		return insertData.ID, err
	}
	userAnswerCount, err := as.answerRepo.GetCountByUserID(ctx, req.UserID)
	if err != nil {
		log.Error("GetCountByUserID error", err.Error())
	}
	err = as.userCommon.UpdateAnswerCount(ctx, req.UserID, int(userAnswerCount))
	if err != nil {
		log.Error("user IncreaseAnswerCount error", err.Error())
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   insertData.UserID,
		ObjectID: insertData.ID,
		Title:    "",
	}
	infoJSON, _ := json.Marshal(insertData)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := as.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return insertData.ID, err
	}
	if insertData.Status == entity.AnswerStatusAvailable {
		as.notificationAnswerTheQuestion(ctx, questionInfo.UserID, questionInfo.ID, insertData.ID, req.UserID, questionInfo.Title,
			htmltext.FetchExcerpt(insertData.ParsedText, "...", 240))
	}

	as.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           insertData.UserID,
		ObjectID:         insertData.ID,
		OriginalObjectID: insertData.ID,
		ActivityTypeKey:  constant.ActAnswerAnswered,
		RevisionID:       revisionID,
	})
	as.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           insertData.UserID,
		ObjectID:         insertData.ID,
		OriginalObjectID: questionInfo.ID,
		ActivityTypeKey:  constant.ActQuestionAnswered,
	})
	as.eventQueueService.Send(ctx, schema.NewEvent(constant.EventAnswerCreate, req.UserID).TID(insertData.ID).
		AID(insertData.ID, insertData.UserID))
	return insertData.ID, nil
}

func (as *AnswerService) Update(ctx context.Context, req *schema.AnswerUpdateReq) (string, error) {
	var canUpdate bool
	_, existUnreviewed, err := as.revisionService.ExistUnreviewedByObjectID(ctx, req.ID)
	if err != nil {
		return "", err
	}
	if existUnreviewed {
		return "", errors.BadRequest(reason.AnswerCannotUpdate)
	}

	questionInfo, exist, err := as.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.QuestionNotFound)
	}

	answerInfo, exist, err := as.answerRepo.GetByID(ctx, req.ID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.BadRequest(reason.AnswerNotFound)
	}

	if answerInfo.Status == entity.AnswerStatusDeleted {
		return "", errors.BadRequest(reason.AnswerCannotUpdate)
	}

	//If the content is the same, ignore it
	if answerInfo.OriginalText == req.Content {
		return "", nil
	}

	insertData := &entity.Answer{}
	insertData.ID = req.ID
	insertData.UserID = answerInfo.UserID
	insertData.QuestionID = req.QuestionID
	insertData.OriginalText = req.Content
	insertData.ParsedText = req.HTML
	insertData.UpdatedAt = time.Now()
	insertData.LastEditUserID = "0"
	if answerInfo.UserID != req.UserID {
		insertData.LastEditUserID = req.UserID
	}

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: req.ID,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview || answerInfo.UserID == req.UserID {
		canUpdate = true
	}

	if !canUpdate {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
	} else {
		insertData.ParsedText, err = as.questionCommon.UpdateQuestionLink(ctx, insertData.QuestionID, insertData.ID, insertData.ParsedText, insertData.OriginalText)
		if err != nil {
			return "", err
		}
		if err = as.answerRepo.UpdateAnswer(ctx, insertData, []string{"original_text", "parsed_text", "updated_at", "last_edit_user_id"}); err != nil {
			return "", err
		}
		err = as.questionCommon.UpdatePostTime(ctx, req.QuestionID)
		if err != nil {
			return insertData.ID, err
		}
		as.notificationUpdateAnswer(ctx, questionInfo.UserID, insertData.ID, req.UserID)
		revisionDTO.Status = entity.RevisionReviewPassStatus
	}

	infoJSON, _ := json.Marshal(insertData)
	revisionDTO.Content = string(infoJSON)
	revisionID, err := as.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return insertData.ID, err
	}
	if canUpdate {
		as.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         insertData.ID,
			OriginalObjectID: insertData.ID,
			ActivityTypeKey:  constant.ActAnswerEdited,
			RevisionID:       revisionID,
		})
		as.eventQueueService.Send(ctx, schema.NewEvent(constant.EventAnswerUpdate, req.UserID).TID(insertData.ID).
			AID(insertData.ID, insertData.UserID))
	}

	return insertData.ID, nil
}

// AcceptAnswer accept answer
func (as *AnswerService) AcceptAnswer(ctx context.Context, req *schema.AcceptAnswerReq) (err error) {
	// find question
	questionInfo, exist, err := as.questionRepo.GetQuestion(ctx, req.QuestionID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.QuestionNotFound)
	}
	questionInfo.ID = uid.DeShortID(questionInfo.ID)
	if questionInfo.AcceptedAnswerID == req.AnswerID {
		return nil
	}

	// find answer
	var acceptedAnswerInfo *entity.Answer
	if len(req.AnswerID) > 1 {
		acceptedAnswerInfo, exist, err = as.answerRepo.GetByID(ctx, req.AnswerID)
		if err != nil {
			return err
		}
		if !exist {
			return errors.BadRequest(reason.AnswerNotFound)
		}
		acceptedAnswerInfo.ID = uid.DeShortID(acceptedAnswerInfo.ID)
	}

	// update answers status
	if err = as.answerRepo.UpdateAcceptedStatus(ctx, req.AnswerID, req.QuestionID); err != nil {
		return err
	}

	// update question status
	err = as.questionCommon.UpdateAccepted(ctx, req.QuestionID, req.AnswerID)
	if err != nil {
		log.Error("UpdateLastAnswer error", err.Error())
	}

	var oldAnswerInfo *entity.Answer
	if len(questionInfo.AcceptedAnswerID) > 1 {
		oldAnswerInfo, _, err = as.answerRepo.GetByID(ctx, questionInfo.AcceptedAnswerID)
		if err != nil {
			return err
		}
		oldAnswerInfo.ID = uid.DeShortID(oldAnswerInfo.ID)
	}

	if acceptedAnswerInfo != nil {
		as.eventQueueService.Send(ctx, schema.NewEvent(constant.EventQuestionAccept, req.UserID).TID(acceptedAnswerInfo.ID).
			QID(questionInfo.ID, questionInfo.UserID).AID(acceptedAnswerInfo.ID, acceptedAnswerInfo.UserID))
	}

	as.updateAnswerRank(ctx, req.UserID, questionInfo, acceptedAnswerInfo, oldAnswerInfo)
	return nil
}

func (as *AnswerService) updateAnswerRank(ctx context.Context, userID string,
	questionInfo *entity.Question, newAnswerInfo *entity.Answer, oldAnswerInfo *entity.Answer,
) {
	// if this question is already been answered, should cancel old answer rank
	if oldAnswerInfo != nil {
		err := as.answerActivityService.CancelAcceptAnswer(ctx, userID,
			questionInfo.AcceptedAnswerID, questionInfo.ID, questionInfo.UserID, oldAnswerInfo.UserID)
		if err != nil {
			log.Error(err)
		}
	}
	if newAnswerInfo != nil {
		err := as.answerActivityService.AcceptAnswer(ctx, userID, newAnswerInfo.ID,
			questionInfo.ID, questionInfo.UserID, newAnswerInfo.UserID, newAnswerInfo.UserID == questionInfo.UserID)
		if err != nil {
			log.Error(err)
		}
	}
}

func (as *AnswerService) Get(ctx context.Context, answerID, loginUserID string) (*schema.AnswerInfo, *schema.QuestionInfoResp, bool, error) {
	answerInfo, has, err := as.answerRepo.GetByID(ctx, answerID)
	if err != nil {
		return nil, nil, has, err
	}
	info := as.ShowFormat(ctx, answerInfo)
	// todo questionFunc
	questionInfo, err := as.questionCommon.Info(ctx, answerInfo.QuestionID, loginUserID)
	if err != nil {
		return nil, nil, has, err
	}
	// todo UserFunc

	userIds := make([]string, 0)
	userIds = append(userIds, answerInfo.UserID)
	userIds = append(userIds, answerInfo.LastEditUserID)
	userInfoMap, err := as.userCommon.BatchUserBasicInfoByID(ctx, userIds)
	if err != nil {
		return nil, nil, has, err
	}

	_, ok := userInfoMap[answerInfo.UserID]
	if ok {
		info.UserInfo = userInfoMap[answerInfo.UserID]
	}
	_, ok = userInfoMap[answerInfo.LastEditUserID]
	if ok {
		info.UpdateUserInfo = userInfoMap[answerInfo.LastEditUserID]
	}

	if loginUserID == "" {
		return info, questionInfo, has, nil
	}

	info.VoteStatus = as.voteRepo.GetVoteStatus(ctx, answerID, loginUserID)

	collectedMap, err := as.collectionCommon.SearchObjectCollected(ctx, loginUserID, []string{answerInfo.ID})
	if err != nil {
		return nil, nil, has, err
	}
	if len(collectedMap) > 0 {
		info.Collected = true
	}

	return info, questionInfo, has, nil
}

func (as *AnswerService) GetDetail(ctx context.Context, answerID string) (*schema.AnswerInfo, error) {
	answerInfo, has, err := as.answerRepo.GetByID(ctx, answerID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.BadRequest(reason.AnswerNotFound)
	}
	info := as.ShowFormat(ctx, answerInfo)
	return info, nil
}

func (as *AnswerService) GetCountByUserIDQuestionID(ctx context.Context, userId string, questionId string) (ids []string, err error) {
	return as.answerRepo.GetIDsByUserIDAndQuestionID(ctx, userId, questionId)
}

func (as *AnswerService) AdminSetAnswerStatus(ctx context.Context, req *schema.AdminUpdateAnswerStatusReq) error {
	setStatus, ok := entity.AdminAnswerSearchStatus[req.Status]
	if !ok {
		return errors.BadRequest(reason.RequestFormatError)
	}
	answerInfo, exist, err := as.answerRepo.GetAnswer(ctx, req.AnswerID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.AnswerNotFound)
	}

	if setStatus == entity.AnswerStatusDeleted {
		if err := as.RemoveAnswer(ctx, &schema.RemoveAnswerReq{
			ID:        req.AnswerID,
			UserID:    req.UserID,
			CanDelete: true,
		}); err != nil {
			return err
		}

		msg := &schema.NotificationMsg{}
		msg.ObjectID = answerInfo.ID
		msg.Type = schema.NotificationTypeInbox
		msg.ReceiverUserID = answerInfo.UserID
		msg.TriggerUserID = answerInfo.UserID
		msg.ObjectType = constant.AnswerObjectType
		msg.NotificationAction = constant.NotificationYourAnswerWasDeleted
		as.notificationQueueService.Send(ctx, msg)
	}

	// recover
	if setStatus == entity.QuestionStatusAvailable && answerInfo.Status == entity.QuestionStatusDeleted {
		if err := as.RecoverAnswer(ctx, &schema.RecoverAnswerReq{
			AnswerID: req.AnswerID,
			UserID:   req.UserID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (as *AnswerService) SearchList(ctx context.Context, req *schema.AnswerListReq) ([]*schema.AnswerInfo, int64, error) {
	list := make([]*schema.AnswerInfo, 0)
	dbSearch := entity.AnswerSearch{}
	dbSearch.QuestionID = req.QuestionID
	dbSearch.Page = req.Page
	dbSearch.PageSize = req.PageSize
	dbSearch.Order = req.Order
	dbSearch.IncludeDeleted = req.CanDelete
	dbSearch.LoginUserID = req.UserID
	answerOriginalList, count, err := as.answerRepo.SearchList(ctx, &dbSearch)
	if err != nil {
		return list, count, err
	}
	answerList, err := as.SearchFormatInfo(ctx, answerOriginalList, req)
	if err != nil {
		return answerList, count, err
	}
	return answerList, count, nil
}

func (as *AnswerService) SearchFormatInfo(ctx context.Context, answers []*entity.Answer, req *schema.AnswerListReq) (
	[]*schema.AnswerInfo, error) {
	list := make([]*schema.AnswerInfo, 0)
	objectIDs := make([]string, 0)
	userIDs := make([]string, 0)
	for _, info := range answers {
		item := as.ShowFormat(ctx, info)
		list = append(list, item)
		objectIDs = append(objectIDs, info.ID)
		userIDs = append(userIDs, info.UserID, info.LastEditUserID)
	}

	userInfoMap, err := as.userCommon.BatchUserBasicInfoByID(ctx, userIDs)
	if err != nil {
		return list, err
	}
	for _, item := range list {
		item.UserInfo = userInfoMap[item.UserID]
		item.UpdateUserInfo = userInfoMap[item.UpdateUserID]
	}
	if len(req.UserID) == 0 {
		return list, nil
	}

	collectedMap, err := as.collectionCommon.SearchObjectCollected(ctx, req.UserID, objectIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		item.VoteStatus = as.voteRepo.GetVoteStatus(ctx, item.ID, req.UserID)
		item.Collected = collectedMap[item.ID]
		item.MemberActions = permission.GetAnswerPermission(ctx,
			req.UserID,
			item.UserID,
			item.Status,
			req.CanEdit,
			req.CanDelete,
			req.CanRecover)
	}
	return list, nil
}

func (as *AnswerService) ShowFormat(ctx context.Context, data *entity.Answer) *schema.AnswerInfo {
	return as.AnswerCommon.ShowFormat(ctx, data)
}

func (as *AnswerService) notificationUpdateAnswer(ctx context.Context, questionUserID, answerID, answerUserID string) {
	// If the answer is updated by me, there is no notification for myself.
	// equivalent behaviour as AnswerService.notificationAnswerTheQuestion
	if questionUserID == answerUserID {
		return
	}
	msg := &schema.NotificationMsg{
		TriggerUserID:  answerUserID,
		ReceiverUserID: questionUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       answerID,
	}
	msg.ObjectType = constant.AnswerObjectType
	msg.NotificationAction = constant.NotificationUpdateAnswer
	as.notificationQueueService.Send(ctx, msg)
}

func (as *AnswerService) notificationAnswerTheQuestion(ctx context.Context,
	questionUserID, questionID, answerID, answerUserID, questionTitle, answerSummary string) {
	// If the question is answered by me, there is no notification for myself.
	if questionUserID == answerUserID {
		return
	}
	msg := &schema.NotificationMsg{
		TriggerUserID:  answerUserID,
		ReceiverUserID: questionUserID,
		Type:           schema.NotificationTypeInbox,
		ObjectID:       answerID,
	}
	msg.ObjectType = constant.AnswerObjectType
	msg.NotificationAction = constant.NotificationAnswerTheQuestion
	as.notificationQueueService.Send(ctx, msg)

	receiverUserInfo, exist, err := as.userRepo.GetByUserID(ctx, questionUserID)
	if err != nil {
		log.Error(err)
		return
	}
	if !exist {
		log.Warnf("user %s not found", questionUserID)
		return
	}

	externalNotificationMsg := &schema.ExternalNotificationMsg{
		ReceiverUserID: receiverUserInfo.ID,
		ReceiverEmail:  receiverUserInfo.EMail,
		ReceiverLang:   receiverUserInfo.Language,
	}
	rawData := &schema.NewAnswerTemplateRawData{
		QuestionTitle:   questionTitle,
		QuestionID:      questionID,
		AnswerID:        answerID,
		AnswerSummary:   answerSummary,
		UnsubscribeCode: token.GenerateToken(),
	}
	answerUser, _, _ := as.userCommon.GetUserBasicInfoByID(ctx, answerUserID)
	if answerUser != nil {
		rawData.AnswerUserDisplayName = answerUser.DisplayName
	}
	externalNotificationMsg.NewAnswerTemplateRawData = rawData
	as.externalNotificationQueueService.Send(ctx, externalNotificationMsg)
}

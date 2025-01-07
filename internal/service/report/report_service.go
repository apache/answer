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

package report

import (
	"encoding/json"
	"github.com/apache/answer/internal/service/event_queue"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/comment_common"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/report_common"
	"github.com/apache/answer/internal/service/report_handle"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/obj"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
)

// ReportService user service
type ReportService struct {
	reportRepo        report_common.ReportRepo
	objectInfoService *object_info.ObjService
	commonUser        *usercommon.UserCommon
	answerRepo        answercommon.AnswerRepo
	questionRepo      questioncommon.QuestionRepo
	commentCommonRepo comment_common.CommentCommonRepo
	reportHandle      *report_handle.ReportHandle
	configService     *config.ConfigService
	eventQueueService event_queue.EventQueueService
}

// NewReportService new report service
func NewReportService(
	reportRepo report_common.ReportRepo,
	objectInfoService *object_info.ObjService,
	commonUser *usercommon.UserCommon,
	answerRepo answercommon.AnswerRepo,
	questionRepo questioncommon.QuestionRepo,
	commentCommonRepo comment_common.CommentCommonRepo,
	reportHandle *report_handle.ReportHandle,
	configService *config.ConfigService,
	eventQueueService event_queue.EventQueueService,
) *ReportService {
	return &ReportService{
		reportRepo:        reportRepo,
		objectInfoService: objectInfoService,
		commonUser:        commonUser,
		answerRepo:        answerRepo,
		questionRepo:      questionRepo,
		commentCommonRepo: commentCommonRepo,
		reportHandle:      reportHandle,
		configService:     configService,
		eventQueueService: eventQueueService,
	}
}

// AddReport add report
func (rs *ReportService) AddReport(ctx context.Context, req *schema.AddReportReq) (err error) {
	objectTypeNumber, err := obj.GetObjectTypeNumberByObjectID(req.ObjectID)
	if err != nil {
		return err
	}

	objInfo, err := rs.objectInfoService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return err
	}
	if objInfo.IsDeleted() {
		return errors.BadRequest(reason.NewObjectAlreadyDeleted)
	}

	cf, err := rs.configService.GetConfigByID(ctx, req.ReportType)
	if err != nil || cf == nil {
		return errors.BadRequest(reason.ReportNotFound)
	}
	if cf.Key == constant.ReasonADuplicate && !checker.IsURL(req.Content) {
		return errors.BadRequest(reason.InvalidURLError)
	}

	report := &entity.Report{
		UserID:         req.UserID,
		ReportedUserID: objInfo.ObjectCreatorUserID,
		ObjectID:       req.ObjectID,
		ObjectType:     objectTypeNumber,
		ReportType:     req.ReportType,
		Content:        req.Content,
		Status:         entity.ReportStatusPending,
	}
	err = rs.reportRepo.AddReport(ctx, report)
	if err != nil {
		return err
	}
	rs.sendEvent(ctx, report, objInfo)
	return nil
}

// GetUnreviewedReportPostPage get unreviewed report post page
func (rs *ReportService) GetUnreviewedReportPostPage(ctx context.Context, req *schema.GetUnreviewedReportPostPageReq) (
	pageModel *pager.PageModel, err error) {
	if !req.IsAdmin {
		return pager.NewPageModel(0, make([]*schema.GetReportListPageResp, 0)), nil
	}
	lang := handler.GetLangByCtx(ctx)
	reports, total, err := rs.reportRepo.GetReportListPage(ctx, &schema.GetReportListPageDTO{
		Page:     req.Page,
		PageSize: 1,
		Status:   entity.ReportStatusPending,
	})
	if err != nil {
		return
	}

	resp := make([]*schema.GetReportListPageResp, 0)
	for _, report := range reports {
		info, err := rs.objectInfoService.GetUnreviewedRevisionInfo(ctx, report.ObjectID)
		if err != nil {
			log.Errorf("GetUnreviewedRevisionInfo failed, err: %v", err)
			continue
		}

		r := &schema.GetReportListPageResp{
			FlagID:           report.ID,
			CreatedAt:        info.CreatedAt,
			ObjectID:         info.ObjectID,
			ObjectType:       info.ObjectType,
			QuestionID:       info.QuestionID,
			AnswerID:         info.AnswerID,
			CommentID:        info.CommentID,
			Title:            info.Title,
			UrlTitle:         htmltext.UrlTitle(info.Title),
			OriginalText:     info.Content,
			ParsedText:       info.Html,
			AnswerCount:      info.AnswerCount,
			AnswerAccepted:   info.AnswerAccepted,
			Tags:             info.Tags,
			SubmitAt:         report.CreatedAt.Unix(),
			ObjectStatus:     info.Status,
			ObjectShowStatus: info.ShowStatus,
			ReasonContent:    report.Content,
		}

		// get user info
		userInfo, exists, e := rs.commonUser.GetUserBasicInfoByID(ctx, info.ObjectCreatorUserID)
		if e != nil {
			log.Errorf("user not found by id: %s, err: %v", info.ObjectCreatorUserID, e)
		}
		if exists {
			_ = copier.Copy(&r.AuthorUserInfo, userInfo)
		}

		// get submitter info
		submitter, exists, e := rs.commonUser.GetUserBasicInfoByID(ctx, report.ReportedUserID)
		if e != nil {
			log.Errorf("user not found by id: %s, err: %v", info.ObjectCreatorUserID, e)
		}
		if exists {
			_ = copier.Copy(&r.SubmitterUser, submitter)
		}

		if report.ReportType > 0 {
			r.Reason = &schema.ReasonItem{ReasonType: report.ReportType}
			cf, err := rs.configService.GetConfigByID(ctx, report.ReportType)
			if err != nil {
				log.Error(err)
			} else {
				_ = json.Unmarshal([]byte(cf.Value), r.Reason)
				r.Reason.Translate(cf.Key, lang)
			}
		}
		resp = append(resp, r)
	}
	return pager.NewPageModel(total, resp), nil
}

// ReviewReport review report
func (rs *ReportService) ReviewReport(ctx context.Context, req *schema.ReviewReportReq) (err error) {
	report, exist, err := rs.reportRepo.GetByID(ctx, req.FlagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ReportNotFound)
	}
	// check if handle or not
	if report.Status != entity.ReportStatusPending {
		return nil
	}

	// ignore this report
	if req.OperationType == constant.ReportOperationIgnoreReport {
		return rs.reportRepo.UpdateStatus(ctx, report.ID, entity.ReportStatusIgnore)
	}

	if err = rs.reportHandle.UpdateReportedObject(ctx, report, req); err != nil {
		return
	}

	return rs.reportRepo.UpdateStatus(ctx, report.ID, entity.ReportStatusCompleted)
}

func (rs *ReportService) sendEvent(ctx context.Context,
	report *entity.Report, objectInfo *schema.SimpleObjectInfo) {
	var event *schema.EventMsg
	switch objectInfo.ObjectType {
	case constant.QuestionObjectType:
		event = schema.NewEvent(constant.EventQuestionFlag, report.UserID).TID(objectInfo.QuestionID).
			QID(objectInfo.QuestionID, objectInfo.ObjectCreatorUserID)
	case constant.AnswerObjectType:
		event = schema.NewEvent(constant.EventAnswerFlag, report.UserID).TID(objectInfo.AnswerID).
			AID(objectInfo.AnswerID, objectInfo.ObjectCreatorUserID)
	case constant.CommentObjectType:
		event = schema.NewEvent(constant.EventCommentFlag, report.UserID).TID(objectInfo.CommentID).
			CID(objectInfo.CommentID, objectInfo.ObjectCreatorUserID)
	default:
		return
	}
	rs.eventQueueService.Send(ctx, event)
}

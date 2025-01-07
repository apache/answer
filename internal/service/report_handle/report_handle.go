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

package report_handle

import (
	"context"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/obj"
)

type ReportHandle struct {
	questionService *content.QuestionService
	answerService   *content.AnswerService
	commentService  *comment.CommentService
}

func NewReportHandle(
	questionService *content.QuestionService,
	answerService *content.AnswerService,
	commentService *comment.CommentService,
) *ReportHandle {
	return &ReportHandle{
		questionService: questionService,
		answerService:   answerService,
		commentService:  commentService,
	}
}

// UpdateReportedObject this handle object status
func (rh *ReportHandle) UpdateReportedObject(ctx context.Context,
	report *entity.Report, req *schema.ReviewReportReq) (err error) {
	objectKey, err := obj.GetObjectTypeStrByObjectID(report.ObjectID)
	if err != nil {
		return err
	}
	switch objectKey {
	case constant.QuestionObjectType:
		err = rh.updateReportedQuestionReport(ctx, report, req)
	case constant.AnswerObjectType:
		err = rh.updateReportedAnswerReport(ctx, report, req)
	case constant.CommentObjectType:
		err = rh.updateReportedCommentReport(ctx, report, req)
	}
	return
}

func (rh *ReportHandle) updateReportedQuestionReport(ctx context.Context,
	report *entity.Report, req *schema.ReviewReportReq) (err error) {
	switch req.OperationType {
	case constant.ReportOperationUnlistPost:
		err = rh.questionService.OperationQuestion(ctx, &schema.OperationQuestionReq{
			ID: report.ObjectID, Operation: schema.QuestionOperationHide, UserID: req.UserID})
	case constant.ReportOperationDeletePost:
		err = rh.questionService.RemoveQuestion(ctx, &schema.RemoveQuestionReq{
			ID: report.ObjectID, UserID: req.UserID, IsAdmin: true})
	case constant.ReportOperationClosePost:
		err = rh.questionService.CloseQuestion(ctx, &schema.CloseQuestionReq{
			ID:        report.ObjectID,
			CloseType: req.CloseType,
			CloseMsg:  req.CloseMsg,
			UserID:    req.UserID,
		})
	case constant.ReportOperationEditPost:
		_, err = rh.questionService.UpdateQuestion(ctx, &schema.QuestionUpdate{
			ID:           report.ObjectID,
			Title:        req.Title,
			Content:      req.Content,
			HTML:         converter.Markdown2HTML(req.Content),
			Tags:         req.Tags,
			UserID:       req.UserID,
			NoNeedReview: true,
		})
	}
	return
}

func (rh *ReportHandle) updateReportedAnswerReport(ctx context.Context, report *entity.Report, req *schema.ReviewReportReq) (err error) {
	switch req.OperationType {
	case constant.ReportOperationDeletePost:
		err = rh.answerService.RemoveAnswer(ctx, &schema.RemoveAnswerReq{
			ID: report.ObjectID, UserID: req.UserID})
	case constant.ReportOperationEditPost:
		_, err = rh.answerService.Update(ctx, &schema.AnswerUpdateReq{
			ID:           report.ObjectID,
			Title:        req.Title,
			Content:      req.Content,
			HTML:         converter.Markdown2HTML(req.Content),
			UserID:       req.UserID,
			NoNeedReview: true,
		})
	}
	return nil
}

func (rh *ReportHandle) updateReportedCommentReport(ctx context.Context, report *entity.Report, req *schema.ReviewReportReq) (err error) {
	switch req.OperationType {
	case constant.ReportOperationDeletePost:
		err = rh.commentService.RemoveComment(ctx, &schema.RemoveCommentReq{
			CommentID: report.ObjectID, UserID: req.UserID})
	case constant.ReportOperationEditPost:
		_, err = rh.commentService.UpdateComment(ctx, &schema.UpdateCommentReq{
			CommentID:    report.ObjectID,
			OriginalText: req.Content,
			ParsedText:   converter.Markdown2HTML(req.Content),
			UserID:       req.UserID,
		})
	}
	return nil
}

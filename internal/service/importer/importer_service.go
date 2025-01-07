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

package importer

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/rank"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// ImporterService importer service
type ImporterService struct {
	questionService *content.QuestionService
	rankService     *rank.RankService
	userCommon      *usercommon.UserCommon
}

// NewRankService new rank service
func NewImporterService(
	questionService *content.QuestionService,
	rankService *rank.RankService,
	userCommon *usercommon.UserCommon) *ImporterService {
	return &ImporterService{
		questionService: questionService,
		rankService:     rankService,
		userCommon:      userCommon,
	}
}

type ImporterFunc struct {
	importerService *ImporterService
}

func (ipfunc *ImporterFunc) AddQuestion(ctx context.Context, questionInfo plugin.QuestionImporterInfo) (err error) {
	ipfunc.importerService.ImportQuestion(ctx, questionInfo)
	return nil
}

func (ip *ImporterService) NewImporterFunc() plugin.ImporterFunc {
	return &ImporterFunc{importerService: ip}
}

func (ip *ImporterService) ImportQuestion(ctx context.Context, questionInfo plugin.QuestionImporterInfo) (err error) {
	req := &schema.QuestionAdd{}
	errFields := make([]*validator.FormErrorField, 0)
	// To limit rate, remove the following code from comment: Part 1/2
	// reject, rejectKey := ipc.rateLimitMiddleware.DuplicateRequestRejection(ctx, req)
	// if reject {
	// 	return
	// }
	userInfo, exist, err := ip.userCommon.GetByEmail(ctx, questionInfo.UserEmail)
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	if !exist {
		return fmt.Errorf("User not found")
	}

	// To limit rate, remove the following code from comment: Part 2/2
	// defer func() {
	// 	// If status is not 200 means that the bad request has been returned, so the record should be cleared
	// 	if ctx.Writer.Status() != http.StatusOK {
	// 		ipc.rateLimitMiddleware.DuplicateRequestClear(ctx, rejectKey)
	// 	}
	// }()
	req.UserID = userInfo.ID
	req.Title = questionInfo.Title
	req.Content = questionInfo.Content
	req.HTML = "<p>" + questionInfo.Content + "</p>"
	req.Tags = make([]*schema.TagItem, len(questionInfo.Tags))
	for i, tag := range questionInfo.Tags {
		req.Tags[i] = &schema.TagItem{
			SlugName:    tag,
			DisplayName: tag,
		}
	}
	canList, requireRanks, err := ip.rankService.CheckOperationPermissionsForRanks(ctx, req.UserID, []string{
		permission.QuestionAdd,
		permission.QuestionEdit,
		permission.QuestionDelete,
		permission.QuestionClose,
		permission.QuestionReopen,
		permission.TagUseReservedTag,
		permission.TagAdd,
		permission.LinkUrlLimit,
	})
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	req.CanAdd = canList[0]
	req.CanEdit = canList[1]
	req.CanDelete = canList[2]
	req.CanClose = canList[3]
	req.CanReopen = canList[4]
	req.CanUseReservedTag = canList[5]
	req.CanAddTag = canList[6]
	if !req.CanAdd {
		log.Errorf("error: %v", err)
		return err
	}
	hasNewTag, err := ip.questionService.HasNewTag(ctx.(*gin.Context), req.Tags)
	if err != nil {
		log.Errorf("error: %v", err)
		return err
	}
	if !req.CanAddTag && hasNewTag {
		lang := handler.GetLang(ctx.(*gin.Context))
		msg := translator.TrWithData(lang, reason.NoEnoughRankToOperate, &schema.PermissionTrTplData{Rank: requireRanks[6]})
		log.Errorf("error: %v", msg)
		return errors.BadRequest(msg)
	}

	errList, err := ip.questionService.CheckAddQuestion(ctx, req)
	if err != nil {
		errlist, ok := errList.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}
	if len(errFields) > 0 {
		return errors.BadRequest(reason.RequestFormatError)
	}
	ginCtx := ctx.(*gin.Context)
	req.UserAgent = ginCtx.GetHeader("User-Agent")
	req.IP = ginCtx.ClientIP()
	resp, err := ip.questionService.AddQuestion(ctx, req)
	if err != nil {
		errlist, ok := resp.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}

	if len(errFields) > 0 {
		log.Errorf("error: RequestFormatError")
		return errors.BadRequest(reason.RequestFormatError)
	}
	log.Info("Add Question Successfully")
	return nil
}

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

package templaterender

import (
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/schema"
	"github.com/jinzhu/copier"
	"golang.org/x/net/context"
)

func (q *TemplateRenderController) TagList(ctx context.Context, req *schema.GetTagWithPageReq) (resp *pager.PageModel, err error) {
	resp, err = q.tagService.GetTagWithPage(ctx, req)
	if err != nil {
		return
	}
	return
}

func (q *TemplateRenderController) TagInfo(ctx context.Context, req *schema.GetTamplateTagInfoReq) (resp *schema.GetTagResp, questionList []*schema.QuestionPageResp, questionCount int64, err error) {
	dto := &schema.GetTagInfoReq{}
	_ = copier.Copy(dto, req)
	resp, err = q.tagService.GetTagInfo(ctx, dto)
	if err != nil {
		return
	}
	searchQuestion := &schema.QuestionPageReq{}
	searchQuestion.Page = req.Page
	searchQuestion.PageSize = req.PageSize
	searchQuestion.OrderCond = "newest"
	searchQuestion.Tag = req.Name
	searchQuestion.LoginUserID = req.UserID
	questionList, questionCount, err = q.questionService.GetQuestionPage(ctx, searchQuestion)
	if err != nil {
		return
	}
	return resp, questionList, questionCount, err
}

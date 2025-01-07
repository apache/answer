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
	"math"

	"github.com/apache/answer/internal/service/content"
	questioncommon "github.com/apache/answer/internal/service/question_common"

	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/google/wire"

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/tag"
)

// ProviderSetTemplateRenderController is template render controller providers.
var ProviderSetTemplateRenderController = wire.NewSet(
	NewTemplateRenderController,
)

type TemplateRenderController struct {
	questionService *content.QuestionService
	userService     *content.UserService
	tagService      *tag.TagService
	answerService   *content.AnswerService
	commentService  *comment.CommentService
	siteInfoService siteinfo_common.SiteInfoCommonService
	questionRepo    questioncommon.QuestionRepo
}

func NewTemplateRenderController(
	questionService *content.QuestionService,
	userService *content.UserService,
	tagService *tag.TagService,
	answerService *content.AnswerService,
	commentService *comment.CommentService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	questionRepo questioncommon.QuestionRepo,
) *TemplateRenderController {
	return &TemplateRenderController{
		questionService: questionService,
		userService:     userService,
		tagService:      tagService,
		answerService:   answerService,
		commentService:  commentService,
		questionRepo:    questionRepo,
		siteInfoService: siteInfoService,
	}
}

// Paginator page
// page : now page
// pageSize : Number per page
// nums : Total
// Returns the contents of the page in the format of 1, 2, 3, 4, and 5. If the contents are less than 5 pages, the page number is returned
func Paginator(page, pageSize int, nums int64) *schema.Paginator {
	if pageSize == 0 {
		pageSize = 10
	}

	var prevpage int //Previous page address
	var nextpage int //Address on the last page
	//Generate the total number of pages based on the total number of nums and the number of prepage pages
	totalpages := int(math.Ceil(float64(nums) / float64(pageSize))) //Total number of Pages
	if page > totalpages {
		page = totalpages
	}
	if page <= 0 {
		page = 1
	}
	var pages []int
	switch {
	case page >= totalpages-5 && totalpages > 5: //The last 5 pages
		start := totalpages - 5 + 1
		prevpage = page - 1
		nextpage = int(math.Min(float64(totalpages), float64(page+1)))
		pages = make([]int, 5)
		for i := range pages {
			pages[i] = start + i
		}
	case page >= 3 && totalpages > 5:
		start := page - 3 + 1
		pages = make([]int, 5)
		prevpage = page - 3
		for i := range pages {
			pages[i] = start + i
		}
		prevpage = page - 1
		nextpage = page + 1
	default:
		pages = make([]int, int(math.Min(5, float64(totalpages))))
		for i := range pages {
			pages[i] = i + 1
		}
		prevpage = int(math.Max(float64(1), float64(page-1)))
		nextpage = page + 1
	}
	paginator := &schema.Paginator{}
	paginator.Pages = pages
	paginator.Totalpages = totalpages
	paginator.Prevpage = prevpage
	paginator.Nextpage = nextpage
	paginator.Currpage = page
	return paginator
}

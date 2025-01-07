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
	"html/template"
	"math"
	"net/http"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/schema"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

func (t *TemplateRenderController) Index(ctx *gin.Context, req *schema.QuestionPageReq) ([]*schema.QuestionPageResp, int64, error) {
	return t.questionService.GetQuestionPage(ctx, req)
}

func (t *TemplateRenderController) QuestionDetail(ctx *gin.Context, id string) (resp *schema.QuestionInfoResp, err error) {
	return t.questionService.GetQuestion(ctx, id, "", schema.QuestionPermission{})
}

func (t *TemplateRenderController) Sitemap(ctx *gin.Context) {
	general, err := t.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error("get site general failed:", err)
		return
	}
	siteInfo, err := t.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error("get site GetSiteSeo failed:", err)
		return
	}

	questions, err := t.questionRepo.SitemapQuestions(ctx, 1, constant.SitemapMaxSize)
	if err != nil {
		log.Errorf("get sitemap questions failed: %s", err)
		return
	}

	ctx.Header("Content-Type", "application/xml")
	if len(questions) < constant.SitemapMaxSize {
		ctx.HTML(
			http.StatusOK, "sitemap.xml", gin.H{
				"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
				"list":      questions,
				"general":   general,
				"hastitle": siteInfo.Permalink == constant.PermalinkQuestionIDAndTitle ||
					siteInfo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID,
			},
		)
		return
	}

	questionNum, err := t.questionRepo.GetQuestionCount(ctx)
	if err != nil {
		log.Error("GetQuestionCount error", err)
		return
	}
	var pageList []int
	totalPages := int(math.Ceil(float64(questionNum) / float64(constant.SitemapMaxSize)))
	for i := 1; i <= totalPages; i++ {
		pageList = append(pageList, i)
	}
	ctx.HTML(
		http.StatusOK, "sitemap-list.xml", gin.H{
			"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"page":      pageList,
			"general":   general,
		},
	)
}

func (t *TemplateRenderController) OpenSearch(ctx *gin.Context) {
	general, err := t.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error("get site general failed:", err)
		return
	}

	favicon := general.SiteUrl + "/favicon.ico"
	branding, err := t.siteInfoService.GetSiteBranding(ctx)
	if err == nil && len(branding.Favicon) > 0 {
		favicon = branding.Favicon
	}

	ctx.Header("Content-Type", "application/xml")
	ctx.HTML(
		http.StatusOK, "opensearch.xml", gin.H{
			"general": general,
			"favicon": favicon,
		},
	)
}

func (t *TemplateRenderController) SitemapPage(ctx *gin.Context, page int) error {
	general, err := t.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error("get site general failed:", err)
		return err
	}
	siteInfo, err := t.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error("get site GetSiteSeo failed:", err)
		return err
	}

	questions, err := t.questionRepo.SitemapQuestions(ctx, page, constant.SitemapMaxSize)
	if err != nil {
		log.Errorf("get sitemap questions failed: %s", err)
		return err
	}
	ctx.Header("Content-Type", "application/xml")
	ctx.HTML(
		http.StatusOK, "sitemap.xml", gin.H{
			"xmlHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"list":      questions,
			"general":   general,
			"hastitle": siteInfo.Permalink == constant.PermalinkQuestionIDAndTitle ||
				siteInfo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID,
		},
	)
	return nil
}

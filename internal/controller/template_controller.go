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

package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/event_queue"
	"github.com/apache/answer/plugin"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	templaterender "github.com/apache/answer/internal/controller/template_render"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/obj"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/ui"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

var SiteUrl = ""

type TemplateController struct {
	scriptPath               []string
	cssPath                  string
	templateRenderController *templaterender.TemplateRenderController
	siteInfoService          siteinfo_common.SiteInfoCommonService
	eventQueueService        event_queue.EventQueueService
	userService              *content.UserService
	questionService          *content.QuestionService
}

// NewTemplateController new controller
func NewTemplateController(
	templateRenderController *templaterender.TemplateRenderController,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	eventQueueService event_queue.EventQueueService,
	userService *content.UserService,
	questionService *content.QuestionService,
) *TemplateController {
	script, css := GetStyle()
	return &TemplateController{
		scriptPath:               script,
		cssPath:                  css,
		templateRenderController: templateRenderController,
		siteInfoService:          siteInfoService,
		eventQueueService:        eventQueueService,
		userService:              userService,
		questionService:          questionService,
	}
}
func GetStyle() (script []string, css string) {
	file, err := ui.Build.ReadFile("build/index.html")
	if err != nil {
		return
	}
	scriptRegexp := regexp.MustCompile(`<script defer="defer" src="([^"]*)"></script>`)
	scriptData := scriptRegexp.FindAllStringSubmatch(string(file), -1)
	for _, s := range scriptData {
		if len(s) == 2 {
			script = append(script, s[1])
		}
	}

	cssRegexp := regexp.MustCompile(`<link href="(.*)" rel="stylesheet">`)
	cssListData := cssRegexp.FindStringSubmatch(string(file))
	if len(cssListData) == 2 {
		css = cssListData[1]
	}
	return
}
func (tc *TemplateController) SiteInfo(ctx *gin.Context) *schema.TemplateSiteInfoResp {
	var err error
	resp := &schema.TemplateSiteInfoResp{}
	resp.General, err = tc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	}
	SiteUrl = resp.General.SiteUrl
	resp.Interface, err = tc.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Branding, err = tc.siteInfoService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.SiteSeo, err = tc.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.CustomCssHtml, err = tc.siteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Year = fmt.Sprintf("%d", time.Now().Year())
	return resp
}

// Index question list
func (tc *TemplateController) Index(ctx *gin.Context) {
	req := &schema.QuestionPageReq{
		OrderCond: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	var page = req.Page

	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil || (len(data) == 0 && pager.ValPageOutOfRange(count, page, req.PageSize)) {
		tc.Page404(ctx)
		return
	}

	hotQuestionReq := &schema.QuestionPageReq{
		Page:      1,
		PageSize:  6,
		OrderCond: "hot",
		InDays:    7,
	}
	hotQuestion, _, _ := tc.templateRenderController.Index(ctx, hotQuestionReq)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = siteInfo.General.SiteUrl

	UrlUseTitle := siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle ||
		siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID

	siteInfo.Title = ""
	tc.html(ctx, http.StatusOK, "question.html", siteInfo, gin.H{
		"data":        data,
		"useTitle":    UrlUseTitle,
		"page":        templaterender.Paginator(page, req.PageSize, count),
		"path":        "questions",
		"hotQuestion": hotQuestion,
	})
}

func (tc *TemplateController) QuestionList(ctx *gin.Context) {
	req := &schema.QuestionPageReq{
		OrderCond: "newest",
	}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	var page = req.Page
	data, count, err := tc.templateRenderController.Index(ctx, req)
	if err != nil || (len(data) == 0 && pager.ValPageOutOfRange(count, page, req.PageSize)) {
		tc.Page404(ctx)
		return
	}

	hotQuestionReq := &schema.QuestionPageReq{
		Page:      1,
		PageSize:  6,
		OrderCond: "hot",
		InDays:    7,
	}
	hotQuestion, _, _ := tc.templateRenderController.Index(ctx, hotQuestionReq)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/questions", siteInfo.General.SiteUrl)
	if page > 1 {
		siteInfo.Canonical = fmt.Sprintf("%s/questions?page=%d", siteInfo.General.SiteUrl, page)
	}

	UrlUseTitle := siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle ||
		siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID

	siteInfo.Title = fmt.Sprintf("%s - %s", translator.Tr(handler.GetLang(ctx), constant.QuestionsTitleTrKey), siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "question.html", siteInfo, gin.H{
		"data":        data,
		"useTitle":    UrlUseTitle,
		"page":        templaterender.Paginator(page, req.PageSize, count),
		"hotQuestion": hotQuestion,
	})
}

func (tc *TemplateController) QuestionInfoRedirect(ctx *gin.Context, siteInfo *schema.TemplateSiteInfoResp, correctTitle bool) (jump bool, url string) {
	questionID := ctx.Param("id")
	title := ctx.Param("title")
	answerID := uid.DeShortID(title)
	titleIsAnswerID := false
	needChangeShortID := false

	objectType, err := obj.GetObjectTypeStrByObjectID(answerID)
	if err == nil && objectType == constant.AnswerObjectType {
		titleIsAnswerID = true
	}

	siteSeo, err := tc.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		return false, ""
	}
	isShortID := uid.IsShortID(questionID)
	if siteSeo.IsShortLink() {
		if !isShortID {
			questionID = uid.EnShortID(questionID)
			needChangeShortID = true
		}
		if titleIsAnswerID {
			answerID = uid.EnShortID(answerID)
		}
	} else {
		if isShortID {
			needChangeShortID = true
			questionID = uid.DeShortID(questionID)
		}
		if titleIsAnswerID {
			answerID = uid.DeShortID(answerID)
		}
	}

	if _, err := tc.templateRenderController.AnswerDetail(ctx, answerID); err != nil {
		answerID = ""
		titleIsAnswerID = false
	}

	url = fmt.Sprintf("%s/questions/%s", siteInfo.General.SiteUrl, questionID)
	if siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionID || siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDByShortID {
		if len(ctx.Request.URL.Query()) > 0 {
			url = fmt.Sprintf("%s?%s", url, ctx.Request.URL.RawQuery)
		}
		if needChangeShortID {
			return true, url
		}
		// not have title
		if titleIsAnswerID || len(title) == 0 {
			return false, ""
		}

		return true, url
	} else {
		detail, err := tc.templateRenderController.QuestionDetail(ctx, questionID)
		if err != nil {
			tc.Page404(ctx)
			return
		}
		url = fmt.Sprintf("%s/%s", url, htmltext.UrlTitle(detail.Title))
		if titleIsAnswerID {
			url = fmt.Sprintf("%s/%s", url, answerID)
		}

		if len(ctx.Request.URL.Query()) > 0 {
			url = fmt.Sprintf("%s?%s", url, ctx.Request.URL.RawQuery)
		}
		// have title
		if len(title) > 0 && !titleIsAnswerID && correctTitle {
			if needChangeShortID {
				return true, url
			}
			return false, ""
		}
		return true, url
	}
}

// QuestionInfo question and answers info
func (tc *TemplateController) QuestionInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	title := ctx.Param("title")
	answerid := ctx.Param("answerid")
	shareUsername := ctx.Query("share")
	if checker.IsQuestionsIgnorePath(id) {
		// if id == "ask" {
		file, err := ui.Build.ReadFile("build/index.html")
		if err != nil {
			log.Error(err)
			tc.Page404(ctx)
			return
		}
		ctx.Header("content-type", "text/html;charset=utf-8")
		ctx.String(http.StatusOK, string(file))
		return
	}

	correctTitle := false

	detail, err := tc.templateRenderController.QuestionDetail(ctx, id)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	if len(shareUsername) > 0 {
		userInfo, err := tc.userService.GetOtherUserInfoByUsername(
			ctx, &schema.GetOtherUserInfoByUsernameReq{Username: shareUsername})
		if err == nil {
			tc.eventQueueService.Send(ctx, schema.NewEvent(constant.EventUserShare, userInfo.ID).
				QID(id, detail.UserID).AID(answerid, ""))
		}
	}
	encodeTitle := htmltext.UrlTitle(detail.Title)
	if encodeTitle == title {
		correctTitle = true
	}

	siteInfo := tc.SiteInfo(ctx)
	jump, jumpurl := tc.QuestionInfoRedirect(ctx, siteInfo, correctTitle)
	if jump {
		ctx.Redirect(http.StatusFound, jumpurl)
		return
	}

	// answers
	answerReq := &schema.AnswerListReq{
		QuestionID: id,
		Order:      "",
		Page:       1,
		PageSize:   999,
		UserID:     "",
	}
	answers, answerCount, err := tc.templateRenderController.AnswerList(ctx, answerReq)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	// comments
	objectIDs := []string{uid.DeShortID(id)}
	for _, answer := range answers {
		answerID := uid.DeShortID(answer.ID)
		objectIDs = append(objectIDs, answerID)
	}
	comments, err := tc.templateRenderController.CommentList(ctx, objectIDs)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	UrlUseTitle := siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle ||
		siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID

	// related question
	userID := middleware.GetLoginUserIDFromContext(ctx)
	relatedQuestion, _, _ := tc.questionService.SimilarQuestion(ctx, id, userID)

	siteInfo.Canonical = fmt.Sprintf("%s/questions/%s/%s", siteInfo.General.SiteUrl, id, encodeTitle)
	if siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionID || siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDByShortID {
		siteInfo.Canonical = fmt.Sprintf("%s/questions/%s", siteInfo.General.SiteUrl, id)
	}
	jsonLD := &schema.QAPageJsonLD{}
	jsonLD.Context = "https://schema.org"
	jsonLD.Type = "QAPage"
	jsonLD.MainEntity.Type = "Question"
	jsonLD.MainEntity.Name = detail.Title
	jsonLD.MainEntity.Text = detail.HTML
	jsonLD.MainEntity.AnswerCount = int(answerCount)
	jsonLD.MainEntity.UpvoteCount = detail.VoteCount
	jsonLD.MainEntity.DateCreated = time.Unix(detail.CreateTime, 0)
	jsonLD.MainEntity.Author.Type = "Person"
	jsonLD.MainEntity.Author.Name = detail.UserInfo.DisplayName
	jsonLD.MainEntity.Author.URL = fmt.Sprintf("%s/users/%s", siteInfo.General.SiteUrl, detail.UserInfo.Username)
	answerList := make([]*schema.SuggestedAnswerItem, 0)
	for _, answer := range answers {
		if answer.Accepted == schema.AnswerAcceptedEnable {
			acceptedAnswerItem := &schema.AcceptedAnswerItem{}
			acceptedAnswerItem.Type = "Answer"
			acceptedAnswerItem.Text = answer.HTML
			acceptedAnswerItem.DateCreated = time.Unix(answer.CreateTime, 0)
			acceptedAnswerItem.UpvoteCount = answer.VoteCount
			acceptedAnswerItem.URL = fmt.Sprintf("%s/%s", siteInfo.Canonical, answer.ID)
			acceptedAnswerItem.Author.Type = "Person"
			acceptedAnswerItem.Author.Name = answer.UserInfo.DisplayName
			acceptedAnswerItem.Author.URL = fmt.Sprintf("%s/users/%s", siteInfo.General.SiteUrl, answer.UserInfo.Username)
			jsonLD.MainEntity.AcceptedAnswer = acceptedAnswerItem
		} else {
			item := &schema.SuggestedAnswerItem{}
			item.Type = "Answer"
			item.Text = answer.HTML
			item.DateCreated = time.Unix(answer.CreateTime, 0)
			item.UpvoteCount = answer.VoteCount
			item.URL = fmt.Sprintf("%s/%s", siteInfo.Canonical, answer.ID)
			item.Author.Type = "Person"
			item.Author.Name = answer.UserInfo.DisplayName
			item.Author.URL = fmt.Sprintf("%s/users/%s", siteInfo.General.SiteUrl, answer.UserInfo.Username)
			answerList = append(answerList, item)
		}
	}
	jsonLD.MainEntity.SuggestedAnswer = answerList
	jsonLDStr, err := json.Marshal(jsonLD)
	if err == nil {
		siteInfo.JsonLD = `<script data-react-helmet="true" type="application/ld+json">` + string(jsonLDStr) + ` </script>`
	}

	siteInfo.Description = htmltext.FetchExcerpt(detail.HTML, "...", 240)
	tags := make([]string, 0)
	for _, tag := range detail.Tags {
		tags = append(tags, tag.DisplayName)
	}
	siteInfo.Keywords = strings.ReplaceAll(strings.Trim(fmt.Sprint(tags), "[]"), " ", ",")
	siteInfo.Title = fmt.Sprintf("%s - %s", detail.Title, siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "question-detail.html", siteInfo, gin.H{
		"id":              id,
		"answerid":        answerid,
		"detail":          detail,
		"answers":         answers,
		"comments":        comments,
		"noindex":         detail.Show == entity.QuestionHide,
		"useTitle":        UrlUseTitle,
		"relatedQuestion": relatedQuestion,
	})
}

// TagList tags list
func (tc *TemplateController) TagList(ctx *gin.Context) {
	req := &schema.GetTagWithPageReq{
		PageSize: constant.DefaultPageSize,
		Page:     1,
	}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	data, err := tc.templateRenderController.TagList(ctx, req)
	if err != nil || pager.ValPageOutOfRange(data.Count, req.Page, req.PageSize) {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(req.Page, req.PageSize, data.Count)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/tags", siteInfo.General.SiteUrl)
	if req.Page > 1 {
		siteInfo.Canonical = fmt.Sprintf("%s/tags?page=%d", siteInfo.General.SiteUrl, req.Page)
	}
	siteInfo.Title = fmt.Sprintf("%s - %s", translator.Tr(handler.GetLang(ctx), constant.TagsListTitleTrKey), siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "tags.html", siteInfo, gin.H{
		"page": page,
		"data": data,
	})
}

// TagInfo taginfo
func (tc *TemplateController) TagInfo(ctx *gin.Context) {
	tag := ctx.Param("tag")
	req := &schema.GetTamplateTagInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		tc.Page404(ctx)
		return
	}
	nowPage := req.Page
	req.Name = tag
	tagInfo, questionList, questionCount, err := tc.templateRenderController.TagInfo(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}
	page := templaterender.Paginator(nowPage, req.PageSize, questionCount)

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/tags/%s", siteInfo.General.SiteUrl, tag)
	if req.Page > 1 {
		siteInfo.Canonical = fmt.Sprintf("%s/tags/%s?page=%d", siteInfo.General.SiteUrl, tag, req.Page)
	}
	siteInfo.Description = htmltext.FetchExcerpt(tagInfo.ParsedText, "...", 240)
	if len(tagInfo.ParsedText) == 0 {
		siteInfo.Description = translator.Tr(handler.GetLang(ctx), constant.TagHasNoDescription)
	}
	siteInfo.Keywords = tagInfo.DisplayName

	UrlUseTitle := siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitle ||
		siteInfo.SiteSeo.Permalink == constant.PermalinkQuestionIDAndTitleByShortID

	siteInfo.Title = fmt.Sprintf("'%s' %s - %s", tagInfo.DisplayName, translator.Tr(handler.GetLang(ctx), constant.QuestionsTitleTrKey), siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "tag-detail.html", siteInfo, gin.H{
		"tag":           tagInfo,
		"questionList":  questionList,
		"questionCount": questionCount,
		"useTitle":      UrlUseTitle,
		"page":          page,
	})
}

// UserInfo user info
func (tc *TemplateController) UserInfo(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		tc.Page404(ctx)
		return
	}

	exist := checker.IsUsersIgnorePath(username)
	if exist {
		file, err := ui.Build.ReadFile("build/index.html")
		if err != nil {
			log.Error(err)
			tc.Page404(ctx)
			return
		}
		ctx.Header("content-type", "text/html;charset=utf-8")
		ctx.String(http.StatusOK, string(file))
		return
	}
	req := &schema.GetOtherUserInfoByUsernameReq{}
	req.Username = username
	userinfo, err := tc.templateRenderController.UserInfo(ctx, req)
	if err != nil {
		tc.Page404(ctx)
		return
	}

	questionList, answerList, err := tc.questionService.SearchUserTopList(ctx, req.Username, "")
	if err != nil {
		tc.Page404(ctx)
		return
	}

	siteInfo := tc.SiteInfo(ctx)
	siteInfo.Canonical = fmt.Sprintf("%s/users/%s", siteInfo.General.SiteUrl, username)
	siteInfo.Title = fmt.Sprintf("%s - %s", username, siteInfo.General.Name)
	tc.html(ctx, http.StatusOK, "homepage.html", siteInfo, gin.H{
		"userinfo":     userinfo,
		"bio":          template.HTML(userinfo.BioHTML),
		"topQuestions": questionList,
		"topAnswers":   answerList,
	})
}

func (tc *TemplateController) Page404(ctx *gin.Context) {
	tc.html(ctx, http.StatusNotFound, "404.html", tc.SiteInfo(ctx), gin.H{})
}

func (tc *TemplateController) html(ctx *gin.Context, code int, tpl string, siteInfo *schema.TemplateSiteInfoResp, data gin.H) {
	prefix := ""
	cssPath := ""
	scriptPath := make([]string, len(tc.scriptPath))

	_ = plugin.CallCDN(func(fn plugin.CDN) error {
		prefix = fn.GetStaticPrefix()
		return nil
	})

	if prefix != "" {
		if prefix[len(prefix)-1:] == "/" {
			prefix = strings.TrimSuffix(prefix, "/")
		}
		cssPath = prefix + tc.cssPath
		for i, path := range tc.scriptPath {
			scriptPath[i] = prefix + path
		}
	} else {
		cssPath = tc.cssPath
		scriptPath = tc.scriptPath
	}

	data["siteinfo"] = siteInfo
	data["baseURL"] = ""
	if parsedUrl, err := url.Parse(siteInfo.General.SiteUrl); err == nil {
		data["baseURL"] = parsedUrl.Path
	}
	data["scriptPath"] = scriptPath
	data["cssPath"] = cssPath
	data["keywords"] = siteInfo.Keywords
	if siteInfo.Description == "" {
		siteInfo.Description = siteInfo.General.Description
	}
	data["title"] = siteInfo.Title
	if siteInfo.Title == "" {
		data["title"] = siteInfo.General.Name
	}
	data["description"] = siteInfo.Description
	data["language"] = handler.GetLang(ctx)
	data["timezone"] = siteInfo.Interface.TimeZone
	language := strings.ReplaceAll(siteInfo.Interface.Language, "_", "-")
	data["lang"] = language
	data["HeadCode"] = siteInfo.CustomCssHtml.CustomHead
	data["HeaderCode"] = siteInfo.CustomCssHtml.CustomHeader
	data["FooterCode"] = siteInfo.CustomCssHtml.CustomFooter
	data["Version"] = constant.Version
	data["Revision"] = constant.Revision
	_, ok := data["path"]
	if !ok {
		data["path"] = ""
	}
	ctx.Header("X-Frame-Options", "DENY")
	ctx.HTML(code, tpl, data)
}

func (tc *TemplateController) OpenSearch(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	tc.templateRenderController.OpenSearch(ctx)
}

func (tc *TemplateController) Sitemap(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	tc.templateRenderController.Sitemap(ctx)
}

func (tc *TemplateController) SitemapPage(ctx *gin.Context) {
	if tc.checkPrivateMode(ctx) {
		tc.Page404(ctx)
		return
	}
	page := 0
	pageParam := ctx.Param("page")
	pageRegexp := regexp.MustCompile(`question-(.*).xml`)
	pageStr := pageRegexp.FindStringSubmatch(pageParam)
	if len(pageStr) != 2 {
		tc.Page404(ctx)
		return
	}
	page = converter.StringToInt(pageStr[1])
	if page == 0 {
		tc.Page404(ctx)
		return
	}
	err := tc.templateRenderController.SitemapPage(ctx, page)
	if err != nil {
		tc.Page404(ctx)
		return
	}
}

func (tc *TemplateController) checkPrivateMode(ctx *gin.Context) bool {
	resp, err := tc.siteInfoService.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
		return false
	}
	if resp.LoginRequired {
		return true
	}
	return false
}

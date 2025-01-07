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

package router

import (
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/controller"
	templaterender "github.com/apache/answer/internal/controller/template_render"
	"github.com/apache/answer/internal/controller_admin"
	"github.com/gin-gonic/gin"
)

type TemplateRouter struct {
	templateController       *controller.TemplateController
	templateRenderController *templaterender.TemplateRenderController
	siteInfoController       *controller_admin.SiteInfoController
	authUserMiddleware       *middleware.AuthUserMiddleware
}

func NewTemplateRouter(
	templateController *controller.TemplateController,
	templateRenderController *templaterender.TemplateRenderController,
	siteInfoController *controller_admin.SiteInfoController,
	authUserMiddleware *middleware.AuthUserMiddleware,

) *TemplateRouter {
	return &TemplateRouter{
		templateController:       templateController,
		templateRenderController: templateRenderController,
		siteInfoController:       siteInfoController,
		authUserMiddleware:       authUserMiddleware,
	}
}

// RegisterTemplateRouter template router
func (a *TemplateRouter) RegisterTemplateRouter(r *gin.RouterGroup, baseURLPath string) {
	seoNoAuth := r.Group(baseURLPath)
	seoNoAuth.GET("/sitemap.xml", a.templateController.Sitemap)
	seoNoAuth.GET("/sitemap/:page", a.templateController.SitemapPage)

	seoNoAuth.GET("/robots.txt", a.siteInfoController.GetRobots)
	seoNoAuth.GET("/custom.css", a.siteInfoController.GetCss)

	seoNoAuth.GET("/404", a.templateController.Page404)

	seoNoAuth.GET("/opensearch.xml", a.templateController.OpenSearch)

	seo := r.Group(baseURLPath)
	seo.Use(a.authUserMiddleware.CheckPrivateMode())
	seo.GET("/", a.templateController.Index)
	seo.GET("/questions", a.templateController.QuestionList)
	seo.GET("/questions/:id", a.templateController.QuestionInfo)
	seo.GET("/questions/:id/:title", a.templateController.QuestionInfo)
	seo.GET("/questions/:id/:title/:answerid", a.templateController.QuestionInfo)
	seo.GET("/tags", a.templateController.TagList)
	seo.GET("/tags/:tag", a.templateController.TagInfo)
	seo.GET("/users/:username", a.templateController.UserInfo)
}

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
	"net/http"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

type SiteInfoController struct {
	siteInfoService siteinfo_common.SiteInfoCommonService
}

// NewSiteInfoController new site info controller.
func NewSiteInfoController(siteInfoService siteinfo_common.SiteInfoCommonService) *SiteInfoController {
	return &SiteInfoController{
		siteInfoService: siteInfoService,
	}
}

// GetSiteInfo get site info
// @Summary get site info
// @Description get site info
// @Tags site
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteInfoResp}
// @Router /answer/api/v1/siteinfo [get]
func (sc *SiteInfoController) GetSiteInfo(ctx *gin.Context) {
	var err error
	resp := &schema.SiteInfoResp{Version: constant.Version, Revision: constant.Revision}
	resp.General, err = sc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Interface, err = sc.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Branding, err = sc.siteInfoService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Login, err = sc.siteInfoService.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.Theme, err = sc.siteInfoService.GetSiteTheme(ctx)
	if err != nil {
		log.Error(err)
	}

	resp.CustomCssHtml, err = sc.siteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.SiteSeo, err = sc.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.SiteUsers, err = sc.siteInfoService.GetSiteUsers(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.Write, err = sc.siteInfoService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
	}

	handler.HandleResponse(ctx, nil, resp)
}

// GetSiteLegalInfo get site legal info
// @Summary get site legal info
// @Description get site legal info
// @Tags site
// @Param info_type query string true "legal information type" Enums(tos, privacy)
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetSiteLegalInfoResp}
// @Router /answer/api/v1/siteinfo/legal [get]
func (sc *SiteInfoController) GetSiteLegalInfo(ctx *gin.Context) {
	req := &schema.GetSiteLegalInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	siteLegal, err := sc.siteInfoService.GetSiteLegal(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	resp := &schema.GetSiteLegalInfoResp{}
	if req.IsTOS() {
		resp.TermsOfServiceOriginalText = siteLegal.TermsOfServiceOriginalText
		resp.TermsOfServiceParsedText = siteLegal.TermsOfServiceParsedText
	} else if req.IsPrivacy() {
		resp.PrivacyPolicyOriginalText = siteLegal.PrivacyPolicyOriginalText
		resp.PrivacyPolicyParsedText = siteLegal.PrivacyPolicyParsedText
	}
	handler.HandleResponse(ctx, nil, resp)
}

// GetManifestJson get manifest.json
func (sc *SiteInfoController) GetManifestJson(ctx *gin.Context) {
	favicon := "favicon.ico"
	resp := &schema.GetManifestJsonResp{
		ManifestVersion: 3,
		Version:         constant.Version,
		Revision:        constant.Revision,
		ShortName:       "Answer",
		Name:            "answer.apache.org",
		Icons: map[string]string{
			"16":  favicon,
			"32":  favicon,
			"48":  favicon,
			"128": favicon,
		},
		StartUrl:        ".",
		Display:         "standalone",
		ThemeColor:      "#000000",
		BackgroundColor: "#ffffff",
	}
	branding, err := sc.siteInfoService.GetSiteBranding(ctx)
	if err != nil {
		log.Error(err)
	} else if len(branding.Favicon) > 0 {
		resp.Icons["16"] = branding.Favicon
		resp.Icons["32"] = branding.Favicon
		resp.Icons["48"] = branding.Favicon
		resp.Icons["128"] = branding.Favicon
	}
	siteGeneral, err := sc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Error(err)
	} else {
		resp.Name = siteGeneral.Name
		resp.ShortName = siteGeneral.Name
	}
	ctx.JSON(http.StatusOK, resp)
}

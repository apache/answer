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

package controller_admin

import (
	"html"
	"net/http"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/siteinfo"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

// SiteInfoController site info controller
type SiteInfoController struct {
	siteInfoService *siteinfo.SiteInfoService
}

// NewSiteInfoController new site info controller
func NewSiteInfoController(siteInfoService *siteinfo.SiteInfoService) *SiteInfoController {
	return &SiteInfoController{
		siteInfoService: siteInfoService,
	}
}

// GetGeneral get site general information
// @Summary get site general information
// @Description get site general information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteGeneralResp}
// @Router /answer/admin/api/siteinfo/general [get]
func (sc *SiteInfoController) GetGeneral(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteGeneral(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetInterface get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteInterfaceResp}
// @Router /answer/admin/api/siteinfo/interface [get]
func (sc *SiteInfoController) GetInterface(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteInterface(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteBranding get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteBrandingResp}
// @Router /answer/admin/api/siteinfo/branding [get]
func (sc *SiteInfoController) GetSiteBranding(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteBranding(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteWrite get site interface
// @Summary get site interface
// @Description get site interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteWriteResp}
// @Router /answer/admin/api/siteinfo/write [get]
func (sc *SiteInfoController) GetSiteWrite(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteWrite(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteLegal Set the legal information for the site
// @Summary Set the legal information for the site
// @Description Set the legal information for the site
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteLegalResp}
// @Router /answer/admin/api/siteinfo/legal [get]
func (sc *SiteInfoController) GetSiteLegal(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteLegal(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSeo get site seo information
// @Summary get site seo information
// @Description get site seo information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteSeoResp}
// @Router /answer/admin/api/siteinfo/seo [get]
func (sc *SiteInfoController) GetSeo(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSeo(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteLogin get site info login config
// @Summary get site info login config
// @Description get site info login config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteLoginResp}
// @Router /answer/admin/api/siteinfo/login [get]
func (sc *SiteInfoController) GetSiteLogin(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteLogin(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteCustomCssHTML get site info custom html css config
// @Summary get site info custom html css config
// @Description get site info custom html css config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteCustomCssHTMLResp}
// @Router /answer/admin/api/siteinfo/custom-css-html [get]
func (sc *SiteInfoController) GetSiteCustomCssHTML(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteCustomCssHTML(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteTheme get site info theme config
// @Summary get site info theme config
// @Description get site info theme config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteThemeResp}
// @Router /answer/admin/api/siteinfo/theme [get]
func (sc *SiteInfoController) GetSiteTheme(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteTheme(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetSiteUsers get site user config
// @Summary get site user config
// @Description get site user config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SiteUsersResp}
// @Router /answer/admin/api/siteinfo/users [get]
func (sc *SiteInfoController) GetSiteUsers(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteUsers(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetRobots get site robots information
// @Summary get site robots information
// @Description get site robots information
// @Tags site
// @Produce json
// @Success 200 {string} txt ""
// @Router /robots.txt [get]
func (sc *SiteInfoController) GetRobots(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSeo(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "")
		return
	}
	ctx.String(http.StatusOK, resp.Robots)
}

// GetCss get site custom CSS
// @Summary get site custom CSS
// @Description get site custom CSS
// @Tags site
// @Produce text/css
// @Success 200 {string} css ""
// @Router /custom.css [get]
func (sc *SiteInfoController) GetCss(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSiteCustomCssHTML(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "")
		return
	}
	ctx.Header("content-type", "text/css;charset=utf-8")
	ctx.String(http.StatusOK, resp.CustomCss)
}

// UpdateSeo update site seo information
// @Summary update site seo information
// @Description update site seo information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteSeoReq true "seo"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/seo [put]
func (sc *SiteInfoController) UpdateSeo(ctx *gin.Context) {
	req := schema.SiteSeoReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := sc.siteInfoService.SaveSeo(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateGeneral update site general information
// @Summary update site general information
// @Description update site general information
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteGeneralReq true "general"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/general [put]
func (sc *SiteInfoController) UpdateGeneral(ctx *gin.Context) {
	req := schema.SiteGeneralReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := sc.siteInfoService.SaveSiteGeneral(ctx, req)
	req.Name = html.UnescapeString(req.Name)
	handler.HandleResponse(ctx, err, req)
}

// UpdateInterface update site interface
// @Summary update site info interface
// @Description update site info interface
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteInterfaceReq true "general"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/interface [put]
func (sc *SiteInfoController) UpdateInterface(ctx *gin.Context) {
	req := schema.SiteInterfaceReq{}
	if handler.BindAndCheck(ctx, &req) {
		return
	}
	err := sc.siteInfoService.SaveSiteInterface(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateBranding update site branding
// @Summary update site info branding
// @Description update site info branding
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteBrandingReq true "branding info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/branding [put]
func (sc *SiteInfoController) UpdateBranding(ctx *gin.Context) {
	req := &schema.SiteBrandingReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	currentBranding, getBrandingErr := sc.siteInfoService.GetSiteBranding(ctx)
	if getBrandingErr == nil {
		cleanUpErr := sc.siteInfoService.CleanUpRemovedBrandingFiles(ctx, req, currentBranding)
		if cleanUpErr != nil {
			log.Errorf("failed to clean up removed branding file(s): %v", cleanUpErr)
		}
	} else {
		log.Errorf("failed to get current site branding: %v", getBrandingErr)
	}
	saveErr := sc.siteInfoService.SaveSiteBranding(ctx, req)
	handler.HandleResponse(ctx, saveErr, nil)
}

// UpdateSiteWrite update site write info
// @Summary update site write info
// @Description update site write info
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteWriteReq true "write info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/write [put]
func (sc *SiteInfoController) UpdateSiteWrite(ctx *gin.Context) {
	req := &schema.SiteWriteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := sc.siteInfoService.SaveSiteWrite(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateSiteLegal update site legal info
// @Summary update site legal info
// @Description update site legal info
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteLegalReq true "write info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/legal [put]
func (sc *SiteInfoController) UpdateSiteLegal(ctx *gin.Context) {
	req := &schema.SiteLegalReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.SaveSiteLegal(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteLogin update site login
// @Summary update site login
// @Description update site login
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteLoginReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/login [put]
func (sc *SiteInfoController) UpdateSiteLogin(ctx *gin.Context) {
	req := &schema.SiteLoginReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.SaveSiteLogin(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteCustomCssHTML update site custom css html config
// @Summary update site custom css html config
// @Description update site custom css html config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteCustomCssHTMLReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/custom-css-html [put]
func (sc *SiteInfoController) UpdateSiteCustomCssHTML(ctx *gin.Context) {
	req := &schema.SiteCustomCssHTMLReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.SaveSiteCustomCssHTML(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// SaveSiteTheme update site custom css html config
// @Summary update site custom css html config
// @Description update site custom css html config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteThemeReq true "login info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/theme [put]
func (sc *SiteInfoController) SaveSiteTheme(ctx *gin.Context) {
	req := &schema.SiteThemeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.SaveSiteTheme(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateSiteUsers update site config about users
// @Summary update site info config about users
// @Description update site info config about users
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SiteUsersReq true "users info"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/siteinfo/users [put]
func (sc *SiteInfoController) UpdateSiteUsers(ctx *gin.Context) {
	req := &schema.SiteUsersReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.SaveSiteUsers(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetSMTPConfig get smtp config
// @Summary GetSMTPConfig get smtp config
// @Description GetSMTPConfig get smtp config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetSMTPConfigResp}
// @Router /answer/admin/api/setting/smtp [get]
func (sc *SiteInfoController) GetSMTPConfig(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetSMTPConfig(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateSMTPConfig update smtp config
// @Summary update smtp config
// @Description update smtp config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.UpdateSMTPConfigReq true "smtp config"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/setting/smtp [put]
func (sc *SiteInfoController) UpdateSMTPConfig(ctx *gin.Context) {
	req := &schema.UpdateSMTPConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.UpdateSMTPConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetPrivilegesConfig get privileges config
// @Summary GetPrivilegesConfig get privileges config
// @Description GetPrivilegesConfig get privileges config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.GetPrivilegesConfigResp}
// @Router /answer/admin/api/setting/privileges [get]
func (sc *SiteInfoController) GetPrivilegesConfig(ctx *gin.Context) {
	resp, err := sc.siteInfoService.GetPrivilegesConfig(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UpdatePrivilegesConfig update privileges config
// @Summary update privileges config
// @Description update privileges config
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.UpdatePrivilegesConfigReq true "config"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/setting/privileges [put]
func (sc *SiteInfoController) UpdatePrivilegesConfig(ctx *gin.Context) {
	req := &schema.UpdatePrivilegesConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := sc.siteInfoService.UpdatePrivilegesConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

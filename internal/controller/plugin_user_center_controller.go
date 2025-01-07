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
	"fmt"
	"net/http"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/internal/service/user_external_login"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

const (
	UserCenterLoginRouter          = "/user-center/login/redirect"
	UserCenterSignUpRedirectRouter = "/user-center/sign-up/redirect"
)

// UserCenterController comment controller
type UserCenterController struct {
	userCenterLoginService *user_external_login.UserCenterLoginService
	siteInfoService        siteinfo_common.SiteInfoCommonService
}

// NewUserCenterController new controller
func NewUserCenterController(
	userCenterLoginService *user_external_login.UserCenterLoginService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
) *UserCenterController {
	return &UserCenterController{
		userCenterLoginService: userCenterLoginService,
		siteInfoService:        siteInfoService,
	}
}

// UserCenterAgent get user center agent info
func (uc *UserCenterController) UserCenterAgent(ctx *gin.Context) {
	resp := &schema.UserCenterAgentResp{}
	resp.Enabled = plugin.UserCenterEnabled()
	if !resp.Enabled {
		handler.HandleResponse(ctx, nil, resp)
		return
	}
	siteGeneral, err := uc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site info failed: %v", err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}

	resp.AgentInfo = &schema.AgentInfo{}
	resp.AgentInfo.LoginRedirectURL = fmt.Sprintf("%s%s%s", siteGeneral.SiteUrl,
		commonRouterPrefix, UserCenterLoginRouter)
	resp.AgentInfo.SignUpRedirectURL = fmt.Sprintf("%s%s%s", siteGeneral.SiteUrl,
		commonRouterPrefix, UserCenterSignUpRedirectRouter)

	_ = plugin.CallUserCenter(func(uc plugin.UserCenter) error {
		info := uc.Description()
		resp.AgentInfo.Name = info.Name
		resp.AgentInfo.DisplayName = info.DisplayName.Translate(ctx)
		resp.AgentInfo.Icon = info.Icon
		resp.AgentInfo.Url = info.Url
		resp.AgentInfo.ControlCenterItems = make([]*schema.ControlCenter, 0)
		resp.AgentInfo.EnabledOriginalUserSystem = info.EnabledOriginalUserSystem
		items := uc.ControlCenterItems()
		for _, item := range items {
			resp.AgentInfo.ControlCenterItems = append(resp.AgentInfo.ControlCenterItems, &schema.ControlCenter{
				Name:  item.Name,
				Label: item.Label,
				Url:   item.Url,
			})
		}
		return nil
	})

	handler.HandleResponse(ctx, nil, resp)
}

// UserCenterPersonalBranding get user center personal user info
func (uc *UserCenterController) UserCenterPersonalBranding(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := uc.userCenterLoginService.UserCenterPersonalBranding(ctx, req.Username)
	handler.HandleResponse(ctx, err, resp)
}

func (uc *UserCenterController) UserCenterLoginRedirect(ctx *gin.Context) {
	var redirectURL string
	_ = plugin.CallUserCenter(func(userCenter plugin.UserCenter) error {
		info := userCenter.Description()
		redirectURL = info.LoginRedirectURL
		return nil
	})
	ctx.Redirect(http.StatusFound, redirectURL)
}

func (uc *UserCenterController) UserCenterSignUpRedirect(ctx *gin.Context) {
	var redirectURL string
	_ = plugin.CallUserCenter(func(userCenter plugin.UserCenter) error {
		info := userCenter.Description()
		redirectURL = info.LoginRedirectURL
		return nil
	})
	ctx.Redirect(http.StatusFound, redirectURL)
}

func (uc *UserCenterController) UserCenterLoginCallback(ctx *gin.Context) {
	siteGeneral, err := uc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site info failed: %v", err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}

	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		ctx.Redirect(http.StatusFound, "/404")
		return
	}
	userInfo, err := userCenter.LoginCallback(ctx)
	if err != nil {
		log.Error(err)
		if !ctx.IsAborted() {
			ctx.Redirect(http.StatusFound, "/50x")
		}
		return
	}

	resp, err := uc.userCenterLoginService.ExternalLogin(ctx, userCenter, userInfo)
	if err != nil {
		log.Errorf("external login failed: %v", err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}
	if len(resp.ErrMsg) > 0 {
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/50x?title=%s&msg=%s", resp.ErrTitle, resp.ErrMsg))
		return
	}
	userCenter.AfterLogin(userInfo.ExternalID, resp.AccessToken)
	ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/users/auth-landing?access_token=%s",
		siteGeneral.SiteUrl, resp.AccessToken))
}

func (uc *UserCenterController) UserCenterSignUpCallback(ctx *gin.Context) {
	siteGeneral, err := uc.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site info failed: %v", err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}

	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		ctx.Redirect(http.StatusFound, "/404")
		return
	}
	userInfo, err := userCenter.SignUpCallback(ctx)
	if err != nil {
		log.Error(err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}

	resp, err := uc.userCenterLoginService.ExternalLogin(ctx, userCenter, userInfo)
	if err != nil {
		log.Errorf("external login failed: %v", err)
		ctx.Redirect(http.StatusFound, "/50x")
		return
	}
	if len(resp.ErrMsg) > 0 {
		ctx.Redirect(http.StatusFound, fmt.Sprintf("/50x?title=%s&msg=%s", resp.ErrTitle, resp.ErrMsg))
		return
	}
	userCenter.AfterLogin(userInfo.ExternalID, resp.AccessToken)
	ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/users/auth-landing?access_token=%s",
		siteGeneral.SiteUrl, resp.AccessToken))
}

// UserCenterUserSettings user center user settings
func (uc *UserCenterController) UserCenterUserSettings(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := uc.userCenterLoginService.UserCenterUserSettings(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// UserCenterAdminFunctionAgent user center admin function agent
func (uc *UserCenterController) UserCenterAdminFunctionAgent(ctx *gin.Context) {
	resp, err := uc.userCenterLoginService.UserCenterAdminFunctionAgent(ctx)
	handler.HandleResponse(ctx, err, resp)
}

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
	"github.com/apache/answer/internal/controller"
	"github.com/gin-gonic/gin"
)

type PluginAPIRouter struct {
	connectorController  *controller.ConnectorController
	userCenterController *controller.UserCenterController
	captchaController    *controller.CaptchaController
	embedController      *controller.EmbedController
	renderController     *controller.RenderController
}

func NewPluginAPIRouter(
	connectorController *controller.ConnectorController,
	userCenterController *controller.UserCenterController,
	captchaController *controller.CaptchaController,
	embedController *controller.EmbedController,
	renderController *controller.RenderController,
) *PluginAPIRouter {
	return &PluginAPIRouter{
		connectorController:  connectorController,
		userCenterController: userCenterController,
		captchaController:    captchaController,
		embedController:      embedController,
		renderController:     renderController,
	}
}

func (pr *PluginAPIRouter) RegisterUnAuthConnectorRouter(r *gin.RouterGroup) {
	// connector plugin
	connectorController := pr.connectorController
	r.GET(controller.ConnectorLoginRouterPrefix+":name", connectorController.ConnectorLoginDispatcher)
	r.GET(controller.ConnectorRedirectRouterPrefix+":name", connectorController.ConnectorRedirectDispatcher)
	r.GET("/connector/info", connectorController.ConnectorsInfo)
	r.POST("/connector/binding/email", connectorController.ExternalLoginBindingUserSendEmail)

	// user center plugin
	r.GET("/user-center/agent", pr.userCenterController.UserCenterAgent)
	r.GET("/user-center/personal/branding", pr.userCenterController.UserCenterPersonalBranding)
	r.GET(controller.UserCenterLoginRouter, pr.userCenterController.UserCenterLoginRedirect)
	r.GET(controller.UserCenterSignUpRedirectRouter, pr.userCenterController.UserCenterSignUpRedirect)
	r.GET("/user-center/login/callback", pr.userCenterController.UserCenterLoginCallback)
	r.GET("/user-center/sign-up/callback", pr.userCenterController.UserCenterSignUpCallback)

	// captcha plugin
	r.GET("/captcha/config", pr.captchaController.GetCaptchaConfig)
	r.GET("/embed/config", pr.embedController.GetEmbedConfig)
	r.GET("/render/config", pr.renderController.GetRenderConfig)
}

func (pr *PluginAPIRouter) RegisterAuthUserConnectorRouter(r *gin.RouterGroup) {
	connectorController := pr.connectorController
	r.GET("/connector/user/info", connectorController.ConnectorsUserInfo)
	r.DELETE("/connector/user/unbinding", connectorController.ExternalLoginUnbinding)

	r.GET("/user-center/user/settings", pr.userCenterController.UserCenterUserSettings)
}

func (pr *PluginAPIRouter) RegisterAuthAdminConnectorRouter(r *gin.RouterGroup) {
	r.GET("/user-center/agent", pr.userCenterController.UserCenterAdminFunctionAgent)
}

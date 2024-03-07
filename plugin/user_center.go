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

package plugin

type UserCenter interface {
	Base
	// Description returns the description of the user center, including the name, icon, url, etc.
	Description() UserCenterDesc
	// ControlCenterItems returns the items that will be displayed in the control center
	ControlCenterItems() []ControlCenter
	// LoginCallback is called when the user center login callback is called
	LoginCallback(ctx *GinContext) (userInfo *UserCenterBasicUserInfo, err error)
	// SignUpCallback is called when the user center sign up callback is called
	SignUpCallback(ctx *GinContext) (userInfo *UserCenterBasicUserInfo, err error)
	// UserInfo returns the user information
	UserInfo(externalID string) (userInfo *UserCenterBasicUserInfo, err error)
	// UserStatus returns the latest user status
	UserStatus(externalID string) (userStatus UserStatus)
	// UserList returns the user list information
	UserList(externalIDs []string) (userInfo []*UserCenterBasicUserInfo, err error)
	// UserSettings returns the user settings
	UserSettings(externalID string) (userSettings *SettingInfo, err error)
	// PersonalBranding returns the personal branding information
	PersonalBranding(externalID string) (branding []*PersonalBranding)
	// AfterLogin is called after the user logs in
	AfterLogin(externalID, accessToken string)
}

type UserCenterDesc struct {
	Name                      string     `json:"name"`
	DisplayName               Translator `json:"display_name"`
	Icon                      string     `json:"icon"`
	Url                       string     `json:"url"`
	LoginRedirectURL          string     `json:"login_redirect_url"`
	SignUpRedirectURL         string     `json:"sign_up_redirect_url"`
	RankAgentEnabled          bool       `json:"rank_agent_enabled"`
	UserStatusAgentEnabled    bool       `json:"user_status_agent_enabled"`
	UserRoleAgentEnabled      bool       `json:"user_role_agent_enabled"`
	MustAuthEmailEnabled      bool       `json:"must_auth_email_enabled"`
	EnabledOriginalUserSystem bool       `json:"enabled_original_user_system"`
}

type UserStatus int

const (
	UserStatusAvailable UserStatus = 1
	UserStatusSuspended UserStatus = 9
	UserStatusDeleted   UserStatus = 10
)

type UserCenterBasicUserInfo struct {
	ExternalID  string     `json:"external_id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Rank        int        `json:"rank"`
	Avatar      string     `json:"avatar"`
	Mobile      string     `json:"mobile"`
	Bio         string     `json:"bio"`
	Status      UserStatus `json:"status"`
	Language    string     `json:"language"`
}

type ControlCenter struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Url   string `json:"url"`
}

type SettingInfo struct {
	ProfileSettingRedirectURL string `json:"profile_setting_redirect_url"`
	AccountSettingRedirectURL string `json:"account_setting_redirect_url"`
}

type PersonalBranding struct {
	Icon  string `json:"icon"`
	Name  string `json:"name"`
	Label string `json:"label"`
	Url   string `json:"url"`
}

var (
	// CallUserCenter is a function that calls all registered parsers
	CallUserCenter,
	registerUserCenter = MakePlugin[UserCenter](false)
)

func UserCenterEnabled() (enabled bool) {
	_ = CallUserCenter(func(fn UserCenter) error {
		enabled = true
		return nil
	})
	return
}

func RankAgentEnabled() (enabled bool) {
	_ = CallUserCenter(func(fn UserCenter) error {
		enabled = fn.Description().RankAgentEnabled
		return nil
	})
	return
}

func GetUserCenter() (uc UserCenter, ok bool) {
	_ = CallUserCenter(func(fn UserCenter) error {
		uc = fn
		ok = true
		return nil
	})
	return
}

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

package schema

// PersonalAccessTokenCreateReq creates a scoped user credential.
type PersonalAccessTokenCreateReq struct {
	Name      string   `validate:"required,notblank,lte=100" json:"name"`
	Scopes    []string `validate:"required,min=1,dive,required" json:"scopes"`
	ExpiresAt int64    `validate:"required" json:"expires_at"`
}

type PersonalAccessTokenInfo struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	TokenSuffix string   `json:"token_suffix"`
	Scopes      []string `json:"scopes"`
	CreatedAt   int64    `json:"created_at"`
	ExpiresAt   int64    `json:"expires_at"`
	RevokedAt   int64    `json:"revoked_at,omitempty"`
	Status      string   `json:"status"`
}

type PersonalAccessTokenCreateResp struct {
	Token string `json:"token"`
	PersonalAccessTokenInfo
}

type PersonalAccessTokenRevokeReq struct {
	ID int64 `validate:"required" form:"id"`
}

type PersonalAccessTokenCurrentResp struct {
	User  PersonalAccessTokenUserInfo `json:"user"`
	Token PersonalAccessTokenInfo     `json:"token"`
}

type PersonalAccessTokenUserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type UserReauthenticateReq struct {
	Password    string `validate:"required" json:"password"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	UserID      string `json:"-"`
	AccessToken string `json:"-"`
}

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

import (
	"encoding/json"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/segmentfault/pacman/errors"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/converter"
	"github.com/jinzhu/copier"
)

// UserVerifyEmailReq user verify email request
type UserVerifyEmailReq struct {
	// code
	Code string `validate:"required,gt=0,lte=500" form:"code"`
	// content
	Content string `json:"-"`
}

// UserLoginResp get user response
type UserLoginResp struct {
	// user id
	ID string `json:"id"`
	// create time
	CreatedAt int64 `json:"created_at"`
	// last login date
	LastLoginDate int64 `json:"last_login_date"`
	// username
	Username string `json:"username"`
	// email
	EMail string `json:"e_mail"`
	// mail status(1 pass 2 to be verified)
	MailStatus int `json:"mail_status"`
	// notice status(1 on 2off)
	NoticeStatus int `json:"notice_status"`
	// follow count
	FollowCount int `json:"follow_count"`
	// answer count
	AnswerCount int `json:"answer_count"`
	// question count
	QuestionCount int `json:"question_count"`
	// rank
	Rank int `json:"rank"`
	// authority group
	AuthorityGroup int `json:"authority_group"`
	// display name
	DisplayName string `json:"display_name"`
	// avatar
	Avatar string `json:"avatar"`
	// mobile
	Mobile string `json:"mobile"`
	// bio markdown
	Bio string `json:"bio"`
	// bio html
	BioHTML string `json:"bio_html"`
	// website
	Website string `json:"website"`
	// location
	Location string `json:"location"`
	// language
	Language string `json:"language"`
	// Color scheme
	ColorScheme string `json:"color_scheme"`
	// access token
	AccessToken string `json:"access_token"`
	// role id
	RoleID int `json:"role_id"`
	// user status
	Status string `json:"status"`
	// user have password
	HavePassword bool `json:"have_password"`
	// visit token
	VisitToken string `json:"visit_token"`
}

func (r *UserLoginResp) ConvertFromUserEntity(userInfo *entity.User) {
	_ = copier.Copy(r, userInfo)
	r.CreatedAt = userInfo.CreatedAt.Unix()
	r.LastLoginDate = userInfo.LastLoginDate.Unix()
	r.Status = constant.ConvertUserStatus(userInfo.Status, userInfo.MailStatus)
	r.HavePassword = len(userInfo.Pass) > 0
}

type GetCurrentLoginUserInfoResp struct {
	*UserLoginResp
	Avatar *AvatarInfo `json:"avatar"`
}

func (r *GetCurrentLoginUserInfoResp) ConvertFromUserEntity(userInfo *entity.User) {
	_ = copier.Copy(r, userInfo)
	r.CreatedAt = userInfo.CreatedAt.Unix()
	r.LastLoginDate = userInfo.LastLoginDate.Unix()
	r.Status = constant.ConvertUserStatus(userInfo.Status, userInfo.MailStatus)
	if len(r.ColorScheme) == 0 {
		r.ColorScheme = constant.ColorSchemeDefault
	}
}

// GetOtherUserInfoByUsernameResp get user response
type GetOtherUserInfoByUsernameResp struct {
	// user id
	ID string `json:"id"`
	// create time
	CreatedAt int64 `json:"created_at"`
	// last login date
	LastLoginDate int64 `json:"last_login_date"`
	// username
	Username string `json:"username"`
	// email
	// follow count
	FollowCount int `json:"follow_count"`
	// answer count
	AnswerCount int `json:"answer_count"`
	// question count
	QuestionCount int `json:"question_count"`
	// rank
	Rank int `json:"rank"`
	// display name
	DisplayName string `json:"display_name"`
	// avatar
	Avatar string `json:"avatar"`
	// mobile
	Mobile string `json:"mobile"`
	// bio markdown
	Bio string `json:"bio"`
	// bio html
	BioHTML string `json:"bio_html"`
	// website
	Website string `json:"website"`
	// location
	Location  string `json:"location"`
	Status    string `json:"status"`
	StatusMsg string `json:"status_msg,omitempty"`
}

func (r *GetOtherUserInfoByUsernameResp) ConvertFromUserEntity(userInfo *entity.User) {
	_ = copier.Copy(r, userInfo)
	r.CreatedAt = userInfo.CreatedAt.Unix()
	r.LastLoginDate = userInfo.LastLoginDate.Unix()
	r.Status = constant.ConvertUserStatus(userInfo.Status, userInfo.MailStatus)
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		statusMsgShow, ok := UserStatusShowMsg[11]
		if ok {
			r.StatusMsg = statusMsgShow
		}
	} else {
		statusMsgShow, ok := UserStatusShowMsg[userInfo.Status]
		if ok {
			r.StatusMsg = statusMsgShow
		}
	}
}

const (
	NoticeStatusOn  = 1
	NoticeStatusOff = 2
)

var UserStatusShowMsg = map[int]string{
	1:  "",
	9:  "<strong>This user was suspended forever.</strong> This user doesn't meet a community guideline.",
	10: "This user was deleted.",
	11: "This user is inactive.",
}

// UserEmailLoginReq user email login request
type UserEmailLoginReq struct {
	Email       string `validate:"required,email,gt=0,lte=500" json:"e_mail"`
	Pass        string `validate:"required,gte=8,lte=32" json:"pass"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

// UserRegisterReq user register request
type UserRegisterReq struct {
	Name        string `validate:"required,gte=2,lte=30" json:"name"`
	Email       string `validate:"required,email,gt=0,lte=500" json:"e_mail" `
	Pass        string `validate:"required,gte=8,lte=32" json:"pass"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	IP          string `json:"-" `
}

func (u *UserRegisterReq) Check() (errFields []*validator.FormErrorField, err error) {
	if err = checker.CheckPassword(u.Pass); err != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   err.Error(),
		})
		return errFields, err
	}
	return nil, nil
}

type UserModifyPasswordReq struct {
	OldPass     string `validate:"omitempty,gte=8,lte=32" json:"old_pass"`
	Pass        string `validate:"required,gte=8,lte=32" json:"pass"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	UserID      string `json:"-"`
	AccessToken string `json:"-"`
}

func (u *UserModifyPasswordReq) Check() (errFields []*validator.FormErrorField, err error) {
	if err = checker.CheckPassword(u.Pass); err != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   err.Error(),
		})
		return errFields, err
	}
	return nil, nil
}

type UpdateInfoRequest struct {
	DisplayName string     `validate:"omitempty,gt=0,lte=30" json:"display_name"`
	Username    string     `validate:"omitempty,gte=2,lte=30" json:"username"`
	Avatar      AvatarInfo `json:"avatar"`
	Bio         string     `validate:"omitempty,gt=0,lte=4096" json:"bio"`
	BioHTML     string     `json:"-"`
	Website     string     `validate:"omitempty,gt=0,lte=500" json:"website"`
	Location    string     `validate:"omitempty,gt=0,lte=100" json:"location"`
	UserID      string     `json:"-"`
	IsAdmin     bool       `json:"-"`
}

type AvatarInfo struct {
	Type     string `validate:"omitempty,gt=0,lte=100"  json:"type"`
	Gravatar string `validate:"omitempty,gt=0,lte=200"  json:"gravatar"`
	Custom   string `validate:"omitempty,gt=0,lte=200"  json:"custom"`
}

func (a *AvatarInfo) ToJsonString() string {
	data, _ := json.Marshal(a)
	return string(data)
}

func (a *AvatarInfo) GetURL() string {
	switch a.Type {
	case constant.AvatarTypeGravatar:
		return a.Gravatar
	case constant.AvatarTypeCustom:
		return a.Custom
	default:
		return ""
	}
}

func CustomAvatar(url string) *AvatarInfo {
	return &AvatarInfo{
		Type:   constant.AvatarTypeCustom,
		Custom: url,
	}
}

func (req *UpdateInfoRequest) Check() (errFields []*validator.FormErrorField, err error) {
	req.BioHTML = converter.Markdown2BasicHTML(req.Bio)
	if len(req.Website) > 0 && !checker.IsURL(req.Website) {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "website",
			ErrorMsg:   reason.InvalidURLError,
		}), errors.BadRequest(reason.InvalidURLError)
	}
	return nil, nil
}

// UpdateUserInterfaceRequest update user interface request
type UpdateUserInterfaceRequest struct {
	// language
	Language string `validate:"required,gt=1,lte=100" json:"language"`
	// Color scheme
	ColorScheme string `validate:"required,gt=1,lte=100" json:"color_scheme"`
	// user id
	UserId string `json:"-"`
}

func (req *UpdateUserInterfaceRequest) Check() (errFields []*validator.FormErrorField, err error) {
	if !translator.CheckLanguageIsValid(req.Language) {
		return nil, errors.BadRequest(reason.LangNotFound)
	}
	if req.ColorScheme != constant.ColorSchemeDefault &&
		req.ColorScheme != constant.ColorSchemeLight &&
		req.ColorScheme != constant.ColorSchemeDark &&
		req.ColorScheme != constant.ColorSchemeSystem {
		req.ColorScheme = constant.ColorSchemeDefault
	}
	return nil, nil
}

type UserRetrievePassWordRequest struct {
	Email       string `validate:"required,email,gt=0,lte=500" json:"e_mail"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type UserRePassWordRequest struct {
	Code    string `validate:"required,gt=0,lte=100" json:"code"`
	Pass    string `validate:"required,gt=0,lte=32" json:"pass"`
	Content string `json:"-"`
}

func (u *UserRePassWordRequest) Check() (errFields []*validator.FormErrorField, err error) {
	if err = checker.CheckPassword(u.Pass); err != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   err.Error(),
		})
		return errFields, err
	}
	return nil, nil
}

type ActionRecordReq struct {
	Action string `validate:"required,oneof=email password edit_userinfo question answer comment edit invitation_answer search report delete vote" form:"action"`
	IP     string `json:"-"`
	UserID string `json:"-"`
}

type ActionRecordResp struct {
	CaptchaID  string `json:"captcha_id"`
	CaptchaImg string `json:"captcha_img"`
	Verify     bool   `json:"verify"`
}

type UserBasicInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Rank        int    `json:"rank"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Website     string `json:"website"`
	Location    string `json:"location"`
	Language    string `json:"language"`
	Status      string `json:"status"`
}

type GetOtherUserInfoByUsernameReq struct {
	Username string `validate:"required,gt=0,lte=500" form:"username"`
	UserID   string `json:"-"`
	IsAdmin  bool   `json:"-"`
}

type GetOtherUserInfoResp struct {
	Info *GetOtherUserInfoByUsernameResp `json:"info"`
}

type UserChangeEmailSendCodeReq struct {
	UserVerifyEmailSendReq
	Email  string `validate:"required,email,gt=0,lte=500" json:"e_mail"`
	Pass   string `validate:"omitempty,gte=8,lte=32" json:"pass"`
	UserID string `json:"-"`
}

type UserChangeEmailVerifyReq struct {
	Code    string `validate:"required,gt=0,lte=500" json:"code"`
	Content string `json:"-"`
}

type UserVerifyEmailSendReq struct {
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

// UserRankingResp user ranking response
type UserRankingResp struct {
	UsersWithTheMostReputation []*UserRankingSimpleInfo `json:"users_with_the_most_reputation"`
	UsersWithTheMostVote       []*UserRankingSimpleInfo `json:"users_with_the_most_vote"`
	Staffs                     []*UserRankingSimpleInfo `json:"staffs"`
}

// UserRankingSimpleInfo user ranking simple info
type UserRankingSimpleInfo struct {
	// username
	Username string `json:"username"`
	// rank
	Rank int `json:"rank"`
	// vote
	VoteCount int `json:"vote_count"`
	// display name
	DisplayName string `json:"display_name"`
	// avatar
	Avatar string `json:"avatar"`
}

// UserUnsubscribeNotificationReq user unsubscribe email notification request
type UserUnsubscribeNotificationReq struct {
	Code    string `validate:"required,gt=0,lte=500" json:"code"`
	Content string `json:"-"`
}

// GetUserStaffReq get user staff request
type GetUserStaffReq struct {
	Username string `validate:"omitempty,gt=0,lte=500" form:"username"`
	PageSize int    `validate:"omitempty,min=1" form:"page_size"`
}

// GetUserStaffResp get user staff response
type GetUserStaffResp struct {
	// username
	Username string `json:"username"`
	// display name
	DisplayName string `json:"display_name"`
	// avatar
	Avatar string `json:"avatar"`
}

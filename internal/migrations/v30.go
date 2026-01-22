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

package migrations

import (
	"context"
	"encoding/json"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func updateAdminMenuSettings(ctx context.Context, x *xorm.Engine) (err error) {
	err = splitWriteMenu(ctx, x)
	if err != nil {
		return
	}

	err = splitInterfaceMenu(ctx, x)
	if err != nil {
		return
	}

	err = splitLegalMenu(ctx, x)
	if err != nil {
		return
	}
	return
}

// splitWriteMenu splits the site write settings into advanced, questions, and tags settings
func splitWriteMenu(ctx context.Context, x *xorm.Engine) error {
	var (
		siteInfo          = &entity.SiteInfo{}
		siteInfoAdvanced  = &entity.SiteInfo{}
		siteInfoQuestions = &entity.SiteInfo{}
		siteInfoTags      = &entity.SiteInfo{}
	)
	exist, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeWrite}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return err
	}
	if !exist {
		return nil
	}
	siteWrite := &schema.SiteWriteResp{}
	if err := json.Unmarshal([]byte(siteInfo.Content), siteWrite); err != nil {
		return err
	}
	// site advanced settings
	siteAdvanced := &schema.SiteAdvancedResp{
		MaxImageSize:                   siteWrite.MaxImageSize,
		MaxAttachmentSize:              siteWrite.MaxAttachmentSize,
		MaxImageMegapixel:              siteWrite.MaxImageMegapixel,
		AuthorizedImageExtensions:      siteWrite.AuthorizedImageExtensions,
		AuthorizedAttachmentExtensions: siteWrite.AuthorizedAttachmentExtensions,
	}
	// site questions settings
	siteQuestions := &schema.SiteQuestionsResp{
		MinimumTags:    siteWrite.MinimumTags,
		MinimumContent: siteWrite.MinimumContent,
		RestrictAnswer: siteWrite.RestrictAnswer,
	}
	// site tags settings
	siteTags := &schema.SiteTagsResp{
		ReservedTags:  siteWrite.ReservedTags,
		RecommendTags: siteWrite.RecommendTags,
		RequiredTag:   siteWrite.RequiredTag,
	}

	// save site settings
	// save advanced settings
	existsAdvanced, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeWrite}).Get(siteInfoAdvanced)
	if err != nil {
		return err
	}
	advancedContent, err := json.Marshal(siteAdvanced)
	if err != nil {
		return err
	}
	if !existsAdvanced {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeAdvanced,
			Content: string(advancedContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save questions settings
	existsQuestions, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeQuestions}).Get(siteInfoQuestions)
	if err != nil {
		return err
	}
	questionsContent, err := json.Marshal(siteQuestions)
	if err != nil {
		return err
	}
	if !existsQuestions {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeQuestions,
			Content: string(questionsContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save tags settings
	existsTags, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeTags}).Get(siteInfoTags)
	if err != nil {
		return err
	}
	tagsContent, err := json.Marshal(siteTags)
	if err != nil {
		return err
	}
	if !existsTags {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeTags,
			Content: string(tagsContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// splitInterfaceMenu splits the site interface settings into interface and user settings
func splitInterfaceMenu(ctx context.Context, x *xorm.Engine) error {
	var (
		siteInfo          = &entity.SiteInfo{}
		siteInfoInterface = &entity.SiteInfo{}
		siteInfoUsers     = &entity.SiteInfo{}
	)
	type SiteInterface struct {
		Language        string `validate:"required,gt=1,lte=128" form:"language" json:"language"`
		TimeZone        string `validate:"required,gt=1,lte=128" form:"time_zone" json:"time_zone"`
		DefaultAvatar   string `validate:"required,oneof=system gravatar" json:"default_avatar"`
		GravatarBaseURL string `validate:"omitempty" json:"gravatar_base_url"`
	}

	exist, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeInterface}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return err
	}
	if !exist {
		return nil
	}
	oldSiteInterface := &SiteInterface{}
	if err := json.Unmarshal([]byte(siteInfo.Content), oldSiteInterface); err != nil {
		return err
	}
	siteUser := &schema.SiteUsersSettingsResp{
		DefaultAvatar:   oldSiteInterface.DefaultAvatar,
		GravatarBaseURL: oldSiteInterface.GravatarBaseURL,
	}
	siteInterface := &schema.SiteInterfaceResp{
		Language: oldSiteInterface.Language,
		TimeZone: oldSiteInterface.TimeZone,
	}

	// save settings
	// save user settings
	existsUsers, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeUsersSettings}).Get(siteInfoUsers)
	if err != nil {
		return err
	}
	userContent, err := json.Marshal(siteUser)
	if err != nil {
		return err
	}
	if !existsUsers {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeUsersSettings,
			Content: string(userContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save interface settings
	existsInterface, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeInterfaceSettings}).Get(siteInfoInterface)
	if err != nil {
		return err
	}
	interfaceContent, err := json.Marshal(siteInterface)
	if err != nil {
		return err
	}
	if !existsInterface {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeInterfaceSettings,
			Content: string(interfaceContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// splitLegalMenu splits the site legal settings into policies and security settings
func splitLegalMenu(ctx context.Context, x *xorm.Engine) error {
	var (
		siteInfo         = &entity.SiteInfo{}
		siteInfoPolices  = &entity.SiteInfo{}
		siteInfoSecurity = &entity.SiteInfo{}
		siteInfoLogin    = &entity.SiteInfo{}
		siteInfoGeneral  = &entity.SiteInfo{}
	)

	type SiteLogin struct {
		AllowNewRegistrations   bool     `json:"allow_new_registrations"`
		AllowEmailRegistrations bool     `json:"allow_email_registrations"`
		AllowPasswordLogin      bool     `json:"allow_password_login"`
		LoginRequired           bool     `json:"login_required"`
		AllowEmailDomains       []string `json:"allow_email_domains"`
	}

	type SiteGeneral struct {
		Name             string `validate:"required,sanitizer,gt=1,lte=128" form:"name" json:"name"`
		ShortDescription string `validate:"omitempty,sanitizer,gt=3,lte=255" form:"short_description" json:"short_description"`
		Description      string `validate:"omitempty,sanitizer,gt=3,lte=2000" form:"description" json:"description"`
		SiteUrl          string `validate:"required,sanitizer,gt=1,lte=512,url" form:"site_url" json:"site_url"`
		ContactEmail     string `validate:"required,sanitizer,gt=1,lte=512,email" form:"contact_email" json:"contact_email"`
		CheckUpdate      bool   `validate:"omitempty,sanitizer" form:"check_update" json:"check_update"`
	}

	// find old site legal settings
	exist, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeLegal}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return err
	}
	if !exist {
		return nil
	}
	oldSiteLegal := &schema.SiteLegalResp{}
	if err := json.Unmarshal([]byte(siteInfo.Content), oldSiteLegal); err != nil {
		return err
	}

	// find old site login settings
	existsLogin, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeLogin}).Get(siteInfoLogin)
	if err != nil {
		return err
	}
	oldSiteLogin := &SiteLogin{}
	if err := json.Unmarshal([]byte(siteInfoLogin.Content), oldSiteLogin); err != nil {
		return err
	}

	// find old site general settings
	existGeneral, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeGeneral}).Get(siteInfoGeneral)
	if err != nil {
		return err
	}
	oldSiteGeneral := &SiteGeneral{}
	if err := json.Unmarshal([]byte(siteInfoLogin.Content), oldSiteGeneral); err != nil {
		return err
	}

	sitePolicies := &schema.SitePoliciesResp{
		TermsOfServiceOriginalText: oldSiteLegal.TermsOfServiceOriginalText,
		TermsOfServiceParsedText:   oldSiteLegal.TermsOfServiceParsedText,
		PrivacyPolicyOriginalText:  oldSiteLegal.PrivacyPolicyOriginalText,
		PrivacyPolicyParsedText:    oldSiteLegal.PrivacyPolicyParsedText,
	}
	siteLogin := &schema.SiteLoginResp{
		AllowNewRegistrations:   oldSiteLogin.AllowNewRegistrations,
		AllowEmailRegistrations: oldSiteLogin.AllowEmailRegistrations,
		AllowPasswordLogin:      oldSiteLogin.AllowPasswordLogin,
		AllowEmailDomains:       oldSiteLogin.AllowEmailDomains,
	}
	siteGeneral := &schema.SiteGeneralReq{
		Name:             oldSiteGeneral.Name,
		ShortDescription: oldSiteGeneral.ShortDescription,
		Description:      oldSiteGeneral.Description,
		SiteUrl:          oldSiteGeneral.SiteUrl,
		ContactEmail:     oldSiteGeneral.ContactEmail,
	}
	siteSecurity := &schema.SiteSecurityResp{
		LoginRequired:          oldSiteLogin.LoginRequired,
		ExternalContentDisplay: oldSiteLegal.ExternalContentDisplay,
		CheckUpdate:            oldSiteGeneral.CheckUpdate,
	}
	if !existsLogin {
		siteSecurity.LoginRequired = false
	}
	if !existGeneral {
		siteSecurity.CheckUpdate = true
	}

	// save settings
	// save policies settings
	existsPolicies, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypePolicies}).Get(siteInfoPolices)
	if err != nil {
		return err
	}
	policiesContent, err := json.Marshal(sitePolicies)
	if err != nil {
		return err
	}
	if !existsPolicies {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypePolicies,
			Content: string(policiesContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save security settings
	existsSecurity, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeSecurity}).Get(siteInfoSecurity)
	if err != nil {
		return err
	}
	securityContent, err := json.Marshal(siteSecurity)
	if err != nil {
		return err
	}
	if !existsSecurity {
		_, err = x.Context(ctx).Insert(&entity.SiteInfo{
			Type:    constant.SiteTypeSecurity,
			Content: string(securityContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save login settings
	if existsLogin {
		loginContent, err := json.Marshal(siteLogin)
		_, err = x.Context(ctx).ID(siteInfoLogin.ID).Update(&entity.SiteInfo{
			Type:    constant.SiteTypeLogin,
			Content: string(loginContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}

	// save general settings
	if existGeneral {
		generalContent, err := json.Marshal(siteGeneral)
		_, err = x.Context(ctx).ID(siteInfoGeneral.ID).Update(&entity.SiteInfo{
			Type:    constant.SiteTypeGeneral,
			Content: string(generalContent),
			Status:  1,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

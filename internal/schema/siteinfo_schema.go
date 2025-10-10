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
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/segmentfault/pacman/errors"
)

// SiteGeneralReq site general request
type SiteGeneralReq struct {
	Name             string `validate:"required,sanitizer,gt=1,lte=128" form:"name" json:"name"`
	ShortDescription string `validate:"omitempty,sanitizer,gt=3,lte=255" form:"short_description" json:"short_description"`
	Description      string `validate:"omitempty,sanitizer,gt=3,lte=2000" form:"description" json:"description"`
	SiteUrl          string `validate:"required,sanitizer,gt=1,lte=512,url" form:"site_url" json:"site_url"`
	ContactEmail     string `validate:"required,sanitizer,gt=1,lte=512,email" form:"contact_email" json:"contact_email"`
	CheckUpdate      bool   `validate:"omitempty,sanitizer" form:"check_update" json:"check_update"`
}

func (r *SiteGeneralReq) FormatSiteUrl() {
	parsedUrl, err := url.Parse(r.SiteUrl)
	if err != nil {
		return
	}
	r.SiteUrl = fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	if len(parsedUrl.Path) > 0 {
		r.SiteUrl = r.SiteUrl + parsedUrl.Path
		r.SiteUrl = strings.TrimSuffix(r.SiteUrl, "/")
	}
}

// SiteInterfaceReq site interface request
type SiteInterfaceReq struct {
	Language        string `validate:"required,gt=1,lte=128" form:"language" json:"language"`
	TimeZone        string `validate:"required,gt=1,lte=128" form:"time_zone" json:"time_zone"`
	DefaultAvatar   string `validate:"required,oneof=system gravatar" json:"default_avatar"`
	GravatarBaseURL string `validate:"omitempty" json:"gravatar_base_url"`
}

// SiteBrandingReq site branding request
type SiteBrandingReq struct {
	Logo       string `validate:"omitempty,gt=0,lte=512" form:"logo" json:"logo"`
	MobileLogo string `validate:"omitempty,gt=0,lte=512" form:"mobile_logo" json:"mobile_logo"`
	SquareIcon string `validate:"omitempty,gt=0,lte=512" form:"square_icon" json:"square_icon"`
	Favicon    string `validate:"omitempty,gt=0,lte=512" form:"favicon" json:"favicon"`
}

// SiteWriteReq site write request
type SiteWriteReq struct {
	RestrictAnswer                 bool            `validate:"omitempty" json:"restrict_answer"`
	MinimumTags                    int             `validate:"omitempty" json:"min_tags"`
	RequiredTag                    bool            `validate:"omitempty" json:"required_tag"`
	RecommendTags                  []*SiteWriteTag `validate:"omitempty,dive" json:"recommend_tags"`
	ReservedTags                   []*SiteWriteTag `validate:"omitempty,dive" json:"reserved_tags"`
	MaxImageSize                   int             `validate:"omitempty,gt=0" json:"max_image_size"`
	MaxAttachmentSize              int             `validate:"omitempty,gt=0" json:"max_attachment_size"`
	MaxImageMegapixel              int             `validate:"omitempty,gt=0" json:"max_image_megapixel"`
	AuthorizedImageExtensions      []string        `validate:"omitempty" json:"authorized_image_extensions"`
	AuthorizedAttachmentExtensions []string        `validate:"omitempty" json:"authorized_attachment_extensions"`
	UserID                         string          `json:"-"`
}

func (s *SiteWriteResp) GetMaxImageSize() int64 {
	if s.MaxImageSize <= 0 {
		return constant.DefaultMaxImageSize
	}
	return int64(s.MaxImageSize) * 1024 * 1024
}

func (s *SiteWriteResp) GetMaxAttachmentSize() int64 {
	if s.MaxAttachmentSize <= 0 {
		return constant.DefaultMaxAttachmentSize
	}
	return int64(s.MaxAttachmentSize) * 1024 * 1024
}

func (s *SiteWriteResp) GetMaxImageMegapixel() int {
	if s.MaxImageMegapixel <= 0 {
		return constant.DefaultMaxImageMegapixel
	}
	return s.MaxImageMegapixel * 1000 * 1000
}

// SiteWriteTag site write response tag
type SiteWriteTag struct {
	SlugName    string `validate:"required" json:"slug_name"`
	DisplayName string `json:"display_name"`
}

// SiteLegalReq site branding request
type SiteLegalReq struct {
	TermsOfServiceOriginalText string `json:"terms_of_service_original_text"`
	TermsOfServiceParsedText   string `json:"terms_of_service_parsed_text"`
	PrivacyPolicyOriginalText  string `json:"privacy_policy_original_text"`
	PrivacyPolicyParsedText    string `json:"privacy_policy_parsed_text"`
	ExternalContentDisplay     string `validate:"required,oneof=always_display ask_before_display" json:"external_content_display"`
}

// GetSiteLegalInfoReq site site legal request
type GetSiteLegalInfoReq struct {
	InfoType string `validate:"required,oneof=tos privacy" form:"info_type"`
}

func (r *GetSiteLegalInfoReq) IsTOS() bool {
	return r.InfoType == "tos"
}

func (r *GetSiteLegalInfoReq) IsPrivacy() bool {
	return r.InfoType == "privacy"
}

// GetSiteLegalInfoResp get site legal info response
type GetSiteLegalInfoResp struct {
	TermsOfServiceOriginalText string `json:"terms_of_service_original_text,omitempty"`
	TermsOfServiceParsedText   string `json:"terms_of_service_parsed_text,omitempty"`
	PrivacyPolicyOriginalText  string `json:"privacy_policy_original_text,omitempty"`
	PrivacyPolicyParsedText    string `json:"privacy_policy_parsed_text,omitempty"`
}

// SiteUsersReq site users config request
type SiteUsersReq struct {
	DefaultAvatar          string `validate:"required,oneof=system gravatar" json:"default_avatar"`
	GravatarBaseURL        string `json:"gravatar_base_url"`
	AllowUpdateDisplayName bool   `json:"allow_update_display_name"`
	AllowUpdateUsername    bool   `json:"allow_update_username"`
	AllowUpdateAvatar      bool   `json:"allow_update_avatar"`
	AllowUpdateBio         bool   `json:"allow_update_bio"`
	AllowUpdateWebsite     bool   `json:"allow_update_website"`
	AllowUpdateLocation    bool   `json:"allow_update_location"`
}

// SiteLoginReq site login request
type SiteLoginReq struct {
	AllowNewRegistrations   bool     `json:"allow_new_registrations"`
	AllowEmailRegistrations bool     `json:"allow_email_registrations"`
	AllowPasswordLogin      bool     `json:"allow_password_login"`
	LoginRequired           bool     `json:"login_required"`
	AllowEmailDomains       []string `json:"allow_email_domains"`
}

// SiteCustomCssHTMLReq site custom css html
type SiteCustomCssHTMLReq struct {
	CustomHead    string `validate:"omitempty,gt=0,lte=65536" json:"custom_head"`
	CustomCss     string `validate:"omitempty,gt=0,lte=65536" json:"custom_css"`
	CustomHeader  string `validate:"omitempty,gt=0,lte=65536" json:"custom_header"`
	CustomFooter  string `validate:"omitempty,gt=0,lte=65536" json:"custom_footer"`
	CustomSideBar string `validate:"omitempty,gt=0,lte=65536" json:"custom_sidebar"`
}

// SiteThemeReq site theme config
type SiteThemeReq struct {
	Theme       string                 `validate:"required,gt=0,lte=255" json:"theme"`
	ThemeConfig map[string]interface{} `validate:"omitempty" json:"theme_config"`
	ColorScheme string                 `validate:"omitempty,gt=0,lte=100" json:"color_scheme"`
}

type SiteSeoReq struct {
	Permalink int    `validate:"required,lte=4,gte=0" form:"permalink" json:"permalink"`
	Robots    string `validate:"required" form:"robots" json:"robots"`
}

func (s *SiteSeoResp) IsShortLink() bool {
	return s.Permalink == constant.PermalinkQuestionIDAndTitleByShortID ||
		s.Permalink == constant.PermalinkQuestionIDByShortID
}

// SiteGeneralResp site general response
type SiteGeneralResp SiteGeneralReq

// SiteInterfaceResp site interface response
type SiteInterfaceResp SiteInterfaceReq

// SiteBrandingResp site branding response
type SiteBrandingResp SiteBrandingReq

// SiteLoginResp site login response
type SiteLoginResp SiteLoginReq

// SiteCustomCssHTMLResp site custom css html response
type SiteCustomCssHTMLResp SiteCustomCssHTMLReq

// SiteUsersResp site users response
type SiteUsersResp SiteUsersReq

// SiteThemeResp site theme response
type SiteThemeResp struct {
	ThemeOptions []*ThemeOption         `json:"theme_options"`
	Theme        string                 `json:"theme"`
	ThemeConfig  map[string]interface{} `json:"theme_config"`
	ColorScheme  string                 `json:"color_scheme"`
}

func (s *SiteThemeResp) TrTheme(ctx context.Context) {
	la := handler.GetLangByCtx(ctx)
	for _, option := range s.ThemeOptions {
		tr := translator.Tr(la, option.Value)
		// if tr is equal the option value means not found translation, so use the original label
		if tr != option.Value {
			option.Label = tr
		}
	}
}

// ThemeOption get label option
type ThemeOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// SiteWriteResp site write response
type SiteWriteResp SiteWriteReq

// SiteLegalResp site write response
type SiteLegalResp SiteLegalReq

// SiteLegalSimpleResp site write response
type SiteLegalSimpleResp struct {
	ExternalContentDisplay string `validate:"required,oneof=always_display ask_before_display" json:"external_content_display"`
}

// SiteSeoResp site write response
type SiteSeoResp SiteSeoReq

// SiteInfoResp get site info response
type SiteInfoResp struct {
	General       *SiteGeneralResp       `json:"general"`
	Interface     *SiteInterfaceResp     `json:"interface"`
	Branding      *SiteBrandingResp      `json:"branding"`
	Login         *SiteLoginResp         `json:"login"`
	Theme         *SiteThemeResp         `json:"theme"`
	CustomCssHtml *SiteCustomCssHTMLResp `json:"custom_css_html"`
	SiteSeo       *SiteSeoResp           `json:"site_seo"`
	SiteUsers     *SiteUsersResp         `json:"site_users"`
	Write         *SiteWriteResp         `json:"site_write"`
	Legal         *SiteLegalSimpleResp   `json:"site_legal"`
	Version       string                 `json:"version"`
	Revision      string                 `json:"revision"`
}
type TemplateSiteInfoResp struct {
	General       *SiteGeneralResp       `json:"general"`
	Interface     *SiteInterfaceResp     `json:"interface"`
	Branding      *SiteBrandingResp      `json:"branding"`
	SiteSeo       *SiteSeoResp           `json:"site_seo"`
	CustomCssHtml *SiteCustomCssHTMLResp `json:"custom_css_html"`
	Title         string
	Year          string
	Canonical     string
	JsonLD        string
	Keywords      string
	Description   string
}

// UpdateSMTPConfigReq get smtp config request
type UpdateSMTPConfigReq struct {
	FromEmail          string `validate:"omitempty,gt=0,lte=256" json:"from_email"`
	FromName           string `validate:"omitempty,gt=0,lte=256" json:"from_name"`
	SMTPHost           string `validate:"omitempty,gt=0,lte=256" json:"smtp_host"`
	SMTPPort           int    `validate:"omitempty,min=1,max=65535" json:"smtp_port"`
	Encryption         string `validate:"omitempty,oneof=SSL TLS" json:"encryption"` // "" SSL TLS
	SMTPUsername       string `validate:"omitempty,gt=0,lte=256" json:"smtp_username"`
	SMTPPassword       string `validate:"omitempty,gt=0,lte=256" json:"smtp_password"`
	SMTPAuthentication bool   `validate:"omitempty" json:"smtp_authentication"`
	TestEmailRecipient string `validate:"omitempty,email" json:"test_email_recipient"`
}

func (r *UpdateSMTPConfigReq) Check() (errField []*validator.FormErrorField, err error) {
	_, err = mail.ParseAddress(r.FromName)
	if err == nil {
		return append(errField, &validator.FormErrorField{
			ErrorField: "from_name",
			ErrorMsg:   reason.SMTPConfigFromNameCannotBeEmail,
		}), errors.BadRequest(reason.SMTPConfigFromNameCannotBeEmail)
	}
	return nil, nil
}

// GetSMTPConfigResp get smtp config response
type GetSMTPConfigResp struct {
	FromEmail          string `json:"from_email"`
	FromName           string `json:"from_name"`
	SMTPHost           string `json:"smtp_host"`
	SMTPPort           int    `json:"smtp_port"`
	Encryption         string `json:"encryption"` // "" SSL TLS
	SMTPUsername       string `json:"smtp_username"`
	SMTPPassword       string `json:"smtp_password"`
	SMTPAuthentication bool   `json:"smtp_authentication"`
}

// GetManifestJsonResp get manifest json response
type GetManifestJsonResp struct {
	ManifestVersion int                `json:"manifest_version"`
	Version         string             `json:"version"`
	Revision        string             `json:"revision"`
	ShortName       string             `json:"short_name"`
	Name            string             `json:"name"`
	Icons           []ManifestJsonIcon `json:"icons"`
	StartUrl        string             `json:"start_url"`
	Display         string             `json:"display"`
	ThemeColor      string             `json:"theme_color"`
	BackgroundColor string             `json:"background_color"`
}

type ManifestJsonIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

func CreateManifestJsonIcons(icon string) []ManifestJsonIcon {
	ext := filepath.Ext(icon)
	if ext == "" {
		ext = "png"
	} else {
		ext = strings.ToLower(ext[1:])
	}
	iconType := fmt.Sprintf("image/%s", ext)
	return []ManifestJsonIcon{
		{
			Src:   icon,
			Sizes: "16x16",
			Type:  iconType,
		},
		{
			Src:   icon,
			Sizes: "32x32",
			Type:  iconType,
		},
		{
			Src:   icon,
			Sizes: "48x48",
			Type:  iconType,
		},
		{
			Src:   icon,
			Sizes: "128x128",
			Type:  iconType,
		},
	}
}

const (
	// PrivilegeLevel1 low
	PrivilegeLevel1 PrivilegeLevel = 1
	// PrivilegeLevel2 medium
	PrivilegeLevel2 PrivilegeLevel = 2
	// PrivilegeLevel3 high
	PrivilegeLevel3 PrivilegeLevel = 3
	// PrivilegeLevelCustom custom
	PrivilegeLevelCustom PrivilegeLevel = 99
)

type PrivilegeLevel int
type PrivilegeOptions []*PrivilegeOption

func (p PrivilegeOptions) Choose(level PrivilegeLevel) (option *PrivilegeOption) {
	for _, op := range p {
		if op.Level == level {
			return op
		}
	}
	return nil
}

// GetPrivilegesConfigResp get privileges config response
type GetPrivilegesConfigResp struct {
	Options       []*PrivilegeOption `json:"options"`
	SelectedLevel PrivilegeLevel     `json:"selected_level"`
}

// PrivilegeOption privilege option
type PrivilegeOption struct {
	Level      PrivilegeLevel        `json:"level"`
	LevelDesc  string                `json:"level_desc"`
	Privileges []*constant.Privilege `validate:"dive" json:"privileges"`
}

// UpdatePrivilegesConfigReq update privileges config request
type UpdatePrivilegesConfigReq struct {
	Level            PrivilegeLevel        `validate:"required,min=1,max=3|eq=99" json:"level"`
	CustomPrivileges []*constant.Privilege `validate:"dive" json:"custom_privileges"`
}

var (
	DefaultPrivilegeOptions      PrivilegeOptions
	DefaultCustomPrivilegeOption *PrivilegeOption
	privilegeOptionsLevelMapping = map[string][]int{
		constant.RankQuestionAddKey:               {1, 1, 1},
		constant.RankAnswerAddKey:                 {1, 1, 1},
		constant.RankCommentAddKey:                {1, 1, 1},
		constant.RankReportAddKey:                 {1, 1, 1},
		constant.RankCommentVoteUpKey:             {1, 1, 1},
		constant.RankLinkUrlLimitKey:              {1, 10, 10},
		constant.RankQuestionVoteUpKey:            {1, 8, 15},
		constant.RankAnswerVoteUpKey:              {1, 8, 15},
		constant.RankQuestionVoteDownKey:          {125, 125, 125},
		constant.RankAnswerVoteDownKey:            {125, 125, 125},
		constant.RankInviteSomeoneToAnswerKey:     {1, 500, 1000},
		constant.RankTagAddKey:                    {1, 750, 1500},
		constant.RankTagEditKey:                   {1, 50, 100},
		constant.RankQuestionEditKey:              {1, 100, 200},
		constant.RankAnswerEditKey:                {1, 100, 200},
		constant.RankQuestionEditWithoutReviewKey: {1, 1000, 2000},
		constant.RankAnswerEditWithoutReviewKey:   {1, 1000, 2000},
		constant.RankQuestionAuditKey:             {1, 1000, 2000},
		constant.RankAnswerAuditKey:               {1, 1000, 2000},
		constant.RankTagAuditKey:                  {1, 2500, 5000},
		constant.RankTagEditWithoutReviewKey:      {1, 10000, 20000},
		constant.RankTagSynonymKey:                {1, 10000, 20000},
	}
)

func init() {
	DefaultPrivilegeOptions = append(DefaultPrivilegeOptions, &PrivilegeOption{
		Level:     PrivilegeLevel1,
		LevelDesc: reason.PrivilegeLevel1Desc,
	}, &PrivilegeOption{
		Level:     PrivilegeLevel2,
		LevelDesc: reason.PrivilegeLevel2Desc,
	}, &PrivilegeOption{
		Level:     PrivilegeLevel3,
		LevelDesc: reason.PrivilegeLevel3Desc,
	})

	for _, option := range DefaultPrivilegeOptions {
		for _, privilege := range constant.RankAllPrivileges {
			if len(privilegeOptionsLevelMapping[privilege.Key]) == 0 {
				continue
			}
			option.Privileges = append(option.Privileges, &constant.Privilege{
				Label: privilege.Label,
				Value: privilegeOptionsLevelMapping[privilege.Key][option.Level-1],
				Key:   privilege.Key,
			})
		}
	}

	// set up default custom privilege option
	DefaultCustomPrivilegeOption = &PrivilegeOption{
		Level:      PrivilegeLevelCustom,
		LevelDesc:  reason.PrivilegeLevelCustomDesc,
		Privileges: DefaultPrivilegeOptions[0].Privileges,
	}
}

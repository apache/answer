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

package siteinfo

import (
	"context"
	"encoding/json"
	errpkg "errors"
	"fmt"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/export"
	"github.com/apache/answer/internal/service/file_record"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	"github.com/apache/answer/plugin"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type SiteInfoService struct {
	siteInfoRepo          siteinfo_common.SiteInfoRepo
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
	emailService          *export.EmailService
	tagCommonService      *tagcommon.TagCommonService
	configService         *config.ConfigService
	questioncommon        *questioncommon.QuestionCommon
	fileRecordService     *file_record.FileRecordService
}

func NewSiteInfoService(
	siteInfoRepo siteinfo_common.SiteInfoRepo,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
	emailService *export.EmailService,
	tagCommonService *tagcommon.TagCommonService,
	configService *config.ConfigService,
	questioncommon *questioncommon.QuestionCommon,
	fileRecordService *file_record.FileRecordService,

) *SiteInfoService {
	plugin.RegisterGetSiteURLFunc(func() string {
		generalSiteInfo, err := siteInfoCommonService.GetSiteGeneral(context.Background())
		if err != nil {
			log.Error(err)
			return ""
		}
		return generalSiteInfo.SiteUrl
	})

	return &SiteInfoService{
		siteInfoRepo:          siteInfoRepo,
		siteInfoCommonService: siteInfoCommonService,
		emailService:          emailService,
		tagCommonService:      tagCommonService,
		configService:         configService,
		questioncommon:        questioncommon,
		fileRecordService:     fileRecordService,
	}
}

// GetSiteGeneral get site info general
func (s *SiteInfoService) GetSiteGeneral(ctx context.Context) (resp *schema.SiteGeneralResp, err error) {
	return s.siteInfoCommonService.GetSiteGeneral(ctx)
}

// GetSiteInterface get site info interface
func (s *SiteInfoService) GetSiteInterface(ctx context.Context) (resp *schema.SiteInterfaceResp, err error) {
	return s.siteInfoCommonService.GetSiteInterface(ctx)
}

// GetSiteBranding get site info branding
func (s *SiteInfoService) GetSiteBranding(ctx context.Context) (resp *schema.SiteBrandingResp, err error) {
	return s.siteInfoCommonService.GetSiteBranding(ctx)
}

// GetSiteUsers get site info about users
func (s *SiteInfoService) GetSiteUsers(ctx context.Context) (resp *schema.SiteUsersResp, err error) {
	return s.siteInfoCommonService.GetSiteUsers(ctx)
}

// GetSiteWrite get site info write
func (s *SiteInfoService) GetSiteWrite(ctx context.Context) (resp *schema.SiteWriteResp, err error) {
	resp = &schema.SiteWriteResp{}
	siteInfo, exist, err := s.siteInfoRepo.GetByType(ctx, constant.SiteTypeWrite)
	if err != nil {
		log.Error(err)
		return resp, nil
	}
	if exist {
		_ = json.Unmarshal([]byte(siteInfo.Content), resp)
	}

	resp.RecommendTags, err = s.tagCommonService.GetSiteWriteRecommendTag(ctx)
	if err != nil {
		log.Error(err)
	}
	resp.ReservedTags, err = s.tagCommonService.GetSiteWriteReservedTag(ctx)
	if err != nil {
		log.Error(err)
	}
	return resp, nil
}

// GetSiteLegal get site legal info
func (s *SiteInfoService) GetSiteLegal(ctx context.Context) (resp *schema.SiteLegalResp, err error) {
	return s.siteInfoCommonService.GetSiteLegal(ctx)
}

// GetSiteLogin get site login info
func (s *SiteInfoService) GetSiteLogin(ctx context.Context) (resp *schema.SiteLoginResp, err error) {
	return s.siteInfoCommonService.GetSiteLogin(ctx)
}

// GetSiteCustomCssHTML get site custom css html config
func (s *SiteInfoService) GetSiteCustomCssHTML(ctx context.Context) (resp *schema.SiteCustomCssHTMLResp, err error) {
	return s.siteInfoCommonService.GetSiteCustomCssHTML(ctx)
}

// GetSiteTheme get site theme config
func (s *SiteInfoService) GetSiteTheme(ctx context.Context) (resp *schema.SiteThemeResp, err error) {
	return s.siteInfoCommonService.GetSiteTheme(ctx)
}

func (s *SiteInfoService) SaveSiteGeneral(ctx context.Context, req schema.SiteGeneralReq) (err error) {
	req.FormatSiteUrl()
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeGeneral,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeGeneral, data)
}

func (s *SiteInfoService) SaveSiteInterface(ctx context.Context, req schema.SiteInterfaceReq) (err error) {
	// check language
	if !translator.CheckLanguageIsValid(req.Language) {
		err = errors.BadRequest(reason.LangNotFound)
		return
	}

	content, _ := json.Marshal(req)
	data := entity.SiteInfo{
		Type:    constant.SiteTypeInterface,
		Content: string(content),
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeInterface, &data)
}

// SaveSiteBranding save site branding information
func (s *SiteInfoService) SaveSiteBranding(ctx context.Context, req *schema.SiteBrandingReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeBranding,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeBranding, data)
}

// SaveSiteWrite save site configuration about write
func (s *SiteInfoService) SaveSiteWrite(ctx context.Context, req *schema.SiteWriteReq) (resp interface{}, err error) {
	recommendTags, reservedTags := make([]string, 0), make([]string, 0)
	recommendTagMapping, reservedTagMapping := make(map[string]bool), make(map[string]bool)
	for _, tag := range req.ReservedTags {
		if !recommendTagMapping[tag.SlugName] {
			reservedTagMapping[tag.SlugName] = true
			reservedTags = append(reservedTags, tag.SlugName)
		}
	}

	// recommend tags can't contain reserved tag
	for _, tag := range req.RecommendTags {
		if reservedTagMapping[tag.SlugName] {
			continue
		}
		if !recommendTagMapping[tag.SlugName] {
			recommendTagMapping[tag.SlugName] = true
			recommendTags = append(recommendTags, tag.SlugName)
		}
	}
	errData, err := s.tagCommonService.SetSiteWriteTag(ctx, recommendTags, reservedTags, req.UserID)
	if err != nil {
		return errData, err
	}

	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeWrite,
		Content: string(content),
		Status:  1,
	}
	return nil, s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeWrite, data)
}

// SaveSiteLegal save site legal configuration
func (s *SiteInfoService) SaveSiteLegal(ctx context.Context, req *schema.SiteLegalReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeLegal,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeLegal, data)
}

// SaveSiteLogin save site legal configuration
func (s *SiteInfoService) SaveSiteLogin(ctx context.Context, req *schema.SiteLoginReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeLogin,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeLogin, data)
}

// SaveSiteCustomCssHTML save site custom html configuration
func (s *SiteInfoService) SaveSiteCustomCssHTML(ctx context.Context, req *schema.SiteCustomCssHTMLReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeCustomCssHTML,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeCustomCssHTML, data)
}

// SaveSiteTheme save site custom html configuration
func (s *SiteInfoService) SaveSiteTheme(ctx context.Context, req *schema.SiteThemeReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeTheme,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeTheme, data)
}

// SaveSiteUsers save site users
func (s *SiteInfoService) SaveSiteUsers(ctx context.Context, req *schema.SiteUsersReq) (err error) {
	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypeUsers,
		Content: string(content),
		Status:  1,
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeUsers, data)
}

// GetSMTPConfig get smtp config
func (s *SiteInfoService) GetSMTPConfig(ctx context.Context) (resp *schema.GetSMTPConfigResp, err error) {
	emailConfig, err := s.emailService.GetEmailConfig(ctx)
	if err != nil {
		return nil, err
	}
	resp = &schema.GetSMTPConfigResp{}
	_ = copier.Copy(resp, emailConfig)
	resp.SMTPPassword = strings.Repeat("*", len(resp.SMTPPassword))
	return resp, nil
}

// UpdateSMTPConfig get smtp config
func (s *SiteInfoService) UpdateSMTPConfig(ctx context.Context, req *schema.UpdateSMTPConfigReq) (err error) {
	emailConfig, err := s.emailService.GetEmailConfig(ctx)
	if err != nil {
		return err
	}

	ec := &export.EmailConfig{}
	_ = copier.Copy(ec, req)

	if len(ec.SMTPPassword) > 0 && ec.SMTPPassword == strings.Repeat("*", len(ec.SMTPPassword)) {
		ec.SMTPPassword = emailConfig.SMTPPassword
	}

	err = s.emailService.SetEmailConfig(ctx, ec)
	if err != nil {
		return err
	}
	if len(req.TestEmailRecipient) > 0 {
		title, body, err := s.emailService.TestTemplate(ctx)
		if err != nil {
			return err
		}
		go s.emailService.Send(ctx, req.TestEmailRecipient, title, body)
	}
	return nil
}

func (s *SiteInfoService) GetSeo(ctx context.Context) (resp *schema.SiteSeoReq, err error) {
	resp = &schema.SiteSeoReq{}
	if err = s.siteInfoCommonService.GetSiteInfoByType(ctx, constant.SiteTypeSeo, resp); err != nil {
		return resp, err
	}
	loginConfig, err := s.GetSiteLogin(ctx)
	if err != nil {
		log.Error(err)
		return resp, nil
	}
	// If the site is set to privacy mode, prohibit crawling any page.
	if loginConfig.LoginRequired {
		resp.Robots = "User-agent: *\nDisallow: /"
		return resp, nil
	}
	return resp, nil
}

func (s *SiteInfoService) SaveSeo(ctx context.Context, req schema.SiteSeoReq) (err error) {
	content, _ := json.Marshal(req)
	data := entity.SiteInfo{
		Type:    constant.SiteTypeSeo,
		Content: string(content),
	}
	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeSeo, &data)
}

func (s *SiteInfoService) GetPrivilegesConfig(ctx context.Context) (resp *schema.GetPrivilegesConfigResp, err error) {
	privilege := &schema.UpdatePrivilegesConfigReq{}
	if err = s.siteInfoCommonService.GetSiteInfoByType(ctx, constant.SiteTypePrivileges, privilege); err != nil {
		return nil, err
	}
	privilegeOptions := schema.DefaultPrivilegeOptions
	if privilege.CustomPrivileges != nil && len(privilege.CustomPrivileges) > 0 {
		privilegeOptions = append(privilegeOptions, &schema.PrivilegeOption{
			Level:      schema.PrivilegeLevelCustom,
			LevelDesc:  reason.PrivilegeLevelCustomDesc,
			Privileges: privilege.CustomPrivileges,
		})
	} else {
		privilegeOptions = append(privilegeOptions, schema.DefaultCustomPrivilegeOption)
	}
	resp = &schema.GetPrivilegesConfigResp{
		Options:       s.translatePrivilegeOptions(ctx, privilegeOptions),
		SelectedLevel: schema.PrivilegeLevel3,
	}
	if privilege != nil && privilege.Level > 0 {
		resp.SelectedLevel = privilege.Level
	}
	return resp, nil
}

func (s *SiteInfoService) translatePrivilegeOptions(ctx context.Context, privilegeOptions []*schema.PrivilegeOption) (options []*schema.PrivilegeOption) {
	la := handler.GetLangByCtx(ctx)
	for _, option := range privilegeOptions {
		op := &schema.PrivilegeOption{
			Level:     option.Level,
			LevelDesc: translator.Tr(la, option.LevelDesc),
		}
		for _, privilege := range option.Privileges {
			op.Privileges = append(op.Privileges, &constant.Privilege{
				Key:   privilege.Key,
				Label: translator.Tr(la, privilege.Label),
				Value: privilege.Value,
			})
		}
		options = append(options, op)
	}
	return
}

func (s *SiteInfoService) UpdatePrivilegesConfig(ctx context.Context, req *schema.UpdatePrivilegesConfigReq) (err error) {
	var choosePrivileges []*constant.Privilege
	if req.Level == schema.PrivilegeLevelCustom {
		choosePrivileges = req.CustomPrivileges
	} else {
		chooseOption := schema.DefaultPrivilegeOptions.Choose(req.Level)
		if chooseOption == nil {
			return nil
		}
		choosePrivileges = chooseOption.Privileges
	}
	if choosePrivileges == nil {
		return nil
	}

	// update site info that user choose which privilege level
	if req.Level == schema.PrivilegeLevelCustom {
		privilegeMap := make(map[string]int)
		for _, privilege := range req.CustomPrivileges {
			privilegeMap[privilege.Key] = privilege.Value
		}
		var privileges []*constant.Privilege
		for _, privilege := range constant.RankAllPrivileges {
			privileges = append(privileges, &constant.Privilege{
				Key:   privilege.Key,
				Label: privilege.Label,
				Value: privilegeMap[privilege.Key],
			})
		}
		req.CustomPrivileges = privileges
	} else {
		privilege := &schema.UpdatePrivilegesConfigReq{}
		if err = s.siteInfoCommonService.GetSiteInfoByType(ctx, constant.SiteTypePrivileges, privilege); err != nil {
			return err
		}
		req.CustomPrivileges = privilege.CustomPrivileges
	}

	content, _ := json.Marshal(req)
	data := &entity.SiteInfo{
		Type:    constant.SiteTypePrivileges,
		Content: string(content),
		Status:  1,
	}
	err = s.siteInfoRepo.SaveByType(ctx, constant.SiteTypePrivileges, data)
	if err != nil {
		return err
	}

	// update privilege in config
	for _, privilege := range choosePrivileges {
		err = s.configService.UpdateConfig(ctx, privilege.Key, fmt.Sprintf("%d", privilege.Value))
		if err != nil {
			return err
		}
	}
	return
}

func (s *SiteInfoService) CleanUpRemovedBrandingFiles(
	ctx context.Context,
	newBranding *schema.SiteBrandingReq,
	currentBranding *schema.SiteBrandingResp,
) error {
	var allErrors []error
	currentFiles := map[string]string{
		"logo":        currentBranding.Logo,
		"mobile_logo": currentBranding.MobileLogo,
		"square_icon": currentBranding.SquareIcon,
		"favicon":     currentBranding.Favicon,
	}

	newFiles := map[string]string{
		"logo":        newBranding.Logo,
		"mobile_logo": newBranding.MobileLogo,
		"square_icon": newBranding.SquareIcon,
		"favicon":     newBranding.Favicon,
	}

	for key, currentFile := range currentFiles {
		newFile := newFiles[key]
		if currentFile != "" && currentFile != newFile {
			fileRecord, err := s.fileRecordService.GetFileRecordByURL(ctx, currentFile)
			if err != nil {
				allErrors = append(allErrors, err)
				continue
			}
			if fileRecord == nil {
				err := errpkg.New("file record is nil for key " + key)
				allErrors = append(allErrors, err)
				continue
			}
			if err := s.fileRecordService.DeleteAndMoveFileRecord(ctx, fileRecord); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}
	if len(allErrors) > 0 {
		return errpkg.Join(allErrors...)
	}
	return nil
}

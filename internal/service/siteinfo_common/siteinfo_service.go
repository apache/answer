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

package siteinfo_common

import (
	"context"
	"encoding/json"
	"html"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/gravatar"
	"github.com/segmentfault/pacman/log"
)

//go:generate mockgen -source=./siteinfo_service.go -destination=../mock/siteinfo_repo_mock.go -package=mock
type SiteInfoRepo interface {
	SaveByType(ctx context.Context, siteType string, data *entity.SiteInfo) (err error)
	GetByType(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo, exist bool, err error)
	IsBrandingFileUsed(ctx context.Context, filePath string) (bool, error)
}

// siteInfoCommonService site info common service
type siteInfoCommonService struct {
	siteInfoRepo SiteInfoRepo
}

type SiteInfoCommonService interface {
	GetSiteGeneral(ctx context.Context) (resp *schema.SiteGeneralResp, err error)
	GetSiteInterface(ctx context.Context) (resp *schema.SiteInterfaceResp, err error)
	GetSiteBranding(ctx context.Context) (resp *schema.SiteBrandingResp, err error)
	GetSiteUsers(ctx context.Context) (resp *schema.SiteUsersResp, err error)
	FormatAvatar(ctx context.Context, originalAvatarData, email string, userStatus int) *schema.AvatarInfo
	FormatListAvatar(ctx context.Context, userList []*entity.User) (userID2AvatarMapping map[string]*schema.AvatarInfo)
	GetSiteWrite(ctx context.Context) (resp *schema.SiteWriteResp, err error)
	GetSiteLegal(ctx context.Context) (resp *schema.SiteLegalResp, err error)
	GetSiteLogin(ctx context.Context) (resp *schema.SiteLoginResp, err error)
	GetSiteCustomCssHTML(ctx context.Context) (resp *schema.SiteCustomCssHTMLResp, err error)
	GetSiteTheme(ctx context.Context) (resp *schema.SiteThemeResp, err error)
	GetSiteSeo(ctx context.Context) (resp *schema.SiteSeoResp, err error)
	GetSiteInfoByType(ctx context.Context, siteType string, resp interface{}) (err error)
	IsBrandingFileUsed(ctx context.Context, filePath string) bool
}

// NewSiteInfoCommonService new site info common service
func NewSiteInfoCommonService(siteInfoRepo SiteInfoRepo) SiteInfoCommonService {
	return &siteInfoCommonService{
		siteInfoRepo: siteInfoRepo,
	}
}

// GetSiteGeneral get site info general
func (s *siteInfoCommonService) GetSiteGeneral(ctx context.Context) (resp *schema.SiteGeneralResp, err error) {
	resp = &schema.SiteGeneralResp{CheckUpdate: true}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeGeneral, resp); err != nil {
		return nil, err
	}
	resp.Name = html.UnescapeString(resp.Name)
	return resp, nil
}

// GetSiteInterface get site info interface
func (s *siteInfoCommonService) GetSiteInterface(ctx context.Context) (resp *schema.SiteInterfaceResp, err error) {
	resp = &schema.SiteInterfaceResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeInterface, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteBranding get site info branding
func (s *siteInfoCommonService) GetSiteBranding(ctx context.Context) (resp *schema.SiteBrandingResp, err error) {
	resp = &schema.SiteBrandingResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeBranding, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteUsers get site info about users
func (s *siteInfoCommonService) GetSiteUsers(ctx context.Context) (resp *schema.SiteUsersResp, err error) {
	resp = &schema.SiteUsersResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeUsers, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// FormatAvatar format avatar
func (s *siteInfoCommonService) FormatAvatar(ctx context.Context, originalAvatarData, email string, userStatus int) *schema.AvatarInfo {
	gravatarBaseURL, defaultAvatar := s.getAvatarDefaultConfig(ctx)
	return s.selectedAvatar(originalAvatarData, defaultAvatar, gravatarBaseURL, email, userStatus)
}

// FormatListAvatar format avatar
func (s *siteInfoCommonService) FormatListAvatar(ctx context.Context, userList []*entity.User) (
	avatarMapping map[string]*schema.AvatarInfo) {
	gravatarBaseURL, defaultAvatar := s.getAvatarDefaultConfig(ctx)
	avatarMapping = make(map[string]*schema.AvatarInfo)
	for _, user := range userList {
		avatarMapping[user.ID] = s.selectedAvatar(user.Avatar, defaultAvatar, gravatarBaseURL, user.EMail, user.Status)
	}
	return avatarMapping
}

func (s *siteInfoCommonService) getAvatarDefaultConfig(ctx context.Context) (string, string) {
	gravatarBaseURL, defaultAvatar := constant.DefaultGravatarBaseURL, constant.DefaultAvatar
	usersConfig, err := s.GetSiteUsers(ctx)
	if err != nil {
		log.Error(err)
	}
	if len(usersConfig.GravatarBaseURL) > 0 {
		gravatarBaseURL = usersConfig.GravatarBaseURL
	}
	if len(usersConfig.DefaultAvatar) > 0 {
		defaultAvatar = usersConfig.DefaultAvatar
	}
	return gravatarBaseURL, defaultAvatar
}

func (s *siteInfoCommonService) selectedAvatar(
	originalAvatarData,
	defaultAvatar, gravatarBaseURL,
	email string, userStatus int) *schema.AvatarInfo {
	avatarInfo := &schema.AvatarInfo{}
	_ = json.Unmarshal([]byte(originalAvatarData), avatarInfo)

	if userStatus == entity.UserStatusDeleted {
		return &schema.AvatarInfo{
			Type: constant.DefaultAvatar,
		}
	}

	if len(avatarInfo.Type) == 0 && defaultAvatar == constant.AvatarTypeGravatar {
		avatarInfo.Type = constant.AvatarTypeGravatar
		avatarInfo.Gravatar = gravatar.GetAvatarURL(gravatarBaseURL, email)
	} else if avatarInfo.Type == constant.AvatarTypeGravatar {
		avatarInfo.Gravatar = gravatar.GetAvatarURL(gravatarBaseURL, email)
	}
	return avatarInfo
}

// GetSiteWrite get site info write
func (s *siteInfoCommonService) GetSiteWrite(ctx context.Context) (resp *schema.SiteWriteResp, err error) {
	resp = &schema.SiteWriteResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeWrite, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteLegal get site info write
func (s *siteInfoCommonService) GetSiteLegal(ctx context.Context) (resp *schema.SiteLegalResp, err error) {
	resp = &schema.SiteLegalResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeLegal, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteLogin get site login config
func (s *siteInfoCommonService) GetSiteLogin(ctx context.Context) (resp *schema.SiteLoginResp, err error) {
	resp = &schema.SiteLoginResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeLogin, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteCustomCssHTML get site custom css html config
func (s *siteInfoCommonService) GetSiteCustomCssHTML(ctx context.Context) (resp *schema.SiteCustomCssHTMLResp, err error) {
	resp = &schema.SiteCustomCssHTMLResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeCustomCssHTML, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSiteTheme get site theme
func (s *siteInfoCommonService) GetSiteTheme(ctx context.Context) (resp *schema.SiteThemeResp, err error) {
	resp = &schema.SiteThemeResp{
		ThemeOptions: schema.GetThemeOptions,
	}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeTheme, resp); err != nil {
		return nil, err
	}
	resp.TrTheme(ctx)
	return resp, nil
}

// GetSiteSeo get site seo
func (s *siteInfoCommonService) GetSiteSeo(ctx context.Context) (resp *schema.SiteSeoResp, err error) {
	resp = &schema.SiteSeoResp{}
	if err = s.GetSiteInfoByType(ctx, constant.SiteTypeSeo, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *siteInfoCommonService) EnableShortID(ctx context.Context) (enabled bool) {
	siteSeo, err := s.GetSiteSeo(ctx)
	if err != nil {
		log.Error(err)
		return false
	}
	return siteSeo.IsShortLink()
}

func (s *siteInfoCommonService) GetSiteInfoByType(ctx context.Context, siteType string, resp interface{}) (err error) {
	siteInfo, exist, err := s.siteInfoRepo.GetByType(ctx, siteType)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	_ = json.Unmarshal([]byte(siteInfo.Content), resp)
	return nil
}

func (s *siteInfoCommonService) IsBrandingFileUsed(ctx context.Context, filePath string) bool {
	used, err := s.siteInfoRepo.IsBrandingFileUsed(ctx, filePath)
	if err != nil {
		log.Errorf("error checking if branding file is used: %v", err)
		// will try again with the next clean up
		return true
	}
	return used
}

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
	"fmt"
	"time"

	"github.com/apache/answer/internal/base/constant"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

type Mentor struct {
	ctx      context.Context
	engine   *xorm.Engine
	userData *InitNeedUserInputData
	err      error
	Done     bool
}

func NewMentor(ctx context.Context, engine *xorm.Engine, data *InitNeedUserInputData) *Mentor {
	return &Mentor{ctx: ctx, engine: engine, userData: data}
}

type InitNeedUserInputData struct {
	Language               string
	SiteName               string
	SiteURL                string
	ContactEmail           string
	AdminName              string
	AdminPassword          string
	AdminEmail             string
	LoginRequired          bool
	ExternalContentDisplay string
}

func (m *Mentor) InitDB() error {
	m.do("check table exist", m.checkTableExist)
	m.do("sync table", m.syncTable)
	m.do("init version table", m.initVersionTable)
	m.do("init admin user", m.initAdminUser)
	m.do("init config", m.initConfig)
	m.do("init default privileges config", m.initDefaultRankPrivileges)
	m.do("init role", m.initRole)
	m.do("init power", m.initPower)
	m.do("init role power rel", m.initRolePowerRel)
	m.do("init admin user role rel", m.initAdminUserRoleRel)
	m.do("init site info interface", m.initSiteInfoInterface)
	m.do("init site info general config", m.initSiteInfoGeneralData)
	m.do("init site info login config", m.initSiteInfoLoginConfig)
	m.do("init site info theme config", m.initSiteInfoThemeConfig)
	m.do("init site info seo config", m.initSiteInfoSEOConfig)
	m.do("init site info user config", m.initSiteInfoUsersConfig)
	m.do("init site info privilege rank", m.initSiteInfoPrivilegeRank)
	m.do("init site info write", m.initSiteInfoWrite)
	m.do("init site info legal", m.initSiteInfoLegalConfig)
	m.do("init default content", m.initDefaultContent)
	m.do("init default badges", m.initDefaultBadges)
	return m.err
}

func (m *Mentor) do(taskName string, fn func()) {
	if m.err != nil || m.Done {
		return
	}
	fn()
	if m.err != nil {
		m.err = fmt.Errorf("%s failed: %s", taskName, m.err)
	}
}

func (m *Mentor) checkTableExist() {
	m.Done, m.err = m.engine.Context(m.ctx).IsTableExist(&entity.Version{})
	if m.Done {
		fmt.Println("[database] already exists")
	}
}

func (m *Mentor) syncTable() {
	m.err = m.engine.Context(m.ctx).Sync(tables...)
}

func (m *Mentor) initVersionTable() {
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.Version{ID: 1, VersionNumber: ExpectedVersion()})
}

func (m *Mentor) initAdminUser() {
	generateFromPassword, _ := bcrypt.GenerateFromPassword([]byte(m.userData.AdminPassword), bcrypt.DefaultCost)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.User{
		ID:           "1",
		Username:     m.userData.AdminName,
		Pass:         string(generateFromPassword),
		EMail:        m.userData.AdminEmail,
		MailStatus:   1,
		NoticeStatus: 1,
		Status:       1,
		Rank:         1,
		DisplayName:  m.userData.AdminName,
	})
}

func (m *Mentor) initConfig() {
	_, m.err = m.engine.Context(m.ctx).Insert(defaultConfigTable)
}

func (m *Mentor) initDefaultRankPrivileges() {
	chooseOption := schema.DefaultPrivilegeOptions.Choose(schema.PrivilegeLevel2)
	for _, privilege := range chooseOption.Privileges {
		_, err := m.engine.Context(m.ctx).Update(
			&entity.Config{Value: fmt.Sprintf("%d", privilege.Value)},
			&entity.Config{Key: privilege.Key},
		)
		if err != nil {
			log.Error(err)
		}
	}
}

func (m *Mentor) initRole() {
	_, m.err = m.engine.Context(m.ctx).Insert(roles)
}

func (m *Mentor) initPower() {
	_, m.err = m.engine.Context(m.ctx).Insert(powers)
}

func (m *Mentor) initRolePowerRel() {
	_, m.err = m.engine.Context(m.ctx).Insert(rolePowerRels)
}

func (m *Mentor) initAdminUserRoleRel() {
	_, m.err = m.engine.Context(m.ctx).Insert(adminUserRoleRel)
}

func (m *Mentor) initSiteInfoInterface() {
	now := time.Now()
	zoneName, offset := now.In(time.Local).Zone()

	localTimezone := "UTC"
	for _, tz := range constant.Timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			continue
		}

		tzNow := now.In(loc)
		tzName, tzOffset := tzNow.Zone()

		if tzName == zoneName && tzOffset == offset {
			localTimezone = tz
			break
		}
	}

	interfaceData := map[string]string{
		"language":          m.userData.Language,
		"time_zone":         localTimezone,
		"default_avatar":    "gravatar",
		"gravatar_base_url": "https://www.gravatar.com/avatar/",
	}
	interfaceDataBytes, _ := json.Marshal(interfaceData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "interface",
		Content: string(interfaceDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoGeneralData() {
	generalData := map[string]string{
		"name":          m.userData.SiteName,
		"site_url":      m.userData.SiteURL,
		"contact_email": m.userData.ContactEmail,
	}
	generalDataBytes, _ := json.Marshal(generalData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "general",
		Content: string(generalDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoLoginConfig() {
	loginConfig := map[string]interface{}{
		"allow_new_registrations":   true,
		"allow_email_registrations": true,
		"allow_password_login":      true,
		"login_required":            m.userData.LoginRequired,
	}
	loginConfigDataBytes, _ := json.Marshal(loginConfig)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "login",
		Content: string(loginConfigDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoLegalConfig() {
	legalConfig := map[string]interface{}{
		"external_content_display": m.userData.ExternalContentDisplay,
	}
	legalConfigDataBytes, _ := json.Marshal(legalConfig)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "legal",
		Content: string(legalConfigDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoThemeConfig() {
	themeConfig := `{"theme":"default","theme_config":{"default":{"navbar_style":"#0033ff","primary_color":"#0033ff"}}}`
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "theme",
		Content: themeConfig,
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoSEOConfig() {
	seoData := map[string]interface{}{
		"permalink": constant.PermalinkQuestionID,
		"robots":    defaultSEORobotTxt + m.userData.SiteURL + "/sitemap.xml",
	}
	seoDataBytes, _ := json.Marshal(seoData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "seo",
		Content: string(seoDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoUsersConfig() {
	usersData := map[string]any{
		"default_avatar":            "gravatar",
		"gravatar_base_url":         "https://www.gravatar.com/avatar/",
		"allow_update_display_name": true,
		"allow_update_username":     true,
		"allow_update_avatar":       true,
		"allow_update_bio":          true,
		"allow_update_website":      true,
		"allow_update_location":     true,
	}
	usersDataBytes, _ := json.Marshal(usersData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "users",
		Content: string(usersDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoPrivilegeRank() {
	privilegeRankData := map[string]interface{}{
		"level": schema.PrivilegeLevel2,
	}
	privilegeRankDataBytes, _ := json.Marshal(privilegeRankData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "privileges",
		Content: string(privilegeRankDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initSiteInfoWrite() {
	writeData := map[string]interface{}{
		"min_content":                      6,
		"restrict_answer":                  true,
		"min_tags":                         1,
		"required_tag":                     false,
		"recommend_tags":                   []string{},
		"reserved_tags":                    []string{},
		"max_image_size":                   4,
		"max_attachment_size":              8,
		"max_image_megapixel":              40,
		"authorized_image_extensions":      []string{"jpg", "jpeg", "png", "gif", "webp"},
		"authorized_attachment_extensions": []string{},
	}
	writeDataBytes, _ := json.Marshal(writeData)
	_, m.err = m.engine.Context(m.ctx).Insert(&entity.SiteInfo{
		Type:    "write",
		Content: string(writeDataBytes),
		Status:  1,
	})
}

func (m *Mentor) initDefaultContent() {
	uniqueIDRepo := unique.NewUniqueIDRepo(&data.Data{DB: m.engine})
	now := time.Now()

	tagId, err := uniqueIDRepo.GenUniqueIDStr(m.ctx, entity.Tag{}.TableName())
	if err != nil {
		m.err = err
		return
	}

	q1Id, err := uniqueIDRepo.GenUniqueIDStr(m.ctx, entity.Question{}.TableName())
	if err != nil {
		m.err = err
		return
	}

	a1Id, err := uniqueIDRepo.GenUniqueIDStr(m.ctx, entity.Answer{}.TableName())
	if err != nil {
		m.err = err
		return
	}

	q2Id, err := uniqueIDRepo.GenUniqueIDStr(m.ctx, entity.Question{}.TableName())
	if err != nil {
		m.err = err
		return
	}

	a2Id, err := uniqueIDRepo.GenUniqueIDStr(m.ctx, entity.Answer{}.TableName())
	if err != nil {
		m.err = err
		return
	}

	tag := entity.Tag{
		ID:            tagId,
		SlugName:      "support",
		DisplayName:   "support",
		OriginalText:  "For general support questions.",
		ParsedText:    "<p>For general support questions.</p>",
		UserID:        "1",
		QuestionCount: 2,
		Status:        entity.TagStatusAvailable,
		RevisionID:    "0",
	}

	q1 := &entity.Question{
		ID:               q1Id,
		CreatedAt:        now,
		UserID:           "1",
		LastEditUserID:   "1",
		Title:            "What is a tag?",
		OriginalText:     "When asking a question, we need to choose tags. What are tags and why should I use them?",
		ParsedText:       "<p>When asking a question, we need to choose tags. What are tags and why should I use them?</p>",
		Pin:              entity.QuestionUnPin,
		Show:             entity.QuestionShow,
		Status:           entity.QuestionStatusAvailable,
		AnswerCount:      1,
		AcceptedAnswerID: "0",
		LastAnswerID:     a1Id,
		PostUpdateTime:   now,
		RevisionID:       "0",
	}

	a1 := &entity.Answer{
		ID:             a1Id,
		CreatedAt:      now,
		QuestionID:     q1Id,
		UserID:         "1",
		LastEditUserID: "0",
		OriginalText:   "Tags help to organize content and make searching easier. It helps your question get more attention from people interested in that tag. Tags also send notifications. If you are interested in some topic, follow that tag to get updates.",
		ParsedText:     "<p>Tags help to organize content and make searching easier. It helps your question get more attention from people interested in that tag. Tags also send notifications. If you are interested in some topic, follow that tag to get updates.</p>",
		Status:         entity.AnswerStatusAvailable,
		RevisionID:     "0",
	}

	q2 := &entity.Question{
		ID:               q2Id,
		CreatedAt:        now,
		UserID:           "1",
		LastEditUserID:   "1",
		Title:            "What is reputation and how do I earn them?",
		OriginalText:     "I see that each user has reputation points, What is it and how do I earn them?",
		ParsedText:       "<p>I see that each user has reputation points, What is it and how do I earn them?</p>",
		Pin:              entity.QuestionUnPin,
		Show:             entity.QuestionShow,
		Status:           entity.QuestionStatusAvailable,
		AnswerCount:      1,
		AcceptedAnswerID: "0",
		LastAnswerID:     a2Id,
		PostUpdateTime:   now,
		RevisionID:       "0",
	}

	a2 := &entity.Answer{
		ID:             a2Id,
		CreatedAt:      now,
		QuestionID:     q2Id,
		UserID:         "1",
		LastEditUserID: "0",
		OriginalText:   "Your reputation points show how much the community values your knowledge. You earn points when someone find your question or answer helpful. You also get points when the person who asked the question thinks you did a good job and accepts your answer.",
		ParsedText:     "<p>Your reputation points show how much the community values your knowledge. You earn points when someone find your question or answer helpful. You also get points when the person who asked the question thinks you did a good job and accepts your answer.</p>",
		Status:         entity.AnswerStatusAvailable,
		RevisionID:     "0",
	}

	_, m.err = m.engine.Context(m.ctx).Insert(tag)
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(q1)
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(a1)
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(entity.TagRel{
		ObjectID: q1.ID,
		TagID:    tag.ID,
		Status:   entity.TagRelStatusAvailable,
	})
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(q2)
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(a2)
	if m.err != nil {
		return
	}

	_, m.err = m.engine.Context(m.ctx).Insert(entity.TagRel{
		ObjectID: q2.ID,
		TagID:    tag.ID,
		Status:   entity.TagRelStatusAvailable,
	})
	if m.err != nil {
		return
	}
}

func (m *Mentor) initDefaultBadges() {
	uniqueIDRepo := unique.NewUniqueIDRepo(&data.Data{DB: m.engine})

	_, m.err = m.engine.Context(m.ctx).Insert(defaultBadgeGroupTable)
	if m.err != nil {
		return
	}
	for _, badge := range defaultBadgeTable {
		badge.ID, m.err = uniqueIDRepo.GenUniqueIDStr(m.ctx, new(entity.Badge).TableName())
		if m.err != nil {
			return
		}
		if _, m.err = m.engine.Context(m.ctx).Insert(badge); m.err != nil {
			return
		}
	}
}

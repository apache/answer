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

package repo

import (
	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/repo/activity"
	"github.com/apache/incubator-answer/internal/repo/activity_common"
	"github.com/apache/incubator-answer/internal/repo/answer"
	"github.com/apache/incubator-answer/internal/repo/auth"
	"github.com/apache/incubator-answer/internal/repo/captcha"
	"github.com/apache/incubator-answer/internal/repo/collection"
	"github.com/apache/incubator-answer/internal/repo/comment"
	"github.com/apache/incubator-answer/internal/repo/config"
	"github.com/apache/incubator-answer/internal/repo/export"
	"github.com/apache/incubator-answer/internal/repo/limit"
	"github.com/apache/incubator-answer/internal/repo/meta"
	"github.com/apache/incubator-answer/internal/repo/notification"
	"github.com/apache/incubator-answer/internal/repo/plugin_config"
	"github.com/apache/incubator-answer/internal/repo/question"
	"github.com/apache/incubator-answer/internal/repo/rank"
	"github.com/apache/incubator-answer/internal/repo/reason"
	"github.com/apache/incubator-answer/internal/repo/report"
	"github.com/apache/incubator-answer/internal/repo/review"
	"github.com/apache/incubator-answer/internal/repo/revision"
	"github.com/apache/incubator-answer/internal/repo/role"
	"github.com/apache/incubator-answer/internal/repo/search_common"
	"github.com/apache/incubator-answer/internal/repo/site_info"
	"github.com/apache/incubator-answer/internal/repo/tag"
	"github.com/apache/incubator-answer/internal/repo/tag_common"
	"github.com/apache/incubator-answer/internal/repo/unique"
	"github.com/apache/incubator-answer/internal/repo/user"
	"github.com/apache/incubator-answer/internal/repo/user_external_login"
	"github.com/apache/incubator-answer/internal/repo/user_notification_config"
	"github.com/google/wire"
)

// ProviderSetRepo is data providers.
var ProviderSetRepo = wire.NewSet(
	data.NewData,
	data.NewDB,
	data.NewCache,
	comment.NewCommentRepo,
	comment.NewCommentCommonRepo,
	captcha.NewCaptchaRepo,
	unique.NewUniqueIDRepo,
	report.NewReportRepo,
	activity_common.NewFollowRepo,
	activity_common.NewVoteRepo,
	config.NewConfigRepo,
	user.NewUserRepo,
	user.NewUserAdminRepo,
	rank.NewUserRankRepo,
	question.NewQuestionRepo,
	answer.NewAnswerRepo,
	activity_common.NewActivityRepo,
	activity.NewVoteRepo,
	activity.NewFollowRepo,
	activity.NewAnswerActivityRepo,
	activity.NewUserActiveActivityRepo,
	activity.NewActivityRepo,
	activity.NewReviewActivityRepo,
	tag.NewTagRepo,
	tag_common.NewTagCommonRepo,
	tag.NewTagRelRepo,
	collection.NewCollectionRepo,
	collection.NewCollectionGroupRepo,
	auth.NewAuthRepo,
	revision.NewRevisionRepo,
	search_common.NewSearchRepo,
	meta.NewMetaRepo,
	export.NewEmailRepo,
	reason.NewReasonRepo,
	site_info.NewSiteInfo,
	notification.NewNotificationRepo,
	role.NewRoleRepo,
	role.NewUserRoleRelRepo,
	role.NewRolePowerRelRepo,
	role.NewPowerRepo,
	user_external_login.NewUserExternalLoginRepo,
	plugin_config.NewPluginConfigRepo,
	user_notification_config.NewUserNotificationConfigRepo,
	limit.NewRateLimitRepo,
	plugin_config.NewPluginUserConfigRepo,
	review.NewReviewRepo,
)

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

package reason

const (
	// Success .
	Success = "base.success"
	// UnknownError unknown error
	UnknownError = "base.unknown"
	// RequestFormatError request format error
	RequestFormatError = "base.request_format_error"
	// UnauthorizedError unauthorized error
	UnauthorizedError = "base.unauthorized_error"
	// DatabaseError database error
	DatabaseError = "base.database_error"
	// ForbiddenError forbidden error
	ForbiddenError = "base.forbidden_error"
	// DuplicateRequestError duplicate request error
	DuplicateRequestError = "base.duplicate_request_error"
)

const (
	EmailOrPasswordWrong             = "error.object.email_or_password_incorrect"
	CommentNotFound                  = "error.comment.not_found"
	CommentCannotEditAfterDeadline   = "error.comment.cannot_edit_after_deadline"
	QuestionNotFound                 = "error.question.not_found"
	QuestionCannotDeleted            = "error.question.cannot_deleted"
	QuestionCannotClose              = "error.question.cannot_close"
	QuestionCannotUpdate             = "error.question.cannot_update"
	QuestionAlreadyDeleted           = "error.question.already_deleted"
	QuestionUnderReview              = "error.question.under_review"
	QuestionContentCannotEmpty       = "error.question.content_cannot_empty"
	QuestionContentLessThanMinimum   = "error.question.content_less_than_minumum"
	AnswerNotFound                   = "error.answer.not_found"
	AnswerCannotDeleted              = "error.answer.cannot_deleted"
	AnswerCannotUpdate               = "error.answer.cannot_update"
	AnswerCannotAddByClosedQuestion  = "error.answer.question_closed_cannot_add"
	AnswerRestrictAnswer             = "error.answer.restrict_answer"
	AnswerContentCannotEmpty         = "error.answer.content_cannot_empty"
	CommentEditWithoutPermission     = "error.comment.edit_without_permission"
	CommentContentCannotEmpty        = "error.comment.content_cannot_empty"
	DisallowVote                     = "error.object.disallow_vote"
	DisallowFollow                   = "error.object.disallow_follow"
	DisallowVoteYourSelf             = "error.object.disallow_vote_your_self"
	CaptchaVerificationFailed        = "error.object.captcha_verification_failed"
	OldPasswordVerificationFailed    = "error.object.old_password_verification_failed"
	NewPasswordSameAsPreviousSetting = "error.object.new_password_same_as_previous_setting"
	NewObjectAlreadyDeleted          = "error.object.already_deleted"
	UserNotFound                     = "error.user.not_found"
	UsernameInvalid                  = "error.user.username_invalid"
	UsernameDuplicate                = "error.user.username_duplicate"
	UserSetAvatar                    = "error.user.set_avatar"
	EmailDuplicate                   = "error.email.duplicate"
	EmailVerifyURLExpired            = "error.email.verify_url_expired"
	EmailNeedToBeVerified            = "error.email.need_to_be_verified"
	EmailIllegalDomainError          = "error.email.illegal_email_domain_error"
	UserSuspended                    = "error.user.suspended"
	ObjectNotFound                   = "error.object.not_found"
	TagNotFound                      = "error.tag.not_found"
	TagNotContainSynonym             = "error.tag.not_contain_synonym_tags"
	TagCannotUpdate                  = "error.tag.cannot_update"
	TagIsUsedCannotDelete            = "error.tag.is_used_cannot_delete"
	TagAlreadyExist                  = "error.tag.already_exist"
	TagMinCount                      = "error.tag.minimum_count"
	RankFailToMeetTheCondition       = "error.rank.fail_to_meet_the_condition"
	VoteRankFailToMeetTheCondition   = "error.rank.vote_fail_to_meet_the_condition"
	NoEnoughRankToOperate            = "error.rank.no_enough_rank_to_operate"
	ThemeNotFound                    = "error.theme.not_found"
	LangNotFound                     = "error.lang.not_found"
	ReportHandleFailed               = "error.report.handle_failed"
	ReportNotFound                   = "error.report.not_found"
	ReadConfigFailed                 = "error.config.read_config_failed"
	DatabaseConnectionFailed         = "error.database.connection_failed"
	InstallCreateTableFailed         = "error.database.create_table_failed"
	InstallConfigFailed              = "error.install.create_config_failed"
	SiteInfoConfigNotFound           = "error.site_info.config_not_found"
	UploadFileSourceUnsupported      = "error.upload.source_unsupported"
	UploadFileUnsupportedFileFormat  = "error.upload.unsupported_file_format"
	RecommendTagNotExist             = "error.tag.recommend_tag_not_found"
	RecommendTagEnter                = "error.tag.recommend_tag_enter"
	RevisionReviewUnderway           = "error.revision.review_underway"
	RevisionNoPermission             = "error.revision.no_permission"
	UserCannotUpdateYourRole         = "error.user.cannot_update_your_role"
	TagCannotSetSynonymAsItself      = "error.tag.cannot_set_synonym_as_itself"
	NotAllowedRegistration           = "error.user.not_allowed_registration"
	NotAllowedLoginViaPassword       = "error.user.not_allowed_login_via_password"
	SMTPConfigFromNameCannotBeEmail  = "error.smtp.config_from_name_cannot_be_email"
	AdminCannotUpdateTheirPassword   = "error.admin.cannot_update_their_password"
	AdminCannotEditTheirProfile      = "error.admin.cannot_edit_their_profile"
	AdminCannotModifySelfStatus      = "error.admin.cannot_modify_self_status"
	UserAccessDenied                 = "error.user.access_denied"
	UserPageAccessDenied             = "error.user.page_access_denied"
	AddBulkUsersFormatError          = "error.user.add_bulk_users_format_error"
	AddBulkUsersAmountError          = "error.user.add_bulk_users_amount_error"
	InvalidURLError                  = "error.common.invalid_url"
	MetaObjectNotFound               = "error.meta.object_not_found"
	BadgeObjectNotFound              = "error.badge.object_not_found"
	StatusInvalid                    = "error.common.status_invalid"
	UserStatusInactive               = "error.user.status_inactive"
	UserStatusSuspendedForever       = "error.user.status_suspended_forever"
	UserStatusSuspendedUntil         = "error.user.status_suspended_until"
	UserStatusDeleted                = "error.user.status_deleted"
)

// user external login reasons
const (
	UserExternalLoginUnbindingForbidden = "error.user.external_login_unbinding_forbidden"
	UserExternalLoginMissingUserID      = "error.user.external_login_missing_user_id"
)

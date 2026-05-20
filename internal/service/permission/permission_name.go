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

package permission

const (
	AdminAccess                 = "admin.access"
	QuestionAdd                 = "question.add"
	QuestionEdit                = "question.edit"
	QuestionEditWithoutReview   = "question.edit_without_review"
	QuestionDelete              = "question.delete"
	QuestionClose               = "question.close"
	QuestionReopen              = "question.reopen"
	QuestionVoteUp              = "question.vote_up"
	QuestionVoteDown            = "question.vote_down"
	QuestionPin                 = "question.pin"
	QuestionUnPin               = "question.unpin"
	QuestionHide                = "question.hide"
	QuestionShow                = "question.show"
	AnswerAdd                   = "answer.add"
	AnswerEdit                  = "answer.edit"
	AnswerEditWithoutReview     = "answer.edit_without_review"
	AnswerDelete                = "answer.delete"
	AnswerAccept                = "answer.accept"
	AnswerVoteUp                = "answer.vote_up"
	AnswerVoteDown              = "answer.vote_down"
	AnswerInviteSomeoneToAnswer = "answer.invite_someone_to_answer"
	CommentAdd                  = "comment.add"
	CommentEdit                 = "comment.edit"
	CommentDelete               = "comment.delete"
	CommentVoteUp               = "comment.vote_up"
	CommentVoteDown             = "comment.vote_down"
	ReportAdd                   = "report.add"
	TagAdd                      = "tag.add"
	TagEdit                     = "tag.edit"
	TagEditSlugName             = "tag.edit_slug_name"
	TagEditWithoutReview        = "tag.edit_without_review"
	TagDelete                   = "tag.delete"
	TagMerge                    = "tag.merge"
	TagSynonym                  = "tag.synonym"
	LinkUrlLimit                = "link.url_limit"
	VoteDetail                  = "vote.detail"
	AnswerAudit                 = "answer.audit"
	QuestionAudit               = "question.audit"
	TagAudit                    = "tag.audit"
	TagUseReservedTag           = "tag.use_reserved_tag"
	AnswerUnDelete              = "answer.undeleted"
	QuestionUnDelete            = "question.undeleted"
	TagUnDelete                 = "tag.undeleted"
)

const (
	reportActionName                = "action.report"
	editActionName                  = "action.edit"
	deleteActionName                = "action.delete"
	mergeActionName                 = "action.merge"
	undeleteActionName              = "action.undelete"
	closeActionName                 = "action.close"
	reopenActionName                = "action.reopen"
	pinActionName                   = "action.pin"
	unpinActionName                 = "action.unpin"
	hideActionName                  = "action.hide"
	showActionName                  = "action.show"
	inviteSomeoneToAnswerActionName = "action.invite_someone_to_answer"
)

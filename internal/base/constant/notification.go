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

package constant

const (
	// NotificationUpdateQuestion update question
	NotificationUpdateQuestion = "notification.action.update_question"
	// NotificationAnswerTheQuestion answer the question
	NotificationAnswerTheQuestion = "notification.action.answer_the_question"
	// NotificationUpVotedTheQuestion up voted the question
	NotificationUpVotedTheQuestion = "notification.action.up_voted_question"
	// NotificationDownVotedTheQuestion down voted the question
	NotificationDownVotedTheQuestion = "notification.action.down_voted_question"
	// NotificationUpdateAnswer update answer
	NotificationUpdateAnswer = "notification.action.update_answer"
	// NotificationAcceptAnswer accept answer
	NotificationAcceptAnswer = "notification.action.accept_answer"
	// NotificationUpVotedTheAnswer up voted the answer
	NotificationUpVotedTheAnswer = "notification.action.up_voted_answer"
	// NotificationDownVotedTheAnswer down voted the answer
	NotificationDownVotedTheAnswer = "notification.action.down_voted_answer"
	// NotificationCommentQuestion comment question
	NotificationCommentQuestion = "notification.action.comment_question"
	// NotificationCommentAnswer comment answer
	NotificationCommentAnswer = "notification.action.comment_answer"
	// NotificationUpVotedTheComment up voted the comment
	NotificationUpVotedTheComment = "notification.action.up_voted_comment"
	// NotificationReplyToYou reply to you
	NotificationReplyToYou = "notification.action.reply_to_you"
	// NotificationMentionYou mention you
	NotificationMentionYou = "notification.action.mention_you"
	// NotificationYourQuestionIsClosed your question is closed
	NotificationYourQuestionIsClosed = "notification.action.your_question_is_closed"
	// NotificationYourQuestionWasDeleted your question was deleted
	NotificationYourQuestionWasDeleted = "notification.action.your_question_was_deleted"
	// NotificationYourAnswerWasDeleted your answer was deleted
	NotificationYourAnswerWasDeleted = "notification.action.your_answer_was_deleted"
	// NotificationYourCommentWasDeleted your comment was deleted
	NotificationYourCommentWasDeleted = "notification.action.your_comment_was_deleted"
	// NotificationInvitedYouToAnswer invited you to answer
	NotificationInvitedYouToAnswer = "notification.action.invited_you_to_answer"
	// NotificationEarnedBadge earned badge
	NotificationEarnedBadge = "notification.action.earned_badge"
	// NotificationAdminMessage a message written by an admin / moderator
	NotificationAdminMessage = "notification.action.admin_message"
)

type NotificationChannelKey string
type NotificationSource string

const (
	InboxSource                          NotificationSource = "inbox"
	AllNewQuestionSource                 NotificationSource = "all_new_question"
	AllNewQuestionForFollowingTagsSource NotificationSource = "all_new_question_for_following_tags"
)

const (
	EmailChannel NotificationChannelKey = "email"
)

const (
	NotificationTypeInbox            = "inbox"
	NotificationTypeAchievement      = "achievement"
	NotificationTypeBadgeAchievement = "badge"
)

var (
	NotificationMsgTypeMapping = map[string]int{
		NotificationUpdateQuestion:         1,
		NotificationAnswerTheQuestion:      1,
		NotificationUpVotedTheQuestion:     2,
		NotificationDownVotedTheQuestion:   2,
		NotificationUpdateAnswer:           1,
		NotificationAcceptAnswer:           1,
		NotificationUpVotedTheAnswer:       2,
		NotificationDownVotedTheAnswer:     2,
		NotificationCommentQuestion:        1,
		NotificationCommentAnswer:          1,
		NotificationUpVotedTheComment:      2,
		NotificationReplyToYou:             1,
		NotificationMentionYou:             1,
		NotificationYourQuestionIsClosed:   1,
		NotificationYourQuestionWasDeleted: 1,
		NotificationYourAnswerWasDeleted:   1,
		NotificationYourCommentWasDeleted:  1,
		NotificationInvitedYouToAnswer:     3,
		NotificationAdminMessage:           4,
	}
)

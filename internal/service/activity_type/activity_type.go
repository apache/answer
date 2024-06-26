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

package activity_type

const (
	QuestionVoteUp    = "question.vote_up"
	QuestionVoteDown  = "question.vote_down"
	QuestionVotedUp   = "question.voted_up"
	QuestionVotedDown = "question.voted_down"
	AnswerVoteUp      = "answer.vote_up"
	AnswerVoteDown    = "answer.vote_down"
	AnswerVotedUp     = "answer.voted_up"
	AnswerVotedDown   = "answer.voted_down"
	AnswerAccepted    = "answer.accepted"
	AnswerAccept      = "answer.accept"
	CommentVoteUp     = "comment.vote_up"
	EditAccepted      = "edit.accepted"
)

var (
	ActivityTypeList = []string{
		QuestionVoteUp,
		QuestionVoteDown,
		QuestionVotedUp,
		QuestionVotedDown,
		AnswerVoteUp,
		AnswerVoteDown,
		AnswerVotedUp,
		AnswerVotedDown,
		AnswerAccepted,
		AnswerAccept,
		CommentVoteUp,
	}
	VoteActivityTypeList = []string{
		QuestionVoteUp,
		QuestionVoteDown,
		QuestionVotedUp,
		QuestionVotedDown,
		AnswerVoteUp,
		AnswerVoteDown,
		AnswerVotedUp,
		AnswerVotedDown,
		CommentVoteUp,
	}
	ActivityTypeFlagMapping = map[string]string{
		QuestionVoteUp:    "action_activity_type.upvote",
		QuestionVoteDown:  "action_activity_type.downvote",
		QuestionVotedUp:   "action_activity_type.upvoted",
		QuestionVotedDown: "action_activity_type.downvoted",
		AnswerVoteUp:      "action_activity_type.upvote",
		AnswerVoteDown:    "action_activity_type.downvote",
		AnswerVotedUp:     "action_activity_type.upvoted",
		AnswerVotedDown:   "action_activity_type.downvoted",
		AnswerAccepted:    "action_activity_type.accepted",
		AnswerAccept:      "action_activity_type.accept",
		CommentVoteUp:     "action_activity_type.upvote",
		EditAccepted:      "action_activity_type.edit",
	}
)

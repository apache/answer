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

package display

import (
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
)

// QuestionURL get question url
func QuestionURL(permalink int, siteUrl, questionID, title string) string {
	u := siteUrl + "/questions"
	if permalink == constant.PermalinkQuestionIDAndTitle || permalink == constant.PermalinkQuestionID {
		questionID = uid.DeShortID(questionID)
	} else {
		questionID = uid.EnShortID(questionID)
	}
	u += "/" + questionID
	if permalink == constant.PermalinkQuestionIDAndTitle || permalink == constant.PermalinkQuestionIDAndTitleByShortID {
		u += "/" + htmltext.UrlTitle(title)
	}
	return u
}

// AnswerURL get answer url
func AnswerURL(permalink int, siteUrl, questionID, title, answerID string) string {
	if permalink == constant.PermalinkQuestionIDAndTitle ||
		permalink == constant.PermalinkQuestionID {
		answerID = uid.DeShortID(answerID)
	} else {
		answerID = uid.EnShortID(answerID)
	}
	return QuestionURL(permalink, siteUrl, questionID, title) + "/" + answerID
}

// CommentURL get comment url
func CommentURL(permalink int, siteUrl, questionID, title, answerID, commentID string) string {
	if len(answerID) > 0 {
		return AnswerURL(permalink, siteUrl, questionID, answerID, title) + "?commentId=" + commentID
	}
	return QuestionURL(permalink, siteUrl, questionID, title) + "?commentId=" + commentID
}

// UserURL get user url
func UserURL(siteUrl, username string) string {
	return siteUrl + "/users/" + username
}

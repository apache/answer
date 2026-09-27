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
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/stretchr/testify/assert"
)

const (
	testSiteURL    = "https://example.com"
	testQuestionID = "10010000000000001"
	testTitle      = "How to install Answer"
	testAnswerID   = "10020000000000002"
	testCommentID  = "10030000000000003"
)

// CommentURL passed title and answerID to AnswerURL in the wrong order, so a
// comment on an answer linked to a URL built from the title where the answer id
// belongs. Under the short id permalinks the raw title, spaces and all, ended up
// in the path.
func TestCommentURLOnAnswer(t *testing.T) {
	cases := []struct {
		name      string
		permalink int
		want      string
	}{
		{
			name:      "question id and title",
			permalink: constant.PermalinkQuestionIDAndTitle,
			want:      "https://example.com/questions/10010000000000001/how-to-install-answer/10020000000000002?commentId=10030000000000003",
		},
		{
			name:      "question id",
			permalink: constant.PermalinkQuestionID,
			want:      "https://example.com/questions/10010000000000001/10020000000000002?commentId=10030000000000003",
		},
		{
			name:      "question id and title by short id",
			permalink: constant.PermalinkQuestionIDAndTitleByShortID,
			want:      "https://example.com/questions/D1D1/how-to-install-answer/E1E1?commentId=10030000000000003",
		},
		{
			name:      "question id by short id",
			permalink: constant.PermalinkQuestionIDByShortID,
			want:      "https://example.com/questions/D1D1/E1E1?commentId=10030000000000003",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CommentURL(tc.permalink, testSiteURL, testQuestionID, testTitle, testAnswerID, testCommentID)
			assert.Equal(t, tc.want, got)
			assert.Equal(t,
				AnswerURL(tc.permalink, testSiteURL, testQuestionID, testTitle, testAnswerID)+"?commentId="+testCommentID,
				got,
				"a comment on an answer should be the answer URL plus the comment query")
		})
	}
}

// A comment on the question itself takes the other branch, which was already
// correct. Pin it so the fix above stays scoped to the answer branch.
func TestCommentURLOnQuestion(t *testing.T) {
	cases := []struct {
		name      string
		permalink int
		want      string
	}{
		{
			name:      "question id and title",
			permalink: constant.PermalinkQuestionIDAndTitle,
			want:      "https://example.com/questions/10010000000000001/how-to-install-answer?commentId=10030000000000003",
		},
		{
			name:      "question id",
			permalink: constant.PermalinkQuestionID,
			want:      "https://example.com/questions/10010000000000001?commentId=10030000000000003",
		},
		{
			name:      "question id and title by short id",
			permalink: constant.PermalinkQuestionIDAndTitleByShortID,
			want:      "https://example.com/questions/D1D1/how-to-install-answer?commentId=10030000000000003",
		},
		{
			name:      "question id by short id",
			permalink: constant.PermalinkQuestionIDByShortID,
			want:      "https://example.com/questions/D1D1?commentId=10030000000000003",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CommentURL(tc.permalink, testSiteURL, testQuestionID, testTitle, "", testCommentID)
			assert.Equal(t, tc.want, got)
		})
	}
}

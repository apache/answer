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

package checker_test

import (
	"testing"

	"github.com/apache/answer/pkg/checker"
	"github.com/stretchr/testify/assert"
)

func TestGetQuestionLink(t *testing.T) {
	// Step 1: Test empty content
	t.Run("Empty content", func(t *testing.T) {
		links := checker.GetQuestionLink("")
		assert.Empty(t, links)
	})

	// Step 2: Test content without link or ID
	t.Run("Content without link or ID", func(t *testing.T) {
		links := checker.GetQuestionLink("This is a random text")
		assert.Empty(t, links)
	})

	// Step 3: Test content with valid question link
	t.Run("Valid question link", func(t *testing.T) {
		links := checker.GetQuestionLink("Check this question: https://example.com/questions/10010000000000060")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
		}, links)
	})

	// Step 4: Test content with valid question and answer link
	t.Run("Valid question and answer link", func(t *testing.T) {
		links := checker.GetQuestionLink("Check this answer: https://example.com/questions/10010000000000060/10020000000000060?from=copy")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "10020000000000060",
			},
		}, links)
	})

	// Step 5: Test content with #questionID
	t.Run("Content with #questionID", func(t *testing.T) {
		links := checker.GetQuestionLink("This is question #10010000000000060")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeID,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
		}, links)
	})

	// Step 6: Test content with #answerID
	t.Run("Content with #answerID", func(t *testing.T) {
		links := checker.GetQuestionLink("This is answer #10020000000000060")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeID,
				QuestionID: "",
				AnswerID:   "10020000000000060",
			},
		}, links)
	})

	// Step 7: Test invalid question ID
	t.Run("Invalid question ID", func(t *testing.T) {
		links := checker.GetQuestionLink("https://example.com/questions/invalid")
		assert.Empty(t, links)
	})

	// Step 8: Test invalid answer ID
	t.Run("Invalid answer ID", func(t *testing.T) {
		links := checker.GetQuestionLink("https://example.com/questions/10010000000000060/invalid")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
		}, links)
	})

	// Step 9: Test content with multiple links and IDs
	t.Run("Multiple links and IDs", func(t *testing.T) {
		content := "Question #10010000000000060 and https://example.com/questions/10010000000000060/10020000000000061 and https://example.com/questions/10010000000000065/10020000000000066 and another #10020000000000066"
		links := checker.GetQuestionLink(content)
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeID,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "10020000000000061",
			},
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000065",
				AnswerID:   "10020000000000066",
			},
		}, links)
	})

	// Step 11: Test URL with www prefix
	t.Run("URL with www prefix", func(t *testing.T) {
		links := checker.GetQuestionLink("Check this question: https://www.example.com/questions/10010000000000060")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
		}, links)
	})

	// Step 12: Test URL without protocol
	t.Run("URL without protocol", func(t *testing.T) {
		links := checker.GetQuestionLink("Check this question: example.com/questions/10010000000000060")
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000060",
				AnswerID:   "",
			},
		}, links)
	})

	// Step 14: Test error id
	t.Run("Error id", func(t *testing.T) {
		links := checker.GetQuestionLink("https://example.com/questions/10110000000000060")
		assert.Empty(t, links)
	})

	// step 15: SEO options
	t.Run("SEO options", func(t *testing.T) {
		content := `
		URL1: http://localhost:3000/questions/D1I2
		URL2: http://localhost:3000/questions/D1I2/hello
		URL3: http://localhost:3000/questions/10010000000000068
		URL4: http://localhost:3000/questions/10010000000000068/hello
		ERROR URL: http://localhost:3000/questions/AAAA/BBBB
		`
		links := checker.GetQuestionLink(content)
		assert.Equal(t, []checker.QuestionLink{
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "D1I2",
				AnswerID:   "",
			},
			{
				LinkType:   checker.QuestionLinkTypeURL,
				QuestionID: "10010000000000068",
				AnswerID:   "",
			},
		}, links)
	})
}

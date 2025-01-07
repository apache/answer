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

package checker

import (
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/pkg/obj"
	"github.com/apache/answer/pkg/uid"
)

const (
	QuestionLinkTypeURL = 1
	QuestionLinkTypeID  = 2
)

type QuestionLink struct {
	LinkType   int
	QuestionID string
	AnswerID   string
}

func GetQuestionLink(content string) []QuestionLink {
	uniqueIDs := make(map[string]struct{})
	var questionLinks []QuestionLink

	// use two pointer to find the link
	left, right := 0, 0
	for right < len(content) {
		// find "/questions/" or "#"
		if right+11 < len(content) && content[right:right+11] == "/questions/" {
			left = right
			right += 11
			processURL(content, &left, &right, uniqueIDs, &questionLinks)
		} else if content[right] == '#' {
			left = right + 1
			right = left
			processID(content, &left, &right, uniqueIDs, &questionLinks)
		} else {
			right++
		}
	}

	return questionLinks
}

func processURL(content string, left, right *int, uniqueIDs map[string]struct{}, questionLinks *[]QuestionLink) {
	for *right < len(content) && (isDigit(content[*right]) || isLetter(content[*right])) {
		*right++
	}
	questionID := content[*left+len("/questions/") : *right]

	answerID := ""
	if *right < len(content) && content[*right] == '/' {
		*left = *right + 1
		*right = *left
		for *right < len(content) && (isDigit(content[*right]) || isLetter(content[*right])) {
			*right++
		}
		answerID = content[*left:*right]
	}

	addUniqueID(questionID, answerID, QuestionLinkTypeURL, uniqueIDs, questionLinks)
}

func processID(content string, left, right *int, uniqueIDs map[string]struct{}, questionLinks *[]QuestionLink) {
	for *right < len(content) && (isDigit(content[*right]) || isLetter(content[*right])) {
		*right++
	}
	id := content[*left:*right]
	addUniqueID(id, "", QuestionLinkTypeID, uniqueIDs, questionLinks)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func addUniqueID(questionID, answerID string, linkType int, uniqueIDs map[string]struct{}, questionLinks *[]QuestionLink) {
	isAdd := false
	if answerID != "" {
		objectType, err := obj.GetObjectTypeStrByObjectID(uid.DeShortID(answerID))
		if err != nil {
			answerID = ""
		}

		if objectType == constant.AnswerObjectType {
			if _, ok := uniqueIDs[answerID]; !ok {
				uniqueIDs[answerID] = struct{}{}
				isAdd = true
			}
		}
	}

	if objectType, err := obj.GetObjectTypeStrByObjectID(uid.DeShortID(questionID)); err == nil {
		if _, ok := uniqueIDs[questionID]; !ok {
			uniqueIDs[questionID] = struct{}{}
			isAdd = true
			if objectType == constant.AnswerObjectType {
				answerID = questionID
				questionID = ""
			}
		}
	}

	if isAdd {
		*questionLinks = append(*questionLinks, QuestionLink{
			LinkType:   linkType,
			QuestionID: questionID,
			AnswerID:   answerID,
		})
	}
}

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

package content

import (
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

func TestCheckQuestionVisibility(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		show   int
		userID string
		viewer string
		per    schema.QuestionPermission
		allow  bool
	}{
		{"public question", entity.QuestionStatusAvailable, entity.QuestionShow, "author", "", schema.QuestionPermission{}, true},
		{"pending question anonymous", entity.QuestionStatusPending, entity.QuestionShow, "author", "", schema.QuestionPermission{}, false},
		{"pending question author", entity.QuestionStatusPending, entity.QuestionShow, "author", "author", schema.QuestionPermission{}, true},
		{"pending question reviewer", entity.QuestionStatusPending, entity.QuestionShow, "author", "reviewer", schema.QuestionPermission{CanReopen: true}, true},
		{"deleted question anonymous", entity.QuestionStatusDeleted, entity.QuestionShow, "author", "", schema.QuestionPermission{}, false},
		{"hidden question anonymous", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "", schema.QuestionPermission{}, false},
		{"hidden question author", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "author", schema.QuestionPermission{}, true},
		{"hidden question moderator", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "moderator", schema.QuestionPermission{IsAdminModerator: true}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			question := &schema.QuestionInfoResp{
				Status: testCase.status,
				Show:   testCase.show,
				UserID: testCase.userID,
			}
			err := checkQuestionVisibility(question, testCase.viewer, testCase.per)
			if testCase.allow && err != nil {
				t.Fatalf("visibility unexpectedly denied: %v", err)
			}
			if !testCase.allow && err == nil {
				t.Fatal("visibility unexpectedly allowed")
			}
		})
	}
}

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

package controller

import (
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

func TestMCPQuestionIsPublic(t *testing.T) {
	testCases := []struct {
		name     string
		question *schema.QuestionInfoResp
		want     bool
	}{
		{"nil", nil, false},
		{"available", &schema.QuestionInfoResp{Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow}, true},
		{"closed", &schema.QuestionInfoResp{Status: entity.QuestionStatusClosed, Show: entity.QuestionShow}, true},
		{"hidden", &schema.QuestionInfoResp{Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}, false},
		{"deleted", &schema.QuestionInfoResp{Status: entity.QuestionStatusDeleted, Show: entity.QuestionShow}, false},
		{"pending", &schema.QuestionInfoResp{Status: entity.QuestionStatusPending, Show: entity.QuestionShow}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := mcpQuestionIsPublic(testCase.question); got != testCase.want {
				t.Fatalf("mcpQuestionIsPublic() = %v, want %v", got, testCase.want)
			}
		})
	}
}

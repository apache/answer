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

	"github.com/apache/answer/internal/schema"
)

func TestCanOperateQuestionUsesMatchingPermission(t *testing.T) {
	testCases := []struct {
		name      string
		operation string
		canList   []bool
		want      bool
	}{
		{"pin", schema.QuestionOperationPin, []bool{true, false, false, false}, true},
		{"unpin", schema.QuestionOperationUnPin, []bool{false, true, false, false}, true},
		{"hide", schema.QuestionOperationHide, []bool{false, false, true, false}, true},
		{"show", schema.QuestionOperationShow, []bool{false, false, false, true}, true},
		{"unpin does not authorize hide", schema.QuestionOperationHide, []bool{false, true, false, false}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := canOperateQuestion(testCase.operation, testCase.canList); got != testCase.want {
				t.Fatalf("canOperateQuestion(%q, %v) = %v, want %v", testCase.operation, testCase.canList, got, testCase.want)
			}
		})
	}
}

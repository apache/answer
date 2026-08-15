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

package rank

import (
	"testing"

	"github.com/apache/answer/internal/service/permission"
	"github.com/stretchr/testify/require"
)

func TestIsMemberPermission(t *testing.T) {
	tests := map[string]bool{
		permission.QuestionAdd:                 true,
		permission.QuestionVoteUp:              true,
		permission.QuestionVoteDown:            true,
		permission.AnswerAdd:                   true,
		permission.AnswerVoteUp:                true,
		permission.AnswerVoteDown:              true,
		permission.AnswerInviteSomeoneToAnswer: true,
		permission.CommentAdd:                  true,
		permission.CommentVoteUp:               true,
		permission.CommentVoteDown:             true,
		permission.ReportAdd:                   true,
		permission.TagAdd:                      true,
		permission.VoteDetail:                  true,
		permission.QuestionEdit:                false,
		permission.QuestionDelete:              false,
		permission.QuestionAudit:               false,
		permission.QuestionPin:                 false,
		permission.TagEdit:                     false,
	}

	for action, expected := range tests {
		t.Run(action, func(t *testing.T) {
			require.Equal(t, expected, isMemberPermission(action))
		})
	}
}

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

	"github.com/apache/answer/pkg/uid"
)

// TestSameObjectID guards the AcceptAnswer ownership check (issue #1541).
//
// When short links are enabled, answerRepo.GetByID re-encodes the answer's
// QuestionID to its short form while the controller de-shorts req.QuestionID
// to its long form. The two encodings of the same question must be treated as
// equal, or accepting any answer fails with "Answer do not found".
func TestSameObjectID(t *testing.T) {
	const longQID = "10010000000000001"
	shortQID := uid.EnShortID(longQID) // e.g. "D1D1"
	if shortQID == "" || shortQID == longQID {
		t.Fatalf("precondition failed: EnShortID(%q)=%q, want a distinct short id", longQID, shortQID)
	}
	otherLongQID := "10010000000000002"

	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "short answer-side id vs long request-side id, same question (the bug)",
			a:    shortQID,
			b:    longQID,
			want: true,
		},
		{
			name: "both long, same question (default permalink)",
			a:    longQID,
			b:    longQID,
			want: true,
		},
		{
			name: "both short, same question",
			a:    shortQID,
			b:    shortQID,
			want: true,
		},
		{
			name: "different questions must stay rejected (privilege-escalation guard)",
			a:    shortQID,
			b:    otherLongQID,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameObjectID(tt.a, tt.b); got != tt.want {
				t.Errorf("sameObjectID(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

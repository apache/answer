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

import "testing"

func TestEmailInAllowEmailDomain(t *testing.T) {
	allowed := []string{"hainanu.edu.cn", "@alumni.hainanu.edu.cn"}
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "student email", email: "student@hainanu.edu.cn", want: true},
		{name: "alumni email", email: "alumni@alumni.hainanu.edu.cn", want: true},
		{name: "case insensitive", email: "student@HAINANU.EDU.CN", want: true},
		{name: "surrounding spaces", email: " student@hainanu.edu.cn ", want: true},
		{name: "subdomain is not exact", email: "student@mail.hainanu.edu.cn", want: false},
		{name: "prefixed domain", email: "student@evil-hainanu.edu.cn", want: false},
		{name: "suffixed domain", email: "student@hainanu.edu.cn.evil.com", want: false},
		{name: "multiple at signs", email: "student@evil@hainanu.edu.cn", want: false},
		{name: "missing local part", email: "@hainanu.edu.cn", want: false},
		{name: "missing domain", email: "student@", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := EmailInAllowEmailDomain(test.email, allowed); got != test.want {
				t.Fatalf("EmailInAllowEmailDomain(%q) = %v, want %v", test.email, got, test.want)
			}
		})
	}
}

func TestEmailInAllowEmailDomainAllowsAllWhenUnconfigured(t *testing.T) {
	if !EmailInAllowEmailDomain("any@example.com", nil) {
		t.Fatal("an empty allow list should preserve the existing allow-all behavior")
	}
}

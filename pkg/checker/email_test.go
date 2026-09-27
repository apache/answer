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
	tests := []struct {
		name              string
		email             string
		allowEmailDomains []string
		want              bool
	}{
		{name: "allow when no domains are configured", email: "user@example.com", want: true},
		{name: "allow legacy domain configuration", email: "user@example.com", allowEmailDomains: []string{"example.com"}, want: true},
		{name: "allow domain configuration with at sign", email: "user@example.com", allowEmailDomains: []string{"@example.com"}, want: true},
		{name: "reject a domain with an allowed suffix", email: "attacker@notexample.com", allowEmailDomains: []string{"example.com"}, want: false},
		{name: "reject a domain with an allowed suffix and at sign", email: "attacker@notexample.com", allowEmailDomains: []string{"@example.com"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmailInAllowEmailDomain(tt.email, tt.allowEmailDomains); got != tt.want {
				t.Errorf("EmailInAllowEmailDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

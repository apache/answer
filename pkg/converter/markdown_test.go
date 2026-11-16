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

package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderLinkIsUrl(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"absolute http URL", "http://example.com/path?q=1#f", true},
		{"absolute https URL", "https://example.com", true},
		{"ftp URL", "ftp://example.com/file", true},
		{"uppercase scheme and host", "HTTP://EXAMPLE.COM", false},
		{"bare domain", "example.com", true},
		{"bare domain with path", "example.com/questions/123", true},
		{"www subdomain", "www.example.com", true},
		{"bare IP", "10.0.0.1", true},
		{"IP with port and path", "10.0.0.1:8080/a", true},
		{"host with port", "localhost:8080", true},
		{"domain with port and path", "example.com:8080/x", true},
		{"IPv6 with port", "[::1]:8080", true},
		{"userinfo", "user:pass@example.com", true},
		{"mailto", "mailto:a@b.com", true},
		{"email-like destination", "a@b.co", true},
		{"userinfo without scheme", "user@h.co", true},
		{"trailing dot FQDN", "example.com.", true},
		{"empty", "", false},
		{"single word", "foo", false},
		{"path segment no dot", "questions/123", false},
		{"absolute path", "/questions/123", true},
		{"scheme-less authority path", "//cdn.example.com/x", true},
		{"anchor", "#section", false},
		{"leading dot", ".hidden", false},
		{"javascript scheme", "javascript:alert(1)", false},
		{"tel scheme", "tel:+1234", false},
		{"host with leading dot", "http://.example.com", false},
		{"trailing colon", "example.com:", false},
		{"single label with scheme", "http://localhost", true},
		{"single label no scheme no port", "localhost", false},
		{"not a url", "not a url", false},
		{"whitespace in path", "h.co/p q", false},
		{"label with leading hyphen", "-ex.com", false},
		{"label with trailing hyphen", "ex-.com", false},
		{"invalid IPv4 quad", "999.1.1.1", false},
		{"IPv4 with leading zeros", "01.2.3.4", false},
		{"three letter domain", "a.b", false},
	}
	r := &DangerousHTMLRenderer{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, r.renderLinkIsUrl(tc.in))
		})
	}
}

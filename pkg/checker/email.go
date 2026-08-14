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

import "strings"

func EmailInAllowEmailDomain(email string, allowEmailDomains []string) bool {
	if len(allowEmailDomains) == 0 {
		return true
	}

	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 0 || at != strings.Index(email, "@") || at == len(email)-1 {
		return false
	}
	emailDomain := email[at+1:]

	for _, domain := range allowEmailDomains {
		domain = strings.TrimSpace(strings.TrimPrefix(domain, "@"))
		if domain != "" && strings.EqualFold(emailDomain, domain) {
			return true
		}
	}

	return false
}

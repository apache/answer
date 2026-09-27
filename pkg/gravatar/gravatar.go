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

package gravatar

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

// GetAvatarURL get avatar url from gravatar by email
func GetAvatarURL(baseURL, email string) string {
	hasher := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	hash := hex.EncodeToString(hasher[:])
	return baseURL + hash
}

// Resize resize avatar by pixel
func Resize(originalAvatarURL string, sizePixel int) (resizedAvatarURL string) {
	if len(originalAvatarURL) == 0 {
		return
	}
	originalURL, err := url.Parse(originalAvatarURL)
	if err != nil {
		return originalAvatarURL
	}
	query := originalURL.Query()
	query.Set("s", fmt.Sprintf("%d", sizePixel))
	originalURL.RawQuery = query.Encode()
	return originalURL.String()
}

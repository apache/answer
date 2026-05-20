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

package constant

const (
	DefaultGravatarBaseURL = "https://www.gravatar.com/avatar/"
	DefaultAvatar          = "system"
	AvatarTypeDefault      = "default"
	AvatarTypeGravatar     = "gravatar"
	AvatarTypeCustom       = "custom"
)

const (
	// PermalinkQuestionIDAndTitle /questions/10010000000000001/post-title
	PermalinkQuestionIDAndTitle = iota + 1
	// PermalinkQuestionID /questions/10010000000000001
	PermalinkQuestionID
	// PermalinkQuestionIDAndTitleByShortID /questions/11/post-title
	PermalinkQuestionIDAndTitleByShortID
	// PermalinkQuestionIDByShortID /questions/11
	PermalinkQuestionIDByShortID
)

const (
	ColorSchemeDefault = "default"
	ColorSchemeLight   = "light"
	ColorSchemeDark    = "dark"
	ColorSchemeSystem  = "system"

	ThemeLayoutFullWidth  = "Full-width"
	ThemeLayoutFixedWidth = "Fixed-width"
)

const (
	EmailConfigKey = "email.config"
)

const (
	DefaultMaxImageMegapixel = 40 * 1000 * 1000
	DefaultMaxImageSize      = 4 * 1024 * 1024
	DefaultMaxAttachmentSize = 8 * 1024 * 1024
)

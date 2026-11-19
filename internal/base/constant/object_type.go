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
	QuestionObjectType     = "question"
	AnswerObjectType       = "answer"
	TagObjectType          = "tag"
	UserObjectType         = "user"
	CollectionObjectType   = "collection"
	CommentObjectType      = "comment"
	ReportObjectType       = "report"
	BadgeObjectType        = "badge"
	BadgeAwardObjectType   = "badge_award"
	AdminMessageObjectType = "admin_message" // notification object type of admin messages
)

var (
	ObjectTypeStrMapping = map[string]int{
		QuestionObjectType:     1,
		AnswerObjectType:       2,
		TagObjectType:          3,
		UserObjectType:         4,
		CollectionObjectType:   6,
		CommentObjectType:      7,
		ReportObjectType:       8,
		BadgeObjectType:        9,
		BadgeAwardObjectType:   10,
		AdminMessageObjectType: 11,
	}

	ObjectTypeNumberMapping = map[int]string{
		1:  QuestionObjectType,
		2:  AnswerObjectType,
		3:  TagObjectType,
		4:  UserObjectType,
		6:  CollectionObjectType,
		7:  CommentObjectType,
		8:  ReportObjectType,
		9:  BadgeObjectType,
		10: BadgeAwardObjectType,
	}
)

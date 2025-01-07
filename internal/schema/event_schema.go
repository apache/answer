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

package schema

import (
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/pkg/uid"
)

// EventMsg event message
type EventMsg struct {
	EventType constant.EventType
	UserID    string

	TriggerObjectID string

	QuestionID     string
	QuestionUserID string

	AnswerID     string
	AnswerUserID string

	CommentID     string
	CommentUserID string

	ExtraInfo map[string]string
}

// NewEvent create a new event
func NewEvent(e constant.EventType, userID string) *EventMsg {
	return &EventMsg{
		UserID:    userID,
		EventType: e,
		ExtraInfo: make(map[string]string),
	}
}

// QID get question id
func (e *EventMsg) QID(questionID, userID string) *EventMsg {
	if len(questionID) > 0 {
		e.QuestionID = uid.DeShortID(questionID)
	}
	e.QuestionUserID = userID
	return e
}

// AID get answer id
func (e *EventMsg) AID(answerID, userID string) *EventMsg {
	if len(answerID) > 0 {
		e.AnswerID = uid.DeShortID(answerID)
	}
	e.AnswerUserID = userID
	return e
}

// CID get comment id
func (e *EventMsg) CID(comment, userID string) *EventMsg {
	e.CommentID = comment
	e.CommentUserID = userID
	return e
}

// TID get trigger object id
func (e *EventMsg) TID(triggerObjectID string) *EventMsg {
	if len(triggerObjectID) > 0 {
		e.TriggerObjectID = uid.DeShortID(triggerObjectID)
	}
	return e
}

// AddExtra add extra info
func (e *EventMsg) AddExtra(key, value string) *EventMsg {
	e.ExtraInfo[key] = value
	return e
}

// GetExtra get extra info
func (e *EventMsg) GetExtra(key string) string {
	if v, ok := e.ExtraInfo[key]; ok {
		return v
	}
	return ""
}

// GetObjectID get object id
func (e *EventMsg) GetObjectID() string {
	if len(e.TriggerObjectID) > 0 {
		return e.TriggerObjectID
	}
	if len(e.CommentID) > 0 {
		return e.CommentID
	}
	if len(e.AnswerID) > 0 {
		return e.AnswerID
	}
	return e.QuestionID
}

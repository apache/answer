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
	"github.com/apache/answer/internal/base/translator"
	"github.com/segmentfault/pacman/i18n"
)

type ReasonItem struct {
	ReasonKey   string `json:"reason_key"`
	ReasonType  int    `json:"reason_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ContentType string `json:"content_type"`
	Placeholder string `json:"placeholder"`
}

type ReasonReq struct {
	// ObjectType
	ObjectType string `validate:"required" form:"object_type" json:"object_type"`
	// Action
	Action string `validate:"required" form:"action" json:"action"`
}

func (r *ReasonItem) Translate(keyPrefix string, lang i18n.Language) {
	trField := func(fieldName, fieldData string) string {
		// If fieldData is empty, means no need to translate
		if len(fieldData) == 0 {
			return fieldData
		}
		key := keyPrefix + "." + fieldName
		fieldTr := translator.Tr(lang, key)
		if fieldTr != key {
			// If i18n key exists, return i18n value
			return fieldTr
		}
		// If i18n key not exists, return fieldData original value
		return fieldData
	}

	r.ReasonKey = keyPrefix
	r.Name = trField("name", r.Name)
	r.Description = trField("desc", r.Description)
	r.Placeholder = trField("placeholder", r.Placeholder)
}

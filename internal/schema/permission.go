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
	"strings"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/segmentfault/pacman/i18n"
)

// PermissionTrTplData template data as for translate permission message
type PermissionTrTplData struct {
	Rank int
}

// PermissionMemberAction permission member action
type PermissionMemberAction struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

// GetPermissionReq get permission request
type GetPermissionReq struct {
	Action  string   `form:"action"`
	Actions []string `validate:"omitempty" form:"actions"`
}

func (r *GetPermissionReq) Check() (errField []*validator.FormErrorField, err error) {
	if len(r.Action) > 0 {
		r.Actions = strings.Split(r.Action, ",")
	}
	return nil, nil
}

// GetPermissionResp get permission response
type GetPermissionResp struct {
	HasPermission bool `json:"has_permission"`
	// only not allow, will return this tip
	NoPermissionTip string `json:"no_permission_tip"`
}

func (r *GetPermissionResp) TrTip(lang i18n.Language, requireRank int) {
	if r.HasPermission {
		return
	}
	if requireRank <= 0 {
		r.NoPermissionTip = translator.Tr(lang, reason.RankFailToMeetTheCondition)
	} else {
		r.NoPermissionTip = translator.TrWithData(
			lang, reason.NoEnoughRankToOperate, &PermissionTrTplData{Rank: requireRank})
	}
}

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

package handler

import (
	"github.com/apache/answer/internal/base/translator"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/i18n"
)

// RespBody response body.
type RespBody struct {
	// http code
	Code int `json:"code"`
	// reason key
	Reason string `json:"reason"`
	// response message
	Message string `json:"msg"`
	// response data
	Data interface{} `json:"data"`
}

// TrMsg translate the reason cause as a message
func (r *RespBody) TrMsg(lang i18n.Language) *RespBody {
	if len(r.Message) == 0 {
		r.Message = translator.Tr(lang, r.Reason)
	}
	return r
}

// NewRespBody new response body
func NewRespBody(code int, reason string) *RespBody {
	return &RespBody{
		Code:   code,
		Reason: reason,
	}
}

// NewRespBodyFromError new response body from error
func NewRespBodyFromError(e *errors.Error) *RespBody {
	return &RespBody{
		Code:    e.Code,
		Reason:  e.Reason,
		Message: e.Message,
	}
}

// NewRespBodyData new response body with data
func NewRespBodyData(code int, reason string, data interface{}) *RespBody {
	return &RespBody{
		Code:   code,
		Reason: reason,
		Data:   data,
	}
}

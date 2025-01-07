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

package controller

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/reason"
	"github.com/gin-gonic/gin"
)

// ReasonController answer controller
type ReasonController struct {
	reasonService *reason.ReasonService
}

// NewReasonController new controller
func NewReasonController(answerService *reason.ReasonService) *ReasonController {
	return &ReasonController{reasonService: answerService}
}

// Reasons godoc
// @Summary get reasons by object type and action
// @Description get reasons by object type and action
// @Tags reason
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param object_type query string true "object_type" Enums(question, answer, comment, user)
// @Param action query string true "action" Enums(status, close, flag, review)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/reasons [get]
// @Router /answer/admin/api/reasons [get]
func (rc *ReasonController) Reasons(ctx *gin.Context) {
	req := &schema.ReasonReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	reasons, err := rc.reasonService.GetReasons(ctx, *req)
	handler.HandleResponse(ctx, err, reasons)
}

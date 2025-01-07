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
	"github.com/apache/answer/internal/service/dashboard"
	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService dashboard.DashboardService
}

// NewDashboardController new controller
func NewDashboardController(
	dashboardService dashboard.DashboardService,
) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

// DashboardInfo godoc
// @Summary DashboardInfo
// @Description DashboardInfo
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /answer/admin/api/dashboard [get]
// @Success 200 {object} handler.RespBody
func (ac *DashboardController) DashboardInfo(ctx *gin.Context) {
	info, err := ac.dashboardService.Statistical(ctx)
	handler.HandleResponse(ctx, err, gin.H{
		"info": info,
	})
}

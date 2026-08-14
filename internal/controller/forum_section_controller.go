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
	forumsectionservice "github.com/apache/answer/internal/service/forum_section"
	"github.com/gin-gonic/gin"
)

type ForumSectionController struct {
	service *forumsectionservice.Service
}

func NewForumSectionController(service *forumsectionservice.Service) *ForumSectionController {
	return &ForumSectionController{service: service}
}

// List returns the campus section tree.
// @Summary list campus forum sections
// @Description returns parent sections and their child sections
// @Tags Forum Section
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.ForumSectionResp}
// @Router /answer/api/v1/forum/sections [get]
func (c *ForumSectionController) List(ctx *gin.Context) {
	sections, err := c.service.ListTree(ctx)
	handler.HandleResponse(ctx, err, sections)
}

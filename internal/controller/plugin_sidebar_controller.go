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
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

// SidebarController is the controller for the sidebar plugin.
type SidebarController struct{}

// NewSidebarController creates a new instance of SidebarController.
func NewSidebarController() *SidebarController {
	return &SidebarController{}
}

// GetSidebarConfig retrieves the sidebar configuration from the registered sidebar plugins.
func (uc *SidebarController) GetSidebarConfig(ctx *gin.Context) {
	resp := &plugin.SidebarConfig{}
	_ = plugin.CallSidebar(func(fn plugin.Sidebar) error {
		cfg, err := fn.GetSidebarConfig()
		if err != nil {
			return err
		}
		resp = cfg
		return nil
	})
	handler.HandleResponse(ctx, nil, resp)
}

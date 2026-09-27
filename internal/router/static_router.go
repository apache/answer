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

package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/apache/answer/pkg/dir"
	"github.com/gin-gonic/gin"
)

// StaticRouter static api router
type StaticRouter struct {
	serviceConfig *service_config.ServiceConfig
}

// NewStaticRouter new static api router
func NewStaticRouter(serviceConfig *service_config.ServiceConfig) *StaticRouter {
	return &StaticRouter{
		serviceConfig: serviceConfig,
	}
}

// RegisterStaticRouter register static api router
func (a *StaticRouter) RegisterStaticRouter(r *gin.RouterGroup) {
	r.Static("/uploads/"+constant.AvatarSubPath, filepath.Join(a.serviceConfig.UploadPath, constant.AvatarSubPath))
	r.Static("/uploads/"+constant.AvatarThumbSubPath, filepath.Join(a.serviceConfig.UploadPath, constant.AvatarThumbSubPath))
	r.Static("/uploads/"+constant.PostSubPath, filepath.Join(a.serviceConfig.UploadPath, constant.PostSubPath))
	r.Static("/uploads/"+constant.BrandingSubPath, filepath.Join(a.serviceConfig.UploadPath, constant.BrandingSubPath))
	r.GET("/uploads/"+constant.FilesPostSubPath+"/*filepath", func(c *gin.Context) {
		// The filepath such as hash/123.pdf
		filePath := c.Param("filepath")
		// The original filename is 123.pdf
		originalFilename := filepath.Base(filePath)
		fileLocalPath, ok := attachmentFileLocalPath(a.serviceConfig.UploadPath, filePath, originalFilename)
		if !ok {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		// If the file is not exist, return 404
		if !dir.CheckFileExist(fileLocalPath) {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		c.FileAttachment(fileLocalPath, originalFilename)
	})
}

func attachmentFileLocalPath(uploadPath, requestPath, originalFilename string) (string, bool) {
	realFilename := strings.TrimSuffix(requestPath, "/"+originalFilename) + filepath.Ext(originalFilename)
	attachmentRoot := filepath.Join(uploadPath, constant.FilesPostSubPath)
	fileLocalPath := filepath.Join(attachmentRoot, realFilename)
	relPath, err := filepath.Rel(attachmentRoot, fileLocalPath)
	if err != nil || filepath.IsAbs(relPath) || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", false
	}
	return fileLocalPath, true
}

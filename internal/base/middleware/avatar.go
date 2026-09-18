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

package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/apache/answer/internal/service/service_config"
	"github.com/apache/answer/internal/service/uploader"
	"github.com/apache/answer/pkg/converter"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

type AvatarMiddleware struct {
	serviceConfig   *service_config.ServiceConfig
	uploaderService uploader.UploaderService
}

// NewAvatarMiddleware new auth user middleware
func NewAvatarMiddleware(serviceConfig *service_config.ServiceConfig,
	uploaderService uploader.UploaderService,
) *AvatarMiddleware {
	return &AvatarMiddleware{
		serviceConfig:   serviceConfig,
		uploaderService: uploaderService,
	}
}

func (am *AvatarMiddleware) AvatarThumb() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uri := ctx.Request.RequestURI
		if strings.Contains(uri, "/uploads/avatar/") {
			size := converter.StringToInt(ctx.Query("s"))
			uriWithoutQuery, _ := url.Parse(uri)
			filename := filepath.Base(uriWithoutQuery.Path)
			filePath := fmt.Sprintf("%s/avatar/%s", am.serviceConfig.UploadPath, filename)
			var err error
			if size != 0 {
				filePath, err = am.uploaderService.AvatarThumbFile(ctx, filename, size)
				if err != nil {
					log.Error(err)
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}
			}
			avatarFile, err := os.ReadFile(filePath)
			if err != nil {
				log.Error(err)
				ctx.Abort()
				return
			}
			ctx.Header("Content-Type", contentTypeByExt(path.Ext(filePath)))
			_, err = ctx.Writer.Write(avatarFile)
			if err != nil {
				log.Error(err)
			}
			ctx.Abort()
			return
		} else {
			urlInfo, err := url.Parse(uri)
			if err != nil {
				ctx.Next()
				return
			}
			ctx.Header("Content-Type", contentTypeByExt(filepath.Ext(urlInfo.Path)))
		}
		ctx.Next()
	}
}

// contentTypeByExt returns the MIME type for a file extension (with leading dot).
// It prefers the registered MIME table (e.g. ".svg" -> "image/svg+xml", which browsers
// require to render SVG in <img>) and falls back to the legacy "image/<ext>" guess.
// Paths without an extension get application/octet-stream rather than an invalid "image/".
func contentTypeByExt(ext string) string {
	ext = strings.ToLower(ext)
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	if name := strings.TrimPrefix(ext, "."); name != "" {
		return fmt.Sprintf("image/%s", name)
	}
	return "application/octet-stream"
}

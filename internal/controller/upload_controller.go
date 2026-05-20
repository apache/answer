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
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/uploader"
	"github.com/apache/answer/pkg/converter"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

const (
	// file is uploaded by markdown(or something else) editor
	fileFromPost = "post"
	// file is used to upload the post attachment
	fileFromPostAttachment = "post_attachment"
	// file is used to change the user's avatar
	fileFromAvatar = "avatar"
	// file is logo/icon images
	fileFromBranding = "branding"
)

// UploadController upload controller
type UploadController struct {
	uploaderService uploader.UploaderService
}

// NewUploadController new controller
func NewUploadController(uploaderService uploader.UploaderService) *UploadController {
	return &UploadController{
		uploaderService: uploaderService,
	}
}

// UploadFile upload file
// @Summary upload file
// @Description upload file
// @Tags Upload
// @Accept multipart/form-data
// @Security ApiKeyAuth
// @Param source formData string true "identify the source of the file upload" Enums(post, post_attachment, avatar, branding)
// @Param file formData file true "file"
// @Success 200 {object} handler.RespBody{data=string}
// @Router /answer/api/v1/file [post]
func (uc *UploadController) UploadFile(ctx *gin.Context) {
	var (
		url string
		err error
	)

	source := ctx.PostForm("source")
	userID := middleware.GetLoginUserIDFromContext(ctx)
	switch source {
	case fileFromAvatar:
		url, err = uc.uploaderService.UploadAvatarFile(ctx, userID)
	case fileFromPost:
		url, err = uc.uploaderService.UploadPostFile(ctx, userID)
	case fileFromBranding:
		if !middleware.GetIsAdminFromContext(ctx) {
			handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
			return
		}
		url, err = uc.uploaderService.UploadBrandingFile(ctx, userID)
	case fileFromPostAttachment:
		url, err = uc.uploaderService.UploadPostAttachment(ctx, userID)
	default:
		handler.HandleResponse(ctx, errors.BadRequest(reason.UploadFileSourceUnsupported), nil)
		return
	}
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, err, url)
}

// PostRender render post content
// @Summary render post content
// @Description render post content
// @Tags Upload
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.PostRenderReq true "PostRenderReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/post/render [post]
func (uc *UploadController) PostRender(ctx *gin.Context) {
	req := &schema.PostRenderReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	handler.HandleResponse(ctx, nil, converter.Markdown2HTML(req.Content))
}

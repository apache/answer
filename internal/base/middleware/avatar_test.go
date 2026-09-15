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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeByExt(t *testing.T) {
	assert.Equal(t, "image/svg+xml", contentTypeByExt(".svg"))
	assert.Equal(t, "image/png", contentTypeByExt(".png"))
	assert.Equal(t, "image/jpeg", contentTypeByExt(".JPG"))
	assert.Equal(t, "image/webp", contentTypeByExt(".webp"))
	// unknown extension keeps the legacy behaviour
	assert.Equal(t, "image/unknownext", contentTypeByExt(".unknownext"))
}

func TestAvatarThumbSetsContentTypeForUploads(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use((&AvatarMiddleware{}).AvatarThumb())
	r.GET("/uploads/branding/:file", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	for path, want := range map[string]string{
		"/uploads/branding/logo.svg": "image/svg+xml",
		"/uploads/branding/logo.png": "image/png",
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.RequestURI = path // set by the real server; the middleware reads it
		r.ServeHTTP(w, req)
		assert.Equal(t, want, w.Header().Get("Content-Type"), path)
	}
}

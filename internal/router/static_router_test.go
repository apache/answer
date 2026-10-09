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
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/gin-gonic/gin"
)

func TestAttachmentFileLocalPathRejectsTraversal(t *testing.T) {
	uploadPath := t.TempDir()

	filePath, ok := attachmentFileLocalPath(uploadPath, "/hash/report.pdf", "report.pdf")
	if !ok {
		t.Fatal("valid attachment path was rejected")
	}
	want := filepath.Join(uploadPath, constant.FilesPostSubPath, "hash.pdf")
	if filePath != want {
		t.Fatalf("attachment path = %q, want %q", filePath, want)
	}

	for _, requestPath := range []string{"/../../outside/secret.txt", "/hash/../../../secret.txt"} {
		if _, ok := attachmentFileLocalPath(uploadPath, requestPath, filepath.Base(requestPath)); ok {
			t.Fatalf("traversal path %q was accepted", requestPath)
		}
	}
}

// The uploader stores attachments under a lowercased extension (Report.PDF is
// saved as hash.pdf), so the download route must look the file up the same way.
func TestAttachmentFileLocalPathLowercasesExtension(t *testing.T) {
	uploadPath := t.TempDir()

	filePath, ok := attachmentFileLocalPath(uploadPath, "/hash/Report.PDF", "Report.PDF")
	if !ok {
		t.Fatal("valid attachment path was rejected")
	}
	want := filepath.Join(uploadPath, constant.FilesPostSubPath, "hash.pdf")
	if filePath != want {
		t.Fatalf("attachment path = %q, want %q", filePath, want)
	}
}

func TestAttachmentDownloadWithUppercaseExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploadPath := t.TempDir()
	attachmentDir := filepath.Join(uploadPath, constant.FilesPostSubPath)
	if err := os.MkdirAll(attachmentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// This is the name UploadPostAttachment gives an upload called Report.PDF.
	if err := os.WriteFile(filepath.Join(attachmentDir, "hash.pdf"), []byte("%PDF-1.4"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	NewStaticRouter(&service_config.ServiceConfig{UploadPath: uploadPath}).RegisterStaticRouter(&r.RouterGroup)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/uploads/files/post/hash/Report.PDF", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d (location %q), want %d", w.Code, w.Header().Get("Location"), http.StatusOK)
	}
	if got := w.Body.String(); got != "%PDF-1.4" {
		t.Fatalf("body = %q, want the stored attachment", got)
	}
}

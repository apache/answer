/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the License, Version 2.0 (the
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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("recovers from panic and returns 500", func(t *testing.T) {
		r := gin.New()
		r.Use(Recovery())
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("response body is not valid JSON: %v", err)
		}

		if body["reason"] == nil || body["reason"] == "" {
			t.Fatal("expected non-empty reason field in response")
		}
	})

	t.Run("does not affect normal requests", func(t *testing.T) {
		r := gin.New()
		r.Use(Recovery())
		r.GET("/ok", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("response body is not valid JSON: %v", err)
		}

		if body["status"] != "ok" {
			t.Fatalf("expected status 'ok', got %v", body["status"])
		}
	})
}
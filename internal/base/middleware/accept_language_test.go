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
	"strings"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/translator"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/i18n"
	"github.com/stretchr/testify/assert"
)

func resolveAcceptLanguage(t *testing.T, header string) i18n.Language {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var resolved i18n.Language
	router := gin.New()
	router.Use(ExtractAndSetAcceptLanguage)
	router.GET("/", func(ctx *gin.Context) {
		value, _ := ctx.Get(constant.AcceptLanguageFlag)
		resolved, _ = value.(i18n.Language)
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set(constant.AcceptLanguageFlag, header)
	}
	router.ServeHTTP(httptest.NewRecorder(), req)
	return resolved
}

func TestExtractAndSetAcceptLanguage(t *testing.T) {
	original := translator.LanguageOptions
	translator.SetLanguageOptions([]*translator.LangOption{
		{Label: "Default", Value: translator.DefaultLangOption},
		{Label: "English", Value: "en_US"},
		{Label: "Español", Value: "es_ES"},
		{Label: "Português do Brasil", Value: "pt_BR"},
		{Label: "Português", Value: "pt_PT"},
		{Label: "简体中文", Value: "zh_CN"},
		{Label: "繁體中文", Value: "zh_TW"},
	})
	t.Cleanup(func() { translator.SetLanguageOptions(original) })

	tests := []struct {
		name   string
		header string
		want   i18n.Language
	}{
		{"exact configured locale", "es-ES", "es_ES"},
		{"exact locale with quality", "pt-BR,pt;q=0.8,en;q=0.5", "pt_BR"},
		{"regional Spanish falls back to configured Spanish", "es-MX", "es_ES"},
		{"Latin American Spanish", "es-419", "es_ES"},
		{"regional Portuguese prefers European Portuguese", "pt-AO", "pt_PT"},
		{"Hong Kong Chinese resolves to Traditional", "zh-HK", "zh_TW"},
		{"Singapore Chinese resolves to Simplified", "zh-SG", "zh_CN"},
		{"regional English", "en-GB", "en_US"},
		{"unsupported language uses English", "tlh", "en_US"},
		{"empty header uses English", "", "en_US"},
		{"malformed header uses English", "!!!", "en_US"},
		{"oversized header uses English", strings.Repeat("es-MX,", 100), "en_US"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveAcceptLanguage(t, tt.header))
		})
	}
}

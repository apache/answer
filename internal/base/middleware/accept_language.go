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
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/i18n"
	"golang.org/x/text/language"
	"strings"
)

// ExtractAndSetAcceptLanguage extract accept language from header and set to context
func ExtractAndSetAcceptLanguage(ctx *gin.Context) {
	// The language of our front-end configuration, like en_US
	lang := handler.GetLang(ctx)
	tag, _, err := language.ParseAcceptLanguage(string(lang))
	if err != nil || len(tag) == 0 {
		ctx.Set(constant.AcceptLanguageFlag, i18n.LanguageEnglish)
		return
	}

	acceptLang := strings.ReplaceAll(tag[0].String(), "-", "_")

	for _, option := range translator.LanguageOptions {
		if option.Value == acceptLang {
			ctx.Set(constant.AcceptLanguageFlag, i18n.Language(acceptLang))
			return
		}
	}

	// default language
	ctx.Set(constant.AcceptLanguageFlag, i18n.LanguageEnglish)
}

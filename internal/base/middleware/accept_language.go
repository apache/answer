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
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/i18n"
)

var (
	langMapping = map[i18n.Language]bool{
		i18n.LanguageChinese:            true,
		i18n.LanguageChineseTraditional: true,
		i18n.LanguageEnglish:            true,
		i18n.LanguageGerman:             true,
		i18n.LanguageSpanish:            true,
		i18n.LanguageFrench:             true,
		i18n.LanguageItalian:            true,
		i18n.LanguageJapanese:           true,
		i18n.LanguageKorean:             true,
		i18n.LanguagePortuguese:         true,
		i18n.LanguageRussian:            true,
		i18n.LanguageVietnamese:         true,
	}
)

// ExtractAndSetAcceptLanguage extract accept language from header and set to context
func ExtractAndSetAcceptLanguage(ctx *gin.Context) {
	// The language of our front-end configuration, like en_US
	lang := handler.GetLang(ctx)
	if langMapping[lang] {
		ctx.Set(constant.AcceptLanguageFlag, lang)
		return
	}

	// default language
	ctx.Set(constant.AcceptLanguageFlag, i18n.LanguageEnglish)
}

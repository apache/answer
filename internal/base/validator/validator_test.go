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

package validator

import (
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/segmentfault/pacman/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleCommentReq struct {
	OriginalText string `validate:"required,notblank,gte=2,lte=5" json:"original_text"`
}

type sampleAccentedFieldReq struct {
	Email string `validate:"required,email" json:"émail"`
}

func TestGetValidatorByLang_RegionalLocaleFallback(t *testing.T) {
	req := &sampleCommentReq{OriginalText: "123456"}

	esMXValidator := GetValidatorByLang(i18n.Language("es_MX"))
	require.NotNil(t, esMXValidator)
	assert.Equal(t, i18n.Language("es_MX"), esMXValidator.Lang)

	errFields, err := esMXValidator.Check(req)
	require.Error(t, err)
	require.Len(t, errFields, 1)
	assert.Contains(t, errFields[0].ErrorMsg, "debe tener un máximo de 5")

	ptBRValidator := GetValidatorByLang(i18n.Language("pt_BR"))
	require.NotNil(t, ptBRValidator)
	assert.Equal(t, i18n.Language("pt_BR"), ptBRValidator.Lang)

	errFieldsPT, errPT := ptBRValidator.Check(req)
	require.Error(t, errPT)
	require.Len(t, errFieldsPT, 1)
	assert.Contains(t, errFieldsPT[0].ErrorMsg, "máximo")
}

// A failing RegisterFunc must release the write lock before the fallback takes
// the read lock, otherwise the lookup deadlocks.
func TestGetValidatorByLang_RegisterFuncFailureDoesNotDeadlock(t *testing.T) {
	failing := &TranslatorLocal{
		La: i18n.Language("xx_XX"),
		Lo: en.New(),
		RegisterFunc: func(*validator.Validate, ut.Translator) error {
			return errors.New("register failed")
		},
	}
	original := allLanguageTranslators
	allLanguageTranslators = append(append([]*TranslatorLocal{}, original...), failing)
	t.Cleanup(func() { allLanguageTranslators = original })

	result := make(chan *MyValidator, 1)
	go func() { result <- GetValidatorByLang(i18n.Language("xx_YY")) }()

	select {
	case v := <-result:
		require.NotNil(t, v)
		assert.Equal(t, i18n.DefaultLanguage, v.Lang)
	case <-time.After(5 * time.Second):
		t.Fatal("GetValidatorByLang deadlocked after a RegisterFunc failure")
	}

	validatorMappingMu.RLock()
	_, cached := GlobalValidatorMapping[i18n.Language("xx_YY")]
	validatorMappingMu.RUnlock()
	assert.False(t, cached, "a failed registration must not be cached")
}

func TestCheck_MultiByteLatinCapitalization(t *testing.T) {
	req := &sampleAccentedFieldReq{Email: "not-an-email"}

	esValidator := GetValidatorByLang(i18n.LanguageSpanish)
	errFields, err := esValidator.Check(req)
	require.Error(t, err)
	require.Len(t, errFields, 1)

	msg := errFields[0].ErrorMsg
	assert.True(t, utf8.ValidString(msg), "error message must be valid UTF-8, got: %q", msg)
	assert.True(t, strings.HasPrefix(msg, "Émail "), "expected prefix 'Émail ', got: %q", msg)
}

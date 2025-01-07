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

package handler

import (
	"errors"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/validator"
	"github.com/gin-gonic/gin"
	myErrors "github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"net/http"
)

// HandleResponse Handle response body
func HandleResponse(ctx *gin.Context, err error, data interface{}) {
	lang := GetLang(ctx)
	// no error
	if err == nil {
		ctx.JSON(http.StatusOK, NewRespBodyData(http.StatusOK, reason.Success, data).TrMsg(lang))
		return
	}

	var myErr *myErrors.Error
	// unknown error
	if !errors.As(err, &myErr) {
		log.Error(err, "\n", myErrors.LogStack(2, 5))
		ctx.JSON(http.StatusInternalServerError, NewRespBody(
			http.StatusInternalServerError, reason.UnknownError).TrMsg(lang))
		return
	}

	// log internal server error
	if myErrors.IsInternalServer(myErr) {
		log.Error(myErr)
	}

	respBody := NewRespBodyFromError(myErr).TrMsg(lang)
	if data != nil {
		respBody.Data = data
	}
	ctx.JSON(myErr.Code, respBody)
}

// BindAndCheck bind request and check
func BindAndCheck(ctx *gin.Context, data interface{}) bool {
	lang := GetLang(ctx)
	ctx.Set(constant.AcceptLanguageFlag, lang)
	if err := ctx.ShouldBind(data); err != nil {
		log.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, myErrors.New(http.StatusBadRequest, reason.RequestFormatError), nil)
		return true
	}

	errField, err := validator.GetValidatorByLang(lang).Check(data)
	if err != nil {
		HandleResponse(ctx, err, errField)
		return true
	}
	return false
}

// BindAndCheckReturnErr bind request and check
func BindAndCheckReturnErr(ctx *gin.Context, data interface{}) (errFields []*validator.FormErrorField) {
	lang := GetLang(ctx)
	if err := ctx.ShouldBind(data); err != nil {
		log.Errorf("http_handle BindAndCheck fail, %s", err.Error())
		HandleResponse(ctx, myErrors.New(http.StatusBadRequest, reason.RequestFormatError), nil)
		ctx.Abort()
		return nil
	}

	errFields, _ = validator.GetValidatorByLang(lang).Check(data)
	return errFields
}

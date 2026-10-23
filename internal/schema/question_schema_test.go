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

package schema

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestQuestionPageReqDoesNotBindInternalFields(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet,
		"/?page=1&page_size=20&username=author&LoginUserID=attacker&UserIDBeSearched=attacker&TagID=tag&ShowPending=true", nil)

	var req QuestionPageReq
	if err := ctx.ShouldBind(&req); err != nil {
		t.Fatalf("bind QuestionPageReq: %v", err)
	}
	if req.Page != 1 || req.PageSize != 20 || req.Username != "author" {
		t.Fatalf("public fields were not bound correctly: %+v", req)
	}
	if req.LoginUserID != "" || req.UserIDBeSearched != "" || req.TagID != "" || req.ShowPending {
		t.Fatalf("internal fields must not be bound from query parameters: %+v", req)
	}
}

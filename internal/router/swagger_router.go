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
	"fmt"

	"github.com/apache/answer/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerRouter swagger api router
type SwaggerRouter struct {
	config *SwaggerConfig
}

// NewSwaggerRouter new swagger api router
func NewSwaggerRouter(config *SwaggerConfig) *SwaggerRouter {
	return &SwaggerRouter{
		config: config,
	}
}

// Register register swagger api router
func (a *SwaggerRouter) Register(r *gin.RouterGroup) {
	if a.config.Show {
		a.InitSwaggerDocs()
		gofmt := fmt.Sprintf("%s://%s%s/swagger/doc.json", a.config.Protocol, a.config.Host, a.config.Address)
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.URL(gofmt)))
	}
}

// InitSwaggerDocs init swagger docs
func (a *SwaggerRouter) InitSwaggerDocs() {
	docs.SwaggerInfo.Host = fmt.Sprintf("%s%s", a.config.Host, a.config.Address)
}

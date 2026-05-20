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

package install

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/apache/answer/configs"
	"github.com/apache/answer/internal/base/conf"
	"github.com/apache/answer/ui"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
	"gopkg.in/yaml.v3"
)

const UIStaticPath = "build/static"

type _resource struct {
	fs embed.FS
}

// Open to implement the interface by http.FS required
func (r *_resource) Open(name string) (fs.File, error) {
	name = fmt.Sprintf(UIStaticPath+"/%s", name)
	log.Debugf("open static path %s", name)
	return r.fs.Open(name)
}

// NewInstallHTTPServer new install http server.
func NewInstallHTTPServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	c := &conf.AllConfig{}
	_ = yaml.Unmarshal(configs.Config, c)

	r.GET("/healthz", func(ctx *gin.Context) { ctx.String(200, "OK") })
	r.StaticFS(c.UI.BaseURL+"/static", http.FS(&_resource{
		fs: ui.Build,
	}))

	// read default config file and extract ui config
	installApi := r.Group("")
	installApi.GET(c.UI.BaseURL+"/", CheckConfigFileAndRedirectToInstallPage)
	installApi.GET(c.UI.BaseURL+"/install", WebPage)
	installApi.GET(c.UI.BaseURL+"/50x", WebPage)
	installApi.GET(c.UI.APIBaseURL+"/installation/language/config", GetLangMapping)
	installApi.GET(c.UI.APIBaseURL+"/installation/language/options", LangOptions)
	installApi.POST(c.UI.APIBaseURL+"/installation/db/check", CheckDatabase)
	installApi.POST(c.UI.APIBaseURL+"/installation/config-file/check", CheckConfigFile)
	installApi.POST(c.UI.APIBaseURL+"/installation/init", InitEnvironment)
	installApi.POST(c.UI.APIBaseURL+"/installation/base-info", InitBaseInfo)

	r.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/50x")
	})
	return r
}

func WebPage(c *gin.Context) {
	filePath := ""
	var file []byte
	var err error
	filePath = "build/index.html"
	c.Header("content-type", "text/html;charset=utf-8")
	file, err = ui.Build.ReadFile(filePath)
	if err != nil {
		log.Error(err)
		c.Status(http.StatusNotFound)
		return
	}
	c.String(http.StatusOK, string(file))
}

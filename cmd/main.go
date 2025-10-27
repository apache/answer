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

package answercmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apache/answer/internal/base/conf"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/cron"
	"github.com/apache/answer/internal/base/path"
	"github.com/apache/answer/internal/schema"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman"
	"github.com/segmentfault/pacman/contrib/log/zap"
	"github.com/segmentfault/pacman/contrib/server/http"
	"github.com/segmentfault/pacman/log"
)

// go build -ldflags "-X github.com/apache/answer/cmd.Version=x.y.z"
var (
	// Name is the name of the project
	Name = "answer"
	// Version is the version of the project
	Version = "0.0.0"
	// Revision is the git short commit revision number
	// If built without a Git repository, this field will be empty.
	Revision = ""
	// Time is the build time of the project
	Time = ""
	// GoVersion is the go version of the project
	GoVersion = "1.23"
	// log level
	logLevel = os.Getenv("LOG_LEVEL")
	// log path
	logPath = os.Getenv("LOG_PATH")
)

// Main
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func Main() {
	log.SetLogger(zap.NewLogger(
		log.ParseLevel(logLevel), zap.WithName("answer"), zap.WithPath(logPath)))
	Execute()
}

func runApp() {
	c, err := conf.ReadConfig(path.GetConfigFilePath())
	if err != nil {
		panic(err)
	}
	app, cleanup, err := initApplication(
		c.Debug, c.Server, c.Data.Database, c.Data.Cache, c.I18n, c.Swaggerui, c.ServiceConfig, c.UI, log.GetLogger())
	if err != nil {
		panic(err)
	}
	constant.Version = Version
	constant.Revision = Revision
	constant.GoVersion = GoVersion
	schema.AppStartTime = time.Now()
	fmt.Println("answer Version:", constant.Version, " Revision:", constant.Revision)

	defer cleanup()
	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}

func newApplication(serverConf *conf.Server, server *gin.Engine, manager *cron.ScheduledTaskManager) *pacman.Application {
	manager.Run()
	return pacman.NewApp(
		pacman.WithName(Name),
		pacman.WithVersion(Version),
		pacman.WithServer(http.NewServer(server, serverConf.HTTP.Addr)),
	)
}

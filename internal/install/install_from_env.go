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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type Env struct {
	AutoInstall string `json:"auto_install"`
	DbType      string `json:"db_type"`
	DbUsername  string `json:"db_username"`
	DbPassword  string `json:"db_password"`
	DbHost      string `json:"db_host"`
	DbName      string `json:"db_name"`
	DbFile      string `json:"db_file"`
	Language    string `json:"lang"`

	SiteName               string `json:"site_name"`
	SiteURL                string `json:"site_url"`
	ContactEmail           string `json:"contact_email"`
	AdminName              string `json:"name"`
	AdminPassword          string `json:"password"`
	AdminEmail             string `json:"email"`
	LoginRequired          bool   `json:"login_required"`
	ExternalContentDisplay string `json:"external_content_display"`
}

func TryToInstallByEnv() (installByEnv bool, err error) {
	env := loadEnv()
	if len(env.AutoInstall) == 0 {
		return false, nil
	}
	fmt.Println("[auto-install] try to install by environment variable")
	return true, initByEnv(env)
}

func loadEnv() (env *Env) {
	return &Env{
		AutoInstall:            os.Getenv("AUTO_INSTALL"),
		DbType:                 os.Getenv("DB_TYPE"),
		DbUsername:             os.Getenv("DB_USERNAME"),
		DbPassword:             os.Getenv("DB_PASSWORD"),
		DbHost:                 os.Getenv("DB_HOST"),
		DbName:                 os.Getenv("DB_NAME"),
		DbFile:                 os.Getenv("DB_FILE"),
		Language:               os.Getenv("LANGUAGE"),
		SiteName:               os.Getenv("SITE_NAME"),
		SiteURL:                os.Getenv("SITE_URL"),
		ContactEmail:           os.Getenv("CONTACT_EMAIL"),
		AdminName:              os.Getenv("ADMIN_NAME"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		AdminEmail:             os.Getenv("ADMIN_EMAIL"),
		ExternalContentDisplay: os.Getenv("EXTERNAL_CONTENT_DISPLAY"),
	}
}

func initByEnv(env *Env) (err error) {
	gin.SetMode(gin.TestMode)
	if err = dbCheck(env); err != nil {
		return err
	}
	if err = initConfigAndDb(env); err != nil {
		return err
	}
	if err = initBaseInfo(env); err != nil {
		return err
	}
	return nil
}

func dbCheck(env *Env) (err error) {
	req := &CheckDatabaseReq{
		DbType:     env.DbType,
		DbUsername: env.DbUsername,
		DbPassword: env.DbPassword,
		DbHost:     env.DbHost,
		DbName:     env.DbName,
		DbFile:     env.DbFile,
	}
	return requestAPI(req, "POST", "/installation/db/check", CheckDatabase)
}

func initConfigAndDb(env *Env) (err error) {
	req := &CheckDatabaseReq{
		DbType:     env.DbType,
		DbUsername: env.DbUsername,
		DbPassword: env.DbPassword,
		DbHost:     env.DbHost,
		DbName:     env.DbName,
		DbFile:     env.DbFile,
	}
	return requestAPI(req, "POST", "/installation/init", InitEnvironment)
}

func initBaseInfo(env *Env) (err error) {
	req := &InitBaseInfoReq{
		Language:               env.Language,
		SiteName:               env.SiteName,
		SiteURL:                env.SiteURL,
		ContactEmail:           env.ContactEmail,
		AdminName:              env.AdminName,
		AdminPassword:          env.AdminPassword,
		AdminEmail:             env.AdminEmail,
		LoginRequired:          env.LoginRequired,
		ExternalContentDisplay: env.ExternalContentDisplay,
	}
	return requestAPI(req, "POST", "/installation/base-info", InitBaseInfo)
}

func requestAPI(req interface{}, method, url string, handlerFunc gin.HandlerFunc) error {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest(method, url, bytes.NewBuffer(body))
	if method == "POST" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	handlerFunc(c)
	if w.Code != http.StatusOK {
		return fmt.Errorf(gjson.Get(w.Body.String(), "msg").String())
	}
	return nil
}

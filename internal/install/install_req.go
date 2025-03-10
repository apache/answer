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
	"fmt"
	"net/url"
	"strings"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/dir"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm/schemas"
)

// CheckConfigFileResp check config file if exist or not response
type CheckConfigFileResp struct {
	ConfigFileExist     bool `json:"config_file_exist"`
	DBConnectionSuccess bool `json:"db_connection_success"`
	DbTableExist        bool `json:"db_table_exist"`
}

// CheckDatabaseReq check database
type CheckDatabaseReq struct {
	DbType       string `validate:"required,oneof=postgres sqlite3 mysql" json:"db_type"`
	DbUsername   string `json:"db_username"`
	DbPassword   string `json:"db_password"`
	DbHost       string `json:"db_host"`
	DbName       string `json:"db_name"`
	DbFile       string `json:"db_file"`
	Ssl          bool   `json:"ssl_enabled"`
	SslMode      string `json:"ssl_mode"`
	SslCrt       string `json:"pem_file"`
	SslKey       string `json:"key_file"`
	SslCrtClient string `json:"cert_file"`
}

// GetConnection get connection string
func (r *CheckDatabaseReq) GetConnection() string {
	if r.DbType == string(schemas.SQLITE) {
		return r.DbFile
	}
	if r.DbType == string(schemas.MYSQL) {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s",
			r.DbUsername, r.DbPassword, r.DbHost, r.DbName)
	}
	if r.DbType == string(schemas.POSTGRES) {
		host, port := parsePgSQLHostPort(r.DbHost)
		if !r.Ssl {
			return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				host, port, r.DbUsername, r.DbPassword, r.DbName)
		} else if r.SslMode == "require" {
			return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				host, port, r.DbUsername, r.DbPassword, r.DbName, r.SslMode)
		} else if r.SslMode == "verify-ca" || r.SslMode == "verify-full" {
			if dir.CheckFileExist(r.SslCrt) {
				log.Warnf("ssl crt file not exist: %s", r.SslCrt)
			}
			if dir.CheckFileExist(r.SslCrtClient) {
				log.Warnf("ssl crt client file not exist: %s", r.SslCrtClient)
			}
			if dir.CheckFileExist(r.SslKey) {
				log.Warnf("ssl key file not exist: %s", r.SslKey)
			}
			return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s sslcert=%s sslkey=%s",
				host, port, r.DbUsername, r.DbPassword, r.DbName, r.SslMode, r.SslCrt, r.SslCrtClient, r.SslKey)
		}
	}
	return ""
}

func parsePgSQLHostPort(dbHost string) (host string, port string) {
	if strings.Contains(dbHost, ":") {
		idx := strings.LastIndex(dbHost, ":")
		host, port = dbHost[:idx], dbHost[idx+1:]
	} else if len(dbHost) > 0 {
		host = dbHost
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "5432"
	}
	return host, port
}

// CheckDatabaseResp check database response
type CheckDatabaseResp struct {
	ConnectionSuccess bool `json:"connection_success"`
}

// InitEnvironmentResp init environment response
type InitEnvironmentResp struct {
	Success            bool   `json:"success"`
	CreateConfigFailed bool   `json:"create_config_failed"`
	DefaultConfig      string `json:"default_config"`
	ErrType            string `json:"err_type"`
}

// InitBaseInfoReq init base info request
type InitBaseInfoReq struct {
	Language      string `validate:"required,gt=0,lte=30" json:"lang"`
	SiteName      string `validate:"required,sanitizer,gt=0,lte=30" json:"site_name"`
	SiteURL       string `validate:"required,gt=0,lte=512,url" json:"site_url"`
	ContactEmail  string `validate:"required,email,gt=0,lte=500" json:"contact_email"`
	AdminName     string `validate:"required,gt=3,lte=30" json:"name"`
	AdminPassword string `validate:"required,gte=8,lte=32" json:"password"`
	AdminEmail    string `validate:"required,email,gt=0,lte=500" json:"email"`
	LoginRequired bool   `json:"login_required"`
}

func (r *InitBaseInfoReq) Check() (errFields []*validator.FormErrorField, err error) {
	if checker.IsInvalidUsername(r.AdminName) {
		errField := &validator.FormErrorField{
			ErrorField: "name",
			ErrorMsg:   reason.UsernameInvalid,
		}
		errFields = append(errFields, errField)
		return errFields, errors.BadRequest(reason.UsernameInvalid)
	}
	return
}

func (r *InitBaseInfoReq) FormatSiteUrl() {
	parsedUrl, err := url.Parse(r.SiteURL)
	if err != nil {
		return
	}
	r.SiteURL = fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	if len(parsedUrl.Path) > 0 {
		r.SiteURL = r.SiteURL + parsedUrl.Path
		r.SiteURL = strings.TrimSuffix(r.SiteURL, "/")
	}
}

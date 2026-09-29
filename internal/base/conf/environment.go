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

package conf

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	envDatabaseDriver     = "ANSWER_DATA_DATABASE_DRIVER"
	envDatabaseConnection = "ANSWER_DATA_DATABASE_CONNECTION"
)

type environmentVariable struct {
	name    string
	aliases []string
	apply   func(*AllConfig, string) error
	value   func(*AllConfig) string
}

func stringEnvironmentVariable(
	name string,
	aliases []string,
	field func(*AllConfig) *string,
) environmentVariable {
	return environmentVariable{
		name:    name,
		aliases: aliases,
		apply: func(config *AllConfig, value string) error {
			*field(config) = value
			return nil
		},
		value: func(config *AllConfig) string {
			return *field(config)
		},
	}
}

func boolEnvironmentVariable(name string, field func(*AllConfig) *bool) environmentVariable {
	return environmentVariable{
		name: name,
		apply: func(config *AllConfig, value string) error {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("%s must be a boolean: %w", name, err)
			}
			*field(config) = parsed
			return nil
		},
		value: func(config *AllConfig) string {
			return strconv.FormatBool(*field(config))
		},
	}
}

func intEnvironmentVariable(name string, field func(*AllConfig) *int) environmentVariable {
	return environmentVariable{
		name: name,
		apply: func(config *AllConfig, value string) error {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("%s must be an integer: %w", name, err)
			}
			*field(config) = parsed
			return nil
		},
		value: func(config *AllConfig) string {
			return strconv.Itoa(*field(config))
		},
	}
}

func runtimeEnvironmentVariables() []environmentVariable {
	return []environmentVariable{
		boolEnvironmentVariable("ANSWER_DEBUG", func(config *AllConfig) *bool {
			return &config.Debug
		}),
		stringEnvironmentVariable("ANSWER_SERVER_HTTP_ADDR", []string{"SITE_ADDR"}, func(config *AllConfig) *string {
			return &config.Server.HTTP.Addr
		}),
		stringEnvironmentVariable(envDatabaseDriver, nil, func(config *AllConfig) *string {
			return &config.Data.Database.Driver
		}),
		stringEnvironmentVariable(envDatabaseConnection, nil, func(config *AllConfig) *string {
			return &config.Data.Database.Connection
		}),
		intEnvironmentVariable("ANSWER_DATA_DATABASE_CONN_MAX_LIFE_TIME", func(config *AllConfig) *int {
			return &config.Data.Database.ConnMaxLifeTime
		}),
		intEnvironmentVariable("ANSWER_DATA_DATABASE_MAX_OPEN_CONN", func(config *AllConfig) *int {
			return &config.Data.Database.MaxOpenConn
		}),
		intEnvironmentVariable("ANSWER_DATA_DATABASE_MAX_IDLE_CONN", func(config *AllConfig) *int {
			return &config.Data.Database.MaxIdleConn
		}),
		stringEnvironmentVariable("ANSWER_DATA_CACHE_FILE_PATH", nil, func(config *AllConfig) *string {
			return &config.Data.Cache.FilePath
		}),
		stringEnvironmentVariable("ANSWER_I18N_BUNDLE_DIR", nil, func(config *AllConfig) *string {
			return &config.I18n.BundleDir
		}),
		stringEnvironmentVariable("ANSWER_SERVICE_CONFIG_UPLOAD_PATH", nil, func(config *AllConfig) *string {
			return &config.ServiceConfig.UploadPath
		}),
		boolEnvironmentVariable("ANSWER_SERVICE_CONFIG_CLEAN_UP_UPLOADS", func(config *AllConfig) *bool {
			return &config.ServiceConfig.CleanUpUploads
		}),
		intEnvironmentVariable("ANSWER_SERVICE_CONFIG_CLEAN_ORPHAN_UPLOADS_PERIOD_HOURS", func(config *AllConfig) *int {
			return &config.ServiceConfig.CleanOrphanUploadsPeriodHours
		}),
		intEnvironmentVariable("ANSWER_SERVICE_CONFIG_PURGE_DELETED_FILES_PERIOD_DAYS", func(config *AllConfig) *int {
			return &config.ServiceConfig.PurgeDeletedFilesPeriodDays
		}),
		boolEnvironmentVariable("ANSWER_SWAGGERUI_SHOW", func(config *AllConfig) *bool {
			return &config.Swaggerui.Show
		}),
		stringEnvironmentVariable("ANSWER_SWAGGERUI_PROTOCOL", nil, func(config *AllConfig) *string {
			return &config.Swaggerui.Protocol
		}),
		stringEnvironmentVariable("ANSWER_SWAGGERUI_HOST", []string{"SWAGGER_HOST"}, func(config *AllConfig) *string {
			return &config.Swaggerui.Host
		}),
		stringEnvironmentVariable("ANSWER_SWAGGERUI_ADDRESS", []string{"SWAGGER_ADDRESS_PORT"}, func(config *AllConfig) *string {
			return &config.Swaggerui.Address
		}),
		stringEnvironmentVariable("ANSWER_UI_BASE_URL", nil, func(config *AllConfig) *string {
			return &config.UI.BaseURL
		}),
		stringEnvironmentVariable("ANSWER_UI_API_BASE_URL", nil, func(config *AllConfig) *string {
			return &config.UI.APIBaseURL
		}),
	}
}

func lookupEnvironment(variable environmentVariable) (string, bool) {
	if value, ok := os.LookupEnv(variable.name); ok {
		return value, true
	}
	for _, alias := range variable.aliases {
		if value, ok := os.LookupEnv(alias); ok && value != "" {
			return value, true
		}
	}
	return "", false
}

// RuntimeEnvironmentConfigured reports whether the required database settings
// are present. Requiring both avoids silently treating an accidentally missing
// config file as a valid environment-only deployment.
func RuntimeEnvironmentConfigured() bool {
	driver, driverSet := os.LookupEnv(envDatabaseDriver)
	connection, connectionSet := os.LookupEnv(envDatabaseConnection)
	return driverSet && driver != "" && connectionSet && connection != ""
}

// SetEnvironmentOverrides applies runtime environment variables to c. Canonical
// ANSWER_* variables take precedence over the three legacy aliases.
func (c *AllConfig) SetEnvironmentOverrides() error {
	c.SetDefault()
	for _, variable := range runtimeEnvironmentVariables() {
		value, ok := lookupEnvironment(variable)
		if !ok {
			continue
		}
		if err := variable.apply(c, value); err != nil {
			return err
		}
	}
	return nil
}

// ExportEnvironment serializes every runtime configuration value in dotenv
// format. Database connection strings may contain credentials, so callers must
// only expose the returned value as the result of an explicit user action.
func ExportEnvironment(c *AllConfig) (string, error) {
	c.SetDefault()
	values := make(map[string]string, len(runtimeEnvironmentVariables()))
	for _, variable := range runtimeEnvironmentVariables() {
		values[variable.name] = variable.value(c)
	}
	return godotenv.Marshal(values)
}

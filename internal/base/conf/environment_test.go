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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/apache/answer/configs"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearRuntimeEnvironment(t *testing.T) {
	t.Helper()
	for _, variable := range runtimeEnvironmentVariables() {
		names := append([]string{variable.name}, variable.aliases...)
		for _, name := range names {
			oldValue, wasSet := os.LookupEnv(name)
			require.NoError(t, os.Unsetenv(name))
			t.Cleanup(func() {
				if wasSet {
					require.NoError(t, os.Setenv(name, oldValue))
					return
				}
				require.NoError(t, os.Unsetenv(name))
			})
		}
	}
}

func writeDefaultConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, configs.Config, 0o600))
	return configPath
}

func countRuntimeConfigFields(configType reflect.Type) int {
	if configType.Kind() == reflect.Pointer {
		configType = configType.Elem()
	}
	if configType.Kind() != reflect.Struct {
		return 1
	}

	count := 0
	for index := range configType.NumField() {
		field := configType.Field(index)
		if !field.IsExported() {
			continue
		}
		count += countRuntimeConfigFields(field.Type)
	}
	return count
}

func TestReadConfigRequiresFileOrRuntimeDatabaseEnvironment(t *testing.T) {
	clearRuntimeEnvironment(t)

	_, err := ReadConfig(filepath.Join(t.TempDir(), "missing.yaml"))

	require.Error(t, err)
}

func TestReadConfigFromEnvironmentWithoutFile(t *testing.T) {
	clearRuntimeEnvironment(t)
	t.Setenv(envDatabaseDriver, "postgres")
	t.Setenv(envDatabaseConnection, "postgres://answer:secret@database/answer?sslmode=require")
	t.Setenv("ANSWER_SERVER_HTTP_ADDR", "0.0.0.0:8080")
	t.Setenv("ANSWER_DATA_DATABASE_MAX_OPEN_CONN", "25")
	t.Setenv("ANSWER_SERVICE_CONFIG_CLEAN_UP_UPLOADS", "false")
	t.Setenv("ANSWER_SWAGGERUI_SHOW", "false")
	t.Setenv("ANSWER_SWAGGERUI_PROTOCOL", "")

	config, err := ReadConfig(filepath.Join(t.TempDir(), "missing.yaml"))

	require.NoError(t, err)
	assert.Equal(t, "postgres", config.Data.Database.Driver)
	assert.Equal(t, "postgres://answer:secret@database/answer?sslmode=require", config.Data.Database.Connection)
	assert.Equal(t, 25, config.Data.Database.MaxOpenConn)
	assert.Equal(t, "0.0.0.0:8080", config.Server.HTTP.Addr)
	assert.False(t, config.ServiceConfig.CleanUpUploads)
	assert.False(t, config.Swaggerui.Show)
	assert.Empty(t, config.Swaggerui.Protocol)
	assert.Equal(t, "/data/cache/cache.db", config.Data.Cache.FilePath)
}

func TestEnvironmentOverridesConfigAndCanonicalNameWins(t *testing.T) {
	clearRuntimeEnvironment(t)
	t.Setenv("SITE_ADDR", "0.0.0.0:8080")
	t.Setenv("ANSWER_SERVER_HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("ANSWER_DATA_DATABASE_CONN_MAX_LIFE_TIME", "120")

	config, err := ReadConfig(writeDefaultConfig(t))

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", config.Server.HTTP.Addr)
	assert.Equal(t, 120, config.Data.Database.ConnMaxLifeTime)
}

func TestReadConfigRejectsInvalidEnvironmentValue(t *testing.T) {
	clearRuntimeEnvironment(t)
	t.Setenv("ANSWER_SWAGGERUI_SHOW", "sometimes")

	_, err := ReadConfig(writeDefaultConfig(t))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ANSWER_SWAGGERUI_SHOW must be a boolean")
}

func TestRuntimeEnvironmentConfiguredRequiresDatabasePair(t *testing.T) {
	clearRuntimeEnvironment(t)
	assert.False(t, RuntimeEnvironmentConfigured())

	t.Setenv(envDatabaseDriver, "mysql")
	assert.False(t, RuntimeEnvironmentConfigured())

	t.Setenv(envDatabaseConnection, "answer:secret@tcp(database:3306)/answer")
	assert.True(t, RuntimeEnvironmentConfigured())
}

func TestExportEnvironmentIncludesEveryCanonicalVariable(t *testing.T) {
	clearRuntimeEnvironment(t)
	config, err := ReadConfig(writeDefaultConfig(t))
	require.NoError(t, err)
	config.Data.Database.Driver = "mysql"
	config.Data.Database.Connection = "answer:p@ss word@tcp(database:3306)/answer"

	output, err := ExportEnvironment(config)
	require.NoError(t, err)
	values, err := godotenv.Unmarshal(output)
	require.NoError(t, err)

	assert.Len(t, values, countRuntimeConfigFields(reflect.TypeOf(AllConfig{})))
	assert.Equal(t, "mysql", values[envDatabaseDriver])
	assert.Equal(t, "answer:p@ss word@tcp(database:3306)/answer", values[envDatabaseConnection])
	for _, variable := range runtimeEnvironmentVariables() {
		assert.Contains(t, values, variable.name)
	}
}

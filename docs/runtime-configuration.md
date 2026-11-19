<!--
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing,
 software distributed under the License is distributed on an
 "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 KIND, either express or implied.  See the License for the
 specific language governing permissions and limitations
 under the License.
-->

# Runtime configuration from environment variables

Answer normally reads its runtime configuration from `/data/conf/config.yaml`.
It can also run without that file when both
`ANSWER_DATA_DATABASE_DRIVER` and `ANSWER_DATA_DATABASE_CONNECTION` are set to
non-empty values. This explicit requirement prevents an accidentally missing
configuration file from silently starting Answer with the embedded SQLite
defaults.

When `config.yaml` exists, canonical `ANSWER_*` environment variables override
the corresponding file values. `SITE_ADDR`, `SWAGGER_HOST`, and
`SWAGGER_ADDRESS_PORT` remain supported for backward compatibility, but their
canonical replacements take precedence when both names are set.

When `config.yaml` does not exist and the two required database variables are
present, Answer starts with the embedded configuration template and applies all
set environment variables over it. Values omitted from the environment retain
the embedded defaults shown below.

| Environment variable | `config.yaml` field | Embedded default |
| --- | --- | --- |
| `ANSWER_DEBUG` | `debug` | `false` |
| `ANSWER_SERVER_HTTP_ADDR` | `server.http.addr` | `0.0.0.0:80` |
| `ANSWER_DATA_DATABASE_DRIVER` | `data.database.driver` | `sqlite3` |
| `ANSWER_DATA_DATABASE_CONNECTION` | `data.database.connection` | `/data/sqlite3/answer.db` |
| `ANSWER_DATA_DATABASE_CONN_MAX_LIFE_TIME` | `data.database.conn_max_life_time` | `0` |
| `ANSWER_DATA_DATABASE_MAX_OPEN_CONN` | `data.database.max_open_conn` | `0` |
| `ANSWER_DATA_DATABASE_MAX_IDLE_CONN` | `data.database.max_idle_conn` | `0` |
| `ANSWER_DATA_CACHE_FILE_PATH` | `data.cache.file_path` | `/data/cache/cache.db` |
| `ANSWER_I18N_BUNDLE_DIR` | `i18n.bundle_dir` | `/data/i18n` |
| `ANSWER_SERVICE_CONFIG_UPLOAD_PATH` | `service_config.upload_path` | `/data/uploads` |
| `ANSWER_SERVICE_CONFIG_CLEAN_UP_UPLOADS` | `service_config.clean_up_uploads` | `true` |
| `ANSWER_SERVICE_CONFIG_CLEAN_ORPHAN_UPLOADS_PERIOD_HOURS` | `service_config.clean_orphan_uploads_period_hours` | `48` |
| `ANSWER_SERVICE_CONFIG_PURGE_DELETED_FILES_PERIOD_DAYS` | `service_config.purge_deleted_files_period_days` | `30` |
| `ANSWER_SWAGGERUI_SHOW` | `swaggerui.show` | `true` |
| `ANSWER_SWAGGERUI_PROTOCOL` | `swaggerui.protocol` | `http` |
| `ANSWER_SWAGGERUI_HOST` | `swaggerui.host` | `127.0.0.1` |
| `ANSWER_SWAGGERUI_ADDRESS` | `swaggerui.address` | `:80` |
| `ANSWER_UI_BASE_URL` | `ui.base_url` | empty |
| `ANSWER_UI_API_BASE_URL` | `ui.api_base_url` | empty |

Boolean values use Go boolean syntax such as `true` or `false`. Integer values
must be base-10 integers. An invalid value stops configuration loading and names
the invalid variable.

## Exporting an installed configuration

After installation, export the effective runtime configuration in dotenv format:

```bash
umask 077
answer config export-env -C /data > answer-runtime.env
```

The command writes only the dotenv document to standard output, so it can also
be piped to a secret-management command. The database connection can contain a
password. Answer never emits the exported document during ordinary `init`,
`upgrade`, or `run` operations; it is produced only by this explicit command.

The exported file contains the complete canonical variable set. Store it as a
secret, not as a ConfigMap or a source-controlled file.

To start another instance against the already initialized database, inject the
exported values and run Answer normally. `AUTO_INSTALL` is not needed and should
not be set for this runtime-only path.

Configuration statelessness does not make uploaded files persistent. Configure
the appropriate upload storage plugin or retain storage for the upload path if
uploads must survive pod replacement.

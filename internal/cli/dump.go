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

package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/apache/answer/internal/base/data"
	"xorm.io/xorm/schemas"
)

// DumpAllData dump all database data to sql
func DumpAllData(dataConf *data.Database, dumpDataPath string) error {
	db, err := data.NewDB(false, dataConf)
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		return err
	}

	name := filepath.Join(dumpDataPath, fmt.Sprintf("answer_dump_data_%s.sql", time.Now().Format("2006-01-02")))
	return db.DumpAllToFile(name, schemas.MYSQL)
}

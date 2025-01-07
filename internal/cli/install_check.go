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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/pkg/dir"
)

func CheckConfigFile(configPath string) bool {
	return dir.CheckFileExist(configPath)
}

func CheckUploadDir() bool {
	return dir.CheckDirExist(UploadFilePath)
}

// CheckDBConnection check database whether the connection is normal
func CheckDBConnection(dataConf *data.Database) bool {
	db, err := data.NewDB(false, dataConf)
	if err != nil {
		fmt.Printf("connection database failed: %s\n", err)
		return false
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		fmt.Printf("connection ping database failed: %s\n", err)
		return false
	}

	return true
}

// CheckDBTableExist check database whether the table is already exists
func CheckDBTableExist(dataConf *data.Database) bool {
	db, err := data.NewDB(false, dataConf)
	if err != nil {
		fmt.Printf("connection database failed: %s\n", err)
		return false
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		fmt.Printf("connection ping database failed: %s\n", err)
		return false
	}

	exist, err := db.IsTableExist(&entity.Version{})
	if err != nil {
		fmt.Printf("check table exist failed: %s\n", err)
		return false
	}
	if !exist {
		fmt.Printf("check table not exist\n")
		return false
	}
	return true
}

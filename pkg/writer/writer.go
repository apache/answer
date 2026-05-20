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

package writer

import (
	"bufio"
	"os"
)

// ReplaceFile remove old file and write new file
func ReplaceFile(filePath, content string) error {
	_ = os.Remove(filePath)
	return WriteFile(filePath, content)
}

// WriteFile write file to path
func WriteFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	writer := bufio.NewWriter(file)
	if _, err := writer.WriteString(content); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

// MoveFile move file to new path
func MoveFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

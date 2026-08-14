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

package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeAndCheckImageFileRejectsUnsupportedExtension(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-an-image.svg")
	if err := os.WriteFile(filePath, []byte("<svg></svg>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if DecodeAndCheckImageFile(filePath, 1_000_000) {
		t.Fatal("unsupported image extensions must be rejected")
	}
}

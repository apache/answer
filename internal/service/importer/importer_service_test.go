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

package importer

import (
	"strings"
	"testing"

	"github.com/apache/answer/plugin"
)

func TestNewImportedQuestionRequestSanitizesContent(t *testing.T) {
	request := newImportedQuestionRequest(plugin.QuestionImporterInfo{
		Title:   "Imported question",
		Content: `<img src=x onerror=alert(1)><script>alert(1)</script>`,
		Tags:    []string{"security"},
	})

	for _, unsafeContent := range []string{"onerror", "<script"} {
		if strings.Contains(strings.ToLower(request.HTML), unsafeContent) {
			t.Fatalf("imported question HTML contains unsafe content %q: %s", unsafeContent, request.HTML)
		}
	}
}

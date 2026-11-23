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

package i18n

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type searchTips struct {
	UI struct {
		Search struct {
			Tips struct {
				Tag      string `yaml:"tag"`
				User     string `yaml:"user"`
				Answer   string `yaml:"answer"`
				Score    string `yaml:"score"`
				Question string `yaml:"question"`
				IsAnswer string `yaml:"is_answer"`
			} `yaml:"tips"`
		} `yaml:"search"`
	} `yaml:"ui"`
}

// TestSearchTipsUseParserOperators guards the examples shown on the search
// page. They are meant to be typed into the search box, but the parser
// (internal/service/search_parser) matches only the English operators, so a
// translated operator inside the <1></1> code span produces a useless example.
func TestSearchTipsUseParserOperators(t *testing.T) {
	files, err := fs.Glob(I18n, "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	codeSpan := regexp.MustCompile(`<1>(.*?)</1>`)
	tag := regexp.MustCompile(`^\[.+\]$`)

	for _, file := range files {
		if file == "i18n.yaml" {
			continue
		}
		data, err := I18n.ReadFile(file)
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		var tr searchTips
		if err = yaml.Unmarshal(data, &tr); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		tips := tr.UI.Search.Tips
		cases := []struct {
			key   string
			value string
			valid func(string) bool
		}{
			{"tag", tips.Tag, tag.MatchString},
			{"user", tips.User, func(v string) bool { return strings.HasPrefix(v, "user:") }},
			{"answer", tips.Answer, func(v string) bool { return strings.HasPrefix(v, "answers:") }},
			{"score", tips.Score, func(v string) bool { return strings.HasPrefix(v, "score:") }},
			{"question", tips.Question, func(v string) bool { return v == "is:question" }},
			{"is_answer", tips.IsAnswer, func(v string) bool { return v == "is:answer" }},
		}
		for _, c := range cases {
			m := codeSpan.FindStringSubmatch(c.value)
			if m == nil {
				t.Errorf("%s: tips.%s has no <1> code span: %q", file, c.key, c.value)
				continue
			}
			if !c.valid(m[1]) {
				t.Errorf("%s: tips.%s code span %q is not an operator the search parser understands", file, c.key, m[1])
			}
		}
	}
}

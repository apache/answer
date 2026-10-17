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
	"encoding/json"
	"fmt"
	"slices"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/plugin"
	"xorm.io/xorm"
)

func fixBranding(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState) (fixResult, error) {
	siteInfo := &entity.SiteInfo{Type: constant.SiteTypeBranding}
	exist, err := x.Get(siteInfo)
	if err != nil {
		return fixResult{}, fmt.Errorf("get site info failed: %w", err)
	}
	if !exist {
		fmt.Printf("[%s] no branding row found, skipping\n", fixTypeBranding)
		return fixResult{}, nil
	}

	content := &schema.SiteBrandingReq{}
	if err = json.Unmarshal([]byte(siteInfo.Content), content); err != nil {
		return fixResult{}, fmt.Errorf("parse branding content failed: %w", err)
	}

	type brandingField struct {
		name  string
		value *string
	}

	fields := []brandingField{
		{name: "logo", value: &content.Logo},
		{name: "mobile_logo", value: &content.MobileLogo},
		{name: "square_icon", value: &content.SquareIcon},
		{name: "favicon", value: &content.Favicon},
	}

	result := fixResult{}
	changed := false

	for _, field := range fields {
		if state.exhausted() {
			break
		}
		if len(*field.value) == 0 {
			continue
		}
		if err := applyValueFix(&result, opts, state, fixTypeBranding, field.name, *field.value, func(newValue string) error {
			*field.value = newValue
			changed = true
			return nil
		}); err != nil {
			return fixResult{}, err
		}
	}

	if changed {
		newContent, err := json.Marshal(content)
		if err != nil {
			return fixResult{}, fmt.Errorf("marshal branding content failed: %w", err)
		}
		_, err = x.ID(siteInfo.ID).Cols("content").NoAutoTime().Update(&entity.SiteInfo{Content: string(newContent)})
		if err != nil {
			return fixResult{}, fmt.Errorf("update site info failed: %w", err)
		}
	}

	records, err := newFileRecordSync(x, opts, fixTypeBranding, string(plugin.AdminBranding))
	if err != nil {
		return fixResult{}, err
	}
	for _, field := range fields {
		// --limit may leave another field on the shared old URL.
		oldURL := swapPrefix(*field.value, opts.DstPrefix, opts.SrcPrefix)
		if slices.ContainsFunc(fields, func(f brandingField) bool { return *f.value == oldURL }) {
			continue
		}
		if err := records.sync(*field.value); err != nil {
			return fixResult{}, err
		}
	}

	return result, nil
}

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

package schema

import (
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

type GetUserPluginListResp struct {
	Name     string `json:"name"`
	SlugName string `json:"slug_name"`
}

type UpdateUserPluginReq struct {
	PluginSlugName string `validate:"required,gt=1,lte=100" json:"plugin_slug_name"`
	UserID         string `json:"-"`
}

type GetUserPluginConfigReq struct {
	PluginSlugName string `validate:"required,gt=1,lte=100" form:"plugin_slug_name"`
	UserID         string `json:"-"`
}

type GetUserPluginConfigResp struct {
	Name         string         `json:"name"`
	SlugName     string         `json:"slug_name"`
	ConfigFields []*ConfigField `json:"config_fields"`
}

func (g *GetUserPluginConfigResp) SetConfigFields(ctx *gin.Context, fields []plugin.ConfigField) {
	for _, field := range fields {
		configField := &ConfigField{
			Name:        field.Name,
			Type:        string(field.Type),
			Title:       field.Title.Translate(ctx),
			Description: field.Description.Translate(ctx),
			Required:    field.Required,
			Value:       field.Value,
			UIOptions: ConfigFieldUIOptions{
				Rows:           field.UIOptions.Rows,
				InputType:      string(field.UIOptions.InputType),
				Variant:        field.UIOptions.Variant,
				ClassName:      field.UIOptions.ClassName,
				FieldClassName: field.UIOptions.FieldClassName,
			},
		}
		configField.UIOptions.Placeholder = field.UIOptions.Placeholder.Translate(ctx)
		configField.UIOptions.Label = field.UIOptions.Label.Translate(ctx)
		configField.UIOptions.Text = field.UIOptions.Text.Translate(ctx)
		if field.UIOptions.Action != nil {
			uiOptionAction := &UIOptionAction{
				Url:    field.UIOptions.Action.Url,
				Method: field.UIOptions.Action.Method,
			}
			if field.UIOptions.Action.Loading != nil {
				uiOptionAction.Loading = &LoadingAction{
					Text:  field.UIOptions.Action.Loading.Text.Translate(ctx),
					State: string(field.UIOptions.Action.Loading.State),
				}
			}
			if field.UIOptions.Action.OnComplete != nil {
				uiOptionAction.OnCompleteAction = &OnCompleteAction{
					ToastReturnMessage: field.UIOptions.Action.OnComplete.ToastReturnMessage,
					RefreshFormConfig:  field.UIOptions.Action.OnComplete.RefreshFormConfig,
				}
			}
			configField.UIOptions.Action = uiOptionAction
		}

		for _, option := range field.Options {
			configField.Options = append(configField.Options, ConfigFieldOption{
				Label: option.Label.Translate(ctx),
				Value: option.Value,
			})
		}
		g.ConfigFields = append(g.ConfigFields, configField)
	}
}

type UpdateUserPluginConfigReq struct {
	PluginSlugName string         `validate:"required,gt=1,lte=100" json:"plugin_slug_name"`
	ConfigFields   map[string]any `json:"config_fields"`
	UserID         string         `json:"-"`
}

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

const (
	PluginStatusActive   PluginStatus = "active"
	PluginStatusInactive PluginStatus = "inactive"
)

type PluginStatus string

type GetPluginListReq struct {
	Status     PluginStatus `form:"status"`
	HaveConfig bool         `form:"have_config"`
}

type GetPluginListResp struct {
	Name        string `json:"name"`
	SlugName    string `json:"slug_name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Enabled     bool   `json:"enabled"`
	HaveConfig  bool   `json:"have_config"`
	Link        string `json:"link"`
}

type GetAllPluginStatusResp struct {
	SlugName string `json:"slug_name"`
	Enabled  bool   `json:"enabled"`
}

type UpdatePluginStatusReq struct {
	PluginSlugName string `validate:"required,gt=1,lte=100" json:"plugin_slug_name"`
	Enabled        bool   `json:"enabled"`
}

type GetPluginConfigReq struct {
	PluginSlugName string `validate:"required,gt=1,lte=100" form:"plugin_slug_name"`
}

type GetPluginConfigResp struct {
	Name         string        `json:"name"`
	SlugName     string        `json:"slug_name"`
	Description  string        `json:"description"`
	Version      string        `json:"version"`
	ConfigFields []ConfigField `json:"config_fields"`
}

func (g *GetPluginConfigResp) SetConfigFields(ctx *gin.Context, fields []plugin.ConfigField) {
	for _, field := range fields {
		configField := ConfigField{
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

type ConfigField struct {
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Value       any                  `json:"value"`
	UIOptions   ConfigFieldUIOptions `json:"ui_options"`
	Options     []ConfigFieldOption  `json:"options,omitempty"`
}

type ConfigFieldUIOptions struct {
	Placeholder    string          `json:"placeholder,omitempty"`
	Rows           string          `json:"rows,omitempty"`
	InputType      string          `json:"input_type,omitempty"`
	Label          string          `json:"label,omitempty"`
	Action         *UIOptionAction `json:"action,omitempty"`
	Variant        string          `json:"variant,omitempty"`
	Text           string          `json:"text,omitempty"`
	ClassName      string          `json:"class_name,omitempty"`
	FieldClassName string          `json:"field_class_name,omitempty"`
}

type ConfigFieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type UIOptionAction struct {
	Url              string            `json:"url"`
	Method           string            `json:"method,omitempty"`
	Loading          *LoadingAction    `json:"loading,omitempty"`
	OnCompleteAction *OnCompleteAction `json:"on_complete,omitempty"`
}

type LoadingAction struct {
	Text  string `json:"text"`
	State string `json:"state"`
}

type OnCompleteAction struct {
	ToastReturnMessage bool `json:"toast_return_message"`
	RefreshFormConfig  bool `json:"refresh_form_config"`
}

type UpdatePluginConfigReq struct {
	PluginSlugName string         `validate:"required,gt=1,lte=100" json:"plugin_slug_name"`
	ConfigFields   map[string]any `json:"config_fields"`
}

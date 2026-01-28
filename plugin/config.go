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

package plugin

type ConfigType string
type InputType string

const (
	ConfigTypeInput       ConfigType = "input"
	ConfigTypeTextarea    ConfigType = "textarea"
	ConfigTypeCheckbox    ConfigType = "checkbox"
	ConfigTypeRadio       ConfigType = "radio"
	ConfigTypeSelect      ConfigType = "select"
	ConfigTypeUpload      ConfigType = "upload"
	ConfigTypeTimezone    ConfigType = "timezone"
	ConfigTypeSwitch      ConfigType = "switch"
	ConfigTypeButton      ConfigType = "button"
	ConfigTypeLegend      ConfigType = "legend"
	ConfigTypeTagSelector ConfigType = "tag_selector"
)

const (
	InputTypeText     InputType = "text"
	InputTypeColor    InputType = "color"
	InputTypeDate     InputType = "date"
	InputTypeDatetime InputType = "datetime-local"
	InputTypeEmail    InputType = "email"
	InputTypeMonth    InputType = "month"
	InputTypeNumber   InputType = "number"
	InputTypePassword InputType = "password"
	InputTypeRange    InputType = "range"
	InputTypeSearch   InputType = "search"
	InputTypeTel      InputType = "tel"
	InputTypeTime     InputType = "time"
	InputTypeUrl      InputType = "url"
	InputTypeWeek     InputType = "week"
)

type ConfigField struct {
	Name        string               `json:"name"`
	Type        ConfigType           `json:"type"`
	Title       Translator           `json:"title"`
	Description Translator           `json:"description"`
	Required    bool                 `json:"required"`
	Value       any                  `json:"value"`
	UIOptions   ConfigFieldUIOptions `json:"ui_options"`
	Options     []ConfigFieldOption  `json:"options,omitempty"`
}

type ConfigFieldUIOptions struct {
	Placeholder    Translator      `json:"placeholder"`
	Rows           string          `json:"rows,omitempty"`
	InputType      InputType       `json:"input_type,omitempty"`
	Label          Translator      `json:"label"`
	Action         *UIOptionAction `json:"action,omitempty"`
	Variant        string          `json:"variant,omitempty"`
	Text           Translator      `json:"text"`
	ClassName      string          `json:"class_name,omitempty"`
	FieldClassName string          `json:"field_class_name,omitempty"`
}

type ConfigFieldOption struct {
	Label Translator `json:"label"`
	Value string     `json:"value"`
}

type UIOptionAction struct {
	Url        string            `json:"url"`
	Method     string            `json:"method,omitempty"`
	Loading    *LoadingAction    `json:"loading,omitempty"`
	OnComplete *OnCompleteAction `json:"on_complete,omitempty"`
}

const (
	LoadingActionStateNone     LoadingActionType = "none"
	LoadingActionStatePending  LoadingActionType = "pending"
	LoadingActionStateComplete LoadingActionType = "completed"
)

type LoadingActionType string

type LoadingAction struct {
	Text  Translator        `json:"text"`
	State LoadingActionType `json:"state"`
}

type OnCompleteAction struct {
	ToastReturnMessage bool `json:"toast_return_message"`
	RefreshFormConfig  bool `json:"refresh_form_config"`
}

// TagSelectorOption represents a tag option in the tag selector config value field
type TagSelectorOption struct {
	TagID       string `json:"tag_id"`
	SlugName    string `json:"slug_name"`
	DisplayName string `json:"display_name"`
	Recommend   bool   `json:"recommend"`
	Reserved    bool   `json:"reserved"`
}

type Config interface {
	Base

	// ConfigFields returns the list of config fields
	ConfigFields() []ConfigField

	// ConfigReceiver receives the config data, it calls when the config is saved or initialized.
	// We recommend to unmarshal the data to a struct, and then use the struct to do something.
	// The config is encoded in JSON format.
	// It depends on the definition of ConfigFields.
	ConfigReceiver(config []byte) error
}

var (
	// CallConfig is a function that calls all registered config plugins
	CallConfig,
	registerConfig = MakePlugin[Config](true)
)

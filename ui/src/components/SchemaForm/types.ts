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

import { ButtonProps } from 'react-bootstrap';
import React from 'react';

import classnames from 'classnames';

import * as Type from '@/common/interface';

export interface FormProps {
  schema: JSONSchema | null;
  uiSchema?: UISchema;
  formData: Type.FormDataType | null;
  refreshConfig?: FormKit['refreshConfig'];
  hiddenSubmit?: boolean;
  onChange?: (data: Type.FormDataType) => void;
  onSubmit?: (e: React.FormEvent) => void;
}

export interface FormRef {
  validator: () => Promise<boolean>;
}

export interface JSONSchema {
  title: string;
  description?: string;
  required?: string[];
  properties: {
    [key: string]: {
      type?: 'string' | 'boolean' | 'number' | Type.Tag[];
      title: string;
      description?: string;
      enum?: Array<string | boolean | number>;
      enumNames?: string[];
      min?: number;
      max?: number;
      default?: string | boolean | number | any[];
      max_length?: number;
    };
  };
}

export interface BaseUIOptions {
  empty?: string;
  // Will be appended to the className of the form component itself
  className?: classnames.Argument;
  class_name?: classnames.Argument;
  // The className that will be attached to a **form field container**
  field_class_name?: classnames.Argument;
  // Make a form component render into simplified mode
  readOnly?: boolean;
  simplify?: boolean;
  validator?: (
    value,
    formData?,
  ) => Promise<string | true | void> | true | string;
}

export interface InputOptions extends BaseUIOptions {
  placeholder?: string;
  inputType?:
    | 'color'
    | 'date'
    | 'datetime-local'
    | 'email'
    | 'month'
    | 'number'
    | 'password'
    | 'range'
    | 'search'
    | 'tel'
    | 'text'
    | 'time'
    | 'url'
    | 'week';
  inputMode?:
    | 'text'
    | 'search'
    | 'none'
    | 'tel'
    | 'url'
    | 'email'
    | 'numeric'
    | 'decimal'
    | undefined;
}
export interface SelectOptions extends BaseUIOptions {}
export interface UploadOptions extends BaseUIOptions {
  acceptType?: string;
  imageType?: Type.UploadType;
}

export interface SwitchOptions extends BaseUIOptions {
  label?: string;
}

export interface TimezoneOptions extends BaseUIOptions {
  placeholder?: string;
}

export interface CheckboxOptions extends BaseUIOptions {}

export interface RadioOptions extends BaseUIOptions {}

export interface TextareaOptions extends BaseUIOptions {
  placeholder?: string;
  rows?: number;
}

export interface ButtonOptions extends BaseUIOptions {
  text: string;
  iconName?: string;
  action?: UIAction;
  actionType: 'click' | 'submit';
  variant?: ButtonProps['variant'];
  size?: ButtonProps['size'];
  title?: string;
  clickCallback?: () => void;
}

export interface InputGroupOptions extends InputOptions {
  prefixText?: string;
  suffixText?: string;
  prefixBtnOptions?: ButtonOptions;
  suffixBtnOptions?: ButtonOptions;
}

export type UIOptions =
  | InputOptions
  | SelectOptions
  | UploadOptions
  | SwitchOptions
  | TimezoneOptions
  | CheckboxOptions
  | RadioOptions
  | TextareaOptions
  | ButtonOptions
  | InputGroupOptions;

export type UIWidget =
  | 'textarea'
  | 'input'
  | 'checkbox'
  | 'radio'
  | 'select'
  | 'upload'
  | 'timezone'
  | 'switch'
  | 'legend'
  | 'button'
  | 'input_group'
  | 'tag_selector';
export interface UISchema {
  [key: string]: {
    'ui:widget'?: UIWidget;
    'ui:options'?: UIOptions;
  };
}

/**
 * A few notes on button control：
 *  - Mainly used to send a request and notify the result of the request, and to update the data as required
 *  - A scenario where a message notification is displayed directly after a click without sending a request, implementing a dedicated control
 *  - Scenarios where the page jumps directly after a click without sending a request, implementing a dedicated control
 *
 * @field url : Target address for sending requests
 * @field method : Method for sending requests, default `get`
 * @field callback: Button event handler function that will fully take over the button events when this field is configured
 *                 *** Incomplete, DO NOT USE ***
 * @field loading: Set button loading information
 * @field on_complete: What needs to be done when the `Action` completes
 * @field on_complete.toast_return_message: Does toast show the returned message
 * @field on_complete.refresh_form_config: Whether to refresh the form configuration (configuration only, no data included)
 */
export interface UIAction {
  url: string;
  method?: 'get' | 'post' | 'put' | 'delete';
  loading?: {
    text: string;
    state?: 'none' | 'pending' | 'completed';
  };
  on_complete?: {
    toast_return_message?: boolean;
    refresh_form_config?: boolean;
  };
}

/**
 * Form tools
 * - Used to get or set the configuration of forms and form items, the value of a form item
 * * @method refreshConfig(): void
 */

export interface FormKit {
  refreshConfig(): void;
}

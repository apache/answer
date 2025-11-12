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

import React, { FC } from 'react';
import { Form } from 'react-bootstrap';

import type * as Type from '@/common/interface';

interface Props {
  type: string | undefined;
  placeholder: string | undefined;
  fieldName: string;
  onChange?: (fd: Type.FormDataType) => void;
  formData: Type.FormDataType;
  readOnly: boolean;
  minValue?: number;
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
const Index: FC<Props> = ({
  type = 'text',
  placeholder = '',
  fieldName,
  onChange,
  formData,
  readOnly = false,
  minValue = 0,
  inputMode = 'text',
}) => {
  const fieldObject = formData[fieldName];
  const handleChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = evt.currentTarget;
    const state = {
      ...formData,
      [name]: {
        ...formData[name],
        value,
        isInvalid: false,
      },
    };
    if (typeof onChange === 'function') {
      onChange(state);
    }
  };

  return (
    <Form.Control
      name={fieldName}
      placeholder={placeholder}
      type={type}
      value={fieldObject?.value || ''}
      min={minValue}
      inputMode={inputMode}
      onChange={handleChange}
      disabled={readOnly}
      isInvalid={fieldObject?.isInvalid}
      style={type === 'color' ? { width: '100px', flex: 'none' } : {}}
    />
  );
};

export default Index;

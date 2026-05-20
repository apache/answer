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

import { FC } from 'react';

import { TagSelector } from '@/components';
import type * as Type from '@/common/interface';

interface Props {
  maxTagLength?: number;
  description?: string;
  fieldName: string;
  onChange?: (fd: Type.FormDataType) => void;
  formData: Type.FormDataType;
}
const Index: FC<Props> = ({
  description,
  maxTagLength,
  fieldName,
  onChange,
  formData,
}) => {
  const fieldObject = formData[fieldName];
  const handleChange = (data: Type.Tag[]) => {
    const state = {
      ...formData,
      [fieldName]: {
        ...formData[fieldName],
        value: data,
        isInvalid: false,
      },
    };
    if (typeof onChange === 'function') {
      onChange(state);
    }
  };

  return (
    <TagSelector
      value={fieldObject?.value || []}
      onChange={handleChange}
      maxTagLength={maxTagLength || 0}
      isInvalid={fieldObject?.isInvalid}
      formText={description}
      errMsg={fieldObject?.errorMsg}
    />
  );
};

export default Index;

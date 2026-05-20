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

import useSWR from 'swr';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const useQueryApiKeys = () => {
  const apiUrl = `/answer/admin/api/api-key/all`;
  const { data, error, mutate } = useSWR<Type.AdminApiKeysItem[], Error>(
    apiUrl,
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const addApiKey = (params: Type.AddOrEditApiKeyParams) => {
  return request.post('/answer/admin/api/api-key', params);
};

export const updateApiKey = (params: Type.AddOrEditApiKeyParams) => {
  return request.put('/answer/admin/api/api-key', params);
};

export const deleteApiKey = (id: string) => {
  return request.delete('/answer/admin/api/api-key', {
    id,
  });
};

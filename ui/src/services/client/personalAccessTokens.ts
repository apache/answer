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

import type * as Type from '@/common/interface';
import request from '@/utils/request';

const endpoint = '/answer/api/v1/personal-access-tokens';

export const usePersonalAccessTokens = () => {
  const { data, error, mutate } = useSWR<Type.PersonalAccessTokenInfo[]>(
    endpoint,
    request.instance.get,
  );
  return { data, error, mutate, isLoading: !data && !error };
};

export const createPersonalAccessToken = (
  params: Type.CreatePersonalAccessTokenParams,
) => request.post<Type.CreatePersonalAccessTokenResp>(endpoint, params);

export const revokePersonalAccessToken = (id: number) =>
  request.delete(endpoint, { id });

export const reauthenticate = (
  params: { password: string } & Type.ImgCodeReq,
) => request.post('/answer/api/v1/user/reauthenticate', params);

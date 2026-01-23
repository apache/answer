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
import qs from 'qs';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const getAiConfig = () => {
  return request.get<Type.AiConfig>('/answer/admin/api/ai-config');
};

export const useQueryAiProvider = () => {
  const apiUrl = `/answer/admin/api/ai-provider`;
  const { data, error, mutate } = useSWR<Type.AiProviderItem[], Error>(
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

export const checkAiConfig = (params) => {
  return request.post('/answer/admin/api/ai-models', params);
};

export const saveAiConfig = (params) => {
  return request.put('/answer/admin/api/ai-config', params);
};

export const useQueryAdminConversationDetail = (id: string) => {
  const apiUrl = !id
    ? null
    : `/answer/admin/api/ai/conversation?conversation_id=${id}`;

  const { data, error, mutate } = useSWR<Type.ConversationDetail, Error>(
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

export const useQueryAdminConversationList = (params: Type.Paging) => {
  const apiUrl = `/answer/admin/api/ai/conversation/page?${qs.stringify(params)}`;
  const { data, error, mutate } = useSWR<
    { count: number; list: Type.AdminConversationListItem[] },
    Error
  >(apiUrl, request.instance.get);
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const deleteAdminConversation = (id: string) => {
  return request.delete('/answer/admin/api/ai/conversation', {
    conversation_id: id,
  });
};

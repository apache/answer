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

import qs from 'qs';
import useSWR from 'swr';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

// messages from admins / moderators to users
export interface AdminMessageUser {
  id: string;
  username: string;
  display_name: string;
  avatar?: string;
  email?: string;
  status?: string;
}

export interface AdminMessageItem {
  id: string;
  created_at: number;
  read_at: number;
  sender: AdminMessageUser;
  receiver: AdminMessageUser;
  title: string;
  body: string;
  object_type?: string;
  object_id?: string;
  question_id?: string;
  answer_id?: string;
  object_title?: string;
  url_title?: string;
}

export interface AdminMessagePageParams {
  page?: number;
  page_size?: number;
  from?: number;
  to?: number;
  username?: string;
  q?: string;
}

export const sendAdminMessage = (params: {
  username: string;
  title: string;
  body: string;
  object_id?: string;
}) => {
  return request.post<AdminMessageItem>('/answer/api/v1/admin-message', params);
};

export const useQueryAdminMessages = (params: AdminMessagePageParams) => {
  const apiUrl = `/answer/api/v1/admin-message/page?${qs.stringify(params, {
    skipNulls: true,
    filter: (_, v) => (v === '' ? undefined : v),
  })}`;
  const { data, error, mutate } = useSWR<
    Type.ListResult<AdminMessageItem>,
    Error
  >(apiUrl, request.instance.get);
  return { data, isLoading: !data && !error, error, mutate };
};

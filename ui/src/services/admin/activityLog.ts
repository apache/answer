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

import axios from 'axios';
import qs from 'qs';
import useSWR from 'swr';

import request from '@/utils/request';
import Storage from '@/utils/storage';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import type * as Type from '@/common/interface';

// community activity log
export interface ActivityLogUser {
  id: string;
  username: string;
  display_name: string;
  avatar?: string;
  status?: string;
}

export interface ActivityLogItem {
  id: number;
  created_at: number;
  action: string;
  user: ActivityLogUser;
  target?: ActivityLogUser;
  object_type: string;
  object_id: string;
  question_id?: string;
  answer_id?: string;
  title?: string;
  url_title?: string;
  rank_delta: number;
  detail: Record<string, any>;
  ip?: string;
}

export interface ActivityLogActionCount {
  action: string;
  count: number;
}

export interface ActivityLogParams {
  page?: number;
  page_size?: number;
  from?: number;
  to?: number;
  username?: string;
  action?: string;
  q?: string;
}

const query = (params: ActivityLogParams) =>
  qs.stringify(params, {
    skipNulls: true,
    filter: (_, v) => (v === '' ? undefined : v),
  });

export const useQueryActivityLog = (params: ActivityLogParams) => {
  const apiUrl = `/answer/admin/api/activity-log?${query(params)}`;
  const { data, error, mutate } = useSWR<
    Type.ListResult<ActivityLogItem>,
    Error
  >(apiUrl, request.instance.get);
  return { data, isLoading: !data && !error, error, mutate };
};

export const useQueryActivityLogActions = (params: ActivityLogParams) => {
  const apiUrl = `/answer/admin/api/activity-log/actions?${query(params)}`;
  const { data, error } = useSWR<ActivityLogActionCount[], Error>(
    apiUrl,
    request.instance.get,
  );
  return { data, isLoading: !data && !error, error };
};

export interface ActivityLogDailyRow {
  date: string;
  from: number;
  to: number;
  questions: number;
  answers: number;
  comments: number;
  reviews: number;
  badges: number;
  views: number;
  logins: number;
  actions: Record<string, number>;
}

export const useQueryActivityLogDaily = (days = 14) => {
  const tzOffset = -new Date().getTimezoneOffset();
  const apiUrl = `/answer/admin/api/activity-log/daily?${qs.stringify({ days, tz_offset: tzOffset })}`;
  const { data, error } = useSWR<ActivityLogDailyRow[], Error>(
    apiUrl,
    request.instance.get,
  );
  return { data, isLoading: !data && !error, error };
};

// the export needs the Authorization header and a raw (non-JSON) body, so bypass the
// JSON-unwrapping interceptor and hand the blob to the browser as a download
export const exportActivityLog = async (params: ActivityLogParams) => {
  const res = await axios.get(
    `/answer/admin/api/activity-log/export?${query(params)}`,
    {
      responseType: 'blob',
      headers: { Authorization: Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '' },
    },
  );
  const disposition = res?.headers?.['content-disposition'] || '';
  const match = /filename="?([^";]+)"?/.exec(disposition);
  const filename = match ? match[1] : 'log-aktywnosci.tsv';
  const blob = res.data instanceof Blob ? res.data : new Blob([res.data]);
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
};

export const postPageView = (params: { path: string; title?: string }) => {
  return request.post('/answer/api/v1/activity-log/view', params, {
    ignoreError: '50X',
  } as any);
};

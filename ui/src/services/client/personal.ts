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

export const usePersonalInfoByName = (username: string) => {
  const apiUrl = '/answer/api/v1/personal/user/info';
  const { data, error, mutate } = useSWR<Type.UserInfoRes, Error>(
    username ? `${apiUrl}?username=${username}` : null,
    (url) =>
      request.get(url, {
        allow404: true,
      }),
  );
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

interface ListReq {
  username?: string;
  page: number;
  page_size: number;
  order?: string;
}

interface ListRes {
  count: number;
  list: any[];
}

export const usePersonalTop = (username: string, tabName: string) => {
  const apiUrl = '/answer/api/v1/personal/content/top?username=';
  const { data, error } = useSWR<{ comments: any[]; posts: any[] }, Error>(
    tabName === 'overview' ? `${apiUrl}${username}` : null,
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const usePersonalListByTabName = (params: ListReq, tabName: string) => {
  let apiUrl: string | null = '';
  if (tabName === 'comments') {
    apiUrl = '/answer/api/v1/personal/post/comment/page';
  }
  if (tabName === 'posts') {
    apiUrl = '/answer/api/v1/personal/post/page';
  }
  if (tabName === 'bookmarks') {
    delete params.order;
    apiUrl = '/answer/api/v1/personal/collection/page';
  }
  if (tabName === 'replies') {
    delete params.order;
    apiUrl = '/answer/api/v1/personal/reply/page';
  }
  if (tabName === 'votes') {
    delete params.username;
    apiUrl = '/answer/api/v1/personal/vote/page';
  }
  if (tabName === 'badges') {
    delete params.order;
    apiUrl = '/answer/api/v1/badge/user/awards';
  }

  const queryParams = qs.stringify(params, { skipNulls: true });
  const { data, error, mutate } = useSWR<ListRes, Error>(
    tabName !== 'overview' ? `${apiUrl}?${queryParams}` : null,
    request.instance.get,
  );

  return {
    data: {
      [tabName]: data,
    },
    isLoading: !data && !error,
    error,
    mutate,
  };
};

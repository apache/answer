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
import type { PluginConfig } from '@/services/admin/plugins';

export const getLanguageConfig = () => {
  return request.get('/answer/api/v1/language/config');
};

export const getLanguageOptions = () => {
  return request.get<Type.LangsType[]>('/answer/api/v1/language/options');
};

interface userSettingInterface {
  language: '';
  color_scheme: '';
}
export const updateUserInterface = (data: userSettingInterface) => {
  return request.put('/answer/api/v1/user/interface', data);
};

export const useGetNotificationConfig = () => {
  return useSWR<Type.NotificationConfig>(
    '/answer/api/v1/user/notification/config',
    request.instance.get,
  );
};

export const putNotificationConfig = (data: Type.NotificationConfig) => {
  return request.put('/answer/api/v1/user/notification/config', data);
};

export const useGetUserPluginList = () => {
  return useSWR<Type.UserPluginsConfigRes[]>(
    '/answer/api/v1/user/plugin/configs',
    request.instance.get,
  );
};

export const useGetUserPluginConfig = (params) => {
  const apiUrl = `/answer/api/v1/user/plugin/config?${qs.stringify(params)}`;
  const { data, error, mutate } = useSWR<PluginConfig, Error>(
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

export const updateUserPluginConfig = (params) => {
  return request.put('/answer/api/v1/user/plugin/config', params);
};

export interface AnonymityConfig {
  enabled: boolean;
}

export const useGetAnonymityConfig = () => {
  return useSWR<AnonymityConfig>(
    '/answer/api/v1/user/anonymity/config',
    request.instance.get,
  );
};

export const putAnonymityConfig = (data: AnonymityConfig) => {
  return request.put('/answer/api/v1/user/anonymity/config', data);
};

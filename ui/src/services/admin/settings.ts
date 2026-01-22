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

export interface AdminSettingsUsers {
  allow_update_avatar: boolean;
  allow_update_bio: boolean;
  allow_update_display_name: boolean;
  allow_update_location: boolean;
  allow_update_username: boolean;
  allow_update_website: boolean;
}

interface PrivilegeLevel {
  level: number;
  level_desc: string;
  privileges: {
    label: string;
    value: number;
    key: string;
  }[];
}
export interface AdminSettingsPrivilege {
  selected_level: number;
  options: PrivilegeLevel[];
}

export interface AdminSettingsPrivilegeReq {
  level: number;
  custom_privileges?: {
    label?: string;
    value: number;
    key: string;
  }[];
}

export const useGeneralSetting = () => {
  const apiUrl = `/answer/admin/api/siteinfo/general`;
  const { data, error } = useSWR<Type.AdminSettingsGeneral, Error>(
    [apiUrl],
    request.instance.get,
  );

  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateGeneralSetting = (params: Type.AdminSettingsGeneral) => {
  const apiUrl = `/answer/admin/api/siteinfo/general`;
  return request.put(apiUrl, params);
};

export const useInterfaceSetting = () => {
  const apiUrl = `/answer/admin/api/siteinfo/interface`;
  const { data, error } = useSWR<Type.AdminSettingsInterface, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateInterfaceSetting = (params: Type.AdminSettingsInterface) => {
  const apiUrl = `/answer/admin/api/siteinfo/interface`;
  return request.put(apiUrl, params);
};

export const useSmtpSetting = () => {
  const apiUrl = `/answer/admin/api/setting/smtp`;
  const { data, error } = useSWR<Type.AdminSettingsSmtp, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateSmtpSetting = (params: Type.AdminSettingsSmtp) => {
  const apiUrl = `/answer/admin/api/setting/smtp`;
  return request.put(apiUrl, params);
};

export const getAdminLanguageOptions = () => {
  const apiUrl = `/answer/admin/api/language/options`;
  return request.get<Type.LangsType[]>(apiUrl);
};

export const getBrandSetting = () => {
  return request.get('/answer/admin/api/siteinfo/branding');
};

export const brandSetting = (params: Type.AdminSettingBranding) => {
  return request.put('/answer/admin/api/siteinfo/branding', params);
};

export const getAdminFilesSetting = () => {
  return request.get('/answer/admin/api/siteinfo/advanced');
};

export const updateAdminFilesSetting = (params: Type.AdminSettingsWrite) => {
  return request.put('/answer/admin/api/siteinfo/advanced', params);
};

export const getSeoSetting = () => {
  return request.get<Type.AdminSettingsSeo>('/answer/admin/api/siteinfo/seo');
};

export const putSeoSetting = (params: Type.AdminSettingsSeo) => {
  return request.put('/answer/admin/api/siteinfo/seo', params);
};

export const getThemeSetting = () => {
  return request.get<Type.AdminSettingsTheme>(
    '/answer/admin/api/siteinfo/theme',
  );
};

export const putThemeSetting = (params: Type.AdminSettingsTheme) => {
  return request.put('/answer/admin/api/siteinfo/theme', params);
};

export const getPageCustom = () => {
  return request.get<Type.AdminSettingsCustom>(
    '/answer/admin/api/siteinfo/custom-css-html',
  );
};

export const putPageCustom = (params: Type.AdminSettingsCustom) => {
  return request.put('/answer/admin/api/siteinfo/custom-css-html', params);
};

export const getLoginSetting = () => {
  return request.get<Type.AdminSettingsLogin>(
    '/answer/admin/api/siteinfo/login',
  );
};

export const putLoginSetting = (params: Type.AdminSettingsLogin) => {
  return request.put('/answer/admin/api/siteinfo/login', params);
};

export const getUsersSetting = () => {
  return request.get<AdminSettingsUsers>('/answer/admin/api/siteinfo/users');
};

export const putUsersSetting = (params: AdminSettingsUsers) => {
  return request.put('/answer/admin/api/siteinfo/users', params);
};

export const getPrivilegeSetting = () => {
  return request.get<AdminSettingsPrivilege>(
    '/answer/admin/api/setting/privileges',
  );
};

export const putPrivilegeSetting = (params: AdminSettingsPrivilegeReq) => {
  return request.put('/answer/admin/api/setting/privileges', params);
};

export const getPoliciesSetting = () => {
  return request.get<Type.AdminSettingsLegal>(
    '/answer/admin/api/siteinfo/polices',
  );
};

export const putPoliciesSetting = (params: Type.AdminSettingsLegal) => {
  return request.put('/answer/admin/api/siteinfo/polices', params);
};

export const getSecuritySetting = () => {
  return request.get<Type.AdminSettingsSecurity>(
    '/answer/admin/api/siteinfo/security',
  );
};

export const putSecuritySetting = (params: Type.AdminSettingsSecurity) => {
  return request.put('/answer/admin/api/siteinfo/security', params);
};

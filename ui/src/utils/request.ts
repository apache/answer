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

import axios, { AxiosResponse } from 'axios';
import type {
  AxiosInstance,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
  AxiosError,
} from 'axios';

import { Modal } from '@/components';
import { loggedUserInfoStore, toastStore, errorCodeStore } from '@/stores';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import { RouteAlias } from '@/router/alias';
import { getCurrentLang } from '@/utils/localize';

import Storage from './storage';
import { floppyNavigation } from './floppyNavigation';
import { isIgnoredPath, IGNORE_PATH_LIST } from './guard';

const baseConfig = {
  baseURL:
    process.env.NODE_ENV === 'development' ? '' : process.env.REACT_APP_API_URL,
  timeout: 10000,
  withCredentials: true,
};

interface ApiConfig extends AxiosRequestConfig {
  // Configure whether to allow takeover of 404 errors
  allow404?: boolean;
  ignoreError?: '403' | '50X';
  // Configure whether to pass errors directly
  passingError?: boolean;
}

class Request {
  instance: AxiosInstance;

  constructor(config: AxiosRequestConfig) {
    this.instance = axios.create(config);
    this.instance.interceptors.request.use(
      (requestConfig: InternalAxiosRequestConfig) => {
        const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
        const lang = getCurrentLang();
        requestConfig.headers.set('Authorization', token);
        requestConfig.headers.set('Accept-Language', lang);
        return requestConfig;
      },
      (err: AxiosError) => {
        console.error('request interceptors error:', err);
      },
    );

    this.instance.interceptors.response.use(
      (res: AxiosResponse) => {
        const { status, data } = res.data;

        if (status === 204) {
          // no content
          return true;
        }
        return data;
      },
      (error) => {
        const {
          status,
          data: errBody,
          config: errConfig,
        } = error.response || {};
        const { data = {}, msg = '' } = errBody || {};

        const errorObject: {
          code: any;
          msg: string;
          data: any;
          // Currently only used for form errors
          isError?: boolean;
          // Currently only used for form errors
          list?: any[];
        } = {
          code: status,
          msg,
          data,
        };

        if (status === 400) {
          if (data?.err_type && errConfig?.passingError) {
            return Promise.reject(errorObject);
          }
          if (data?.err_type) {
            if (data.err_type === 'toast') {
              // toast error message
              toastStore.getState().show({
                msg,
                variant: 'danger',
              });
            }

            if (data.err_type === 'alert') {
              return Promise.reject({
                msg,
                ...data,
              });
            }

            if (data.err_type === 'modal') {
              // modal error message
              Modal.confirm({
                content: msg,
              });
            }

            return Promise.reject(false);
          }

          if (data instanceof Array && data.length > 0) {
            // handle form error
            errorObject.isError = true;
            errorObject.list = data;
            return Promise.reject(errorObject);
          }

          if (!data || Object.keys(data).length <= 0) {
            // default error msg will show modal
            Modal.confirm({
              content: msg,
              showConfirm: false,
              cancelText: 'close',
            });
            return Promise.reject(false);
          }
        }
        // 401: Re-login required
        if (status === 401) {
          // clear userinfo
          errorCodeStore.getState().reset();
          loggedUserInfoStore.getState().clear();
          floppyNavigation.navigateToLogin();
          return Promise.reject(false);
        }

        if (status === 403) {
          // Permission interception
          if (data?.type === 'url_expired') {
            // url expired
            floppyNavigation.navigate(RouteAlias.activationFailed, {
              handler: 'replace',
            });
            return Promise.reject(false);
          }
          if (data?.type === 'inactive') {
            // inactivated
            floppyNavigation.navigate(RouteAlias.inactive);
            return Promise.reject(false);
          }

          if (data?.type === 'suspended') {
            loggedUserInfoStore.getState().clear();
            floppyNavigation.navigate(RouteAlias.suspended, {
              handler: 'replace',
            });
            return Promise.reject(false);
          }

          if (isIgnoredPath(IGNORE_PATH_LIST)) {
            return Promise.reject(false);
          }
          if (error.config?.url.includes('/admin/api')) {
            errorCodeStore.getState().update('403');
            return Promise.reject(false);
          }

          if (msg) {
            toastStore.getState().show({
              msg,
              variant: 'danger',
            });
          }
          return Promise.reject(false);
        }

        if (status === 404 && error.config?.allow404) {
          if (isIgnoredPath(IGNORE_PATH_LIST)) {
            return Promise.reject(false);
          }
          errorCodeStore.getState().update('404');
          return Promise.reject(false);
        }

        if (status >= 500) {
          if (isIgnoredPath(IGNORE_PATH_LIST)) {
            return Promise.reject(false);
          }

          if (error.config?.ignoreError !== '50X') {
            errorCodeStore.getState().update('50X');
          }

          console.error(
            `Request failed with status code ${status}, ${msg || ''}`,
          );
        }
        return Promise.reject(errorObject);
      },
    );
  }

  public request(config: AxiosRequestConfig): Promise<AxiosResponse> {
    return this.instance.request(config);
  }

  public get<T = any>(url: string, config?: ApiConfig): Promise<T> {
    return this.instance.get(url, config);
  }

  public post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.instance.post(url, data, config);
  }

  public put<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.instance.put(url, data, config);
  }

  public delete<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    return this.instance.delete(url, {
      data,
      ...config,
    });
  }
}

export default new Request(baseConfig);

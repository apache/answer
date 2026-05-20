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

import { Modal } from '@/components';
import { loggedUserInfoStore, toastStore, errorCodeStore } from '@/stores';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import { RouteAlias } from '@/router/alias';
import { getCurrentLang } from '@/utils/localize';
import Storage from '@/utils/storage';
import { floppyNavigation } from '@/utils/floppyNavigation';
import { isIgnoredPath, IGNORE_PATH_LIST } from '@/utils/guard';

interface RequestAiOptions extends RequestInit {
  onMessage?: (text: any) => void;
  onError?: (error: Error) => void;
  onComplete?: () => void;
  signal?: AbortSignal;
  // 添加项目配置选项
  allow404?: boolean;
  ignoreError?: '403' | '50X';
  passingError?: boolean;
}

// create a object to track the current request state
const requestState = {
  currentReader: null as ReadableStreamDefaultReader<Uint8Array> | null,
  abortController: null as AbortController | null,
  isProcessing: false,
};

// HTTP error handling function (based on request.ts logic)
const handleHttpError = async (
  response: Response,
  options: RequestAiOptions,
): Promise<void> => {
  const { status } = response;
  let errBody: any = {};

  try {
    const text = await response.text();
    errBody = text ? JSON.parse(text) : {};
  } catch {
    errBody = { msg: response.statusText };
  }

  const { data = {}, msg = '', config } = errBody || {};

  const errorObject = {
    code: status,
    msg,
    data,
  };

  if (status === 400) {
    if (data?.err_type && options?.passingError) {
      return Promise.reject(errorObject);
    }

    if (data?.err_type) {
      if (data.err_type === 'toast') {
        toastStore.getState().show({
          msg,
          variant: 'danger',
        });
      }

      if (data.err_type === 'alert') {
        return Promise.reject({ msg, ...data });
      }

      if (data.err_type === 'modal') {
        Modal.confirm({
          content: msg,
        });
      }
      return Promise.reject(false);
    }

    if (Array.isArray(data) && data.length > 0) {
      return Promise.reject({
        ...errorObject,
        isError: true,
        list: data,
      });
    }

    if (!data || Object.keys(data).length <= 0) {
      Modal.confirm({
        content: msg,
        showConfirm: false,
        cancelText: 'close',
      });
      return Promise.reject(false);
    }
  }

  // 401: 重新登录
  if (status === 401) {
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
    if (config?.url.includes('/admin/api')) {
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

  if (status === 404 && config?.allow404) {
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

    if (config?.ignoreError !== '50X') {
      errorCodeStore.getState().update('50X');
    }

    console.error(`Request failed with status code ${status}, ${msg || ''}`);
  }
  return Promise.reject(errorObject);
};
const requestAi = async (url: string, options: RequestAiOptions) => {
  try {
    // if there is a previous request being processed, cancel it
    if (requestState.isProcessing && requestState.abortController) {
      requestState.abortController.abort();
    }

    // create a new AbortController
    const abortController = new AbortController();
    requestState.abortController = abortController;

    // merge the incoming signal with the new created signal
    const combinedSignal = options.signal || abortController.signal;

    // mark as being processed
    requestState.isProcessing = true;

    // get the authentication information and language settings (consistent with request.ts)
    const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
    console.log(token);
    const lang = getCurrentLang();

    const response = await fetch(url, {
      ...options,
      method: 'POST',
      signal: combinedSignal,
      headers: {
        Authorization: token,
        'Accept-Language': lang,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    // unified error handling (based on request.ts logic)
    if (!response.ok) {
      await handleHttpError(response, options);
      return;
    }

    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('ReadableStream not supported');
    }

    // store the current reader so it can be cancelled later
    requestState.currentReader = reader;

    const decoder = new TextDecoder();
    let buffer = '';

    const processStream = async (): Promise<void> => {
      try {
        const { value, done } = await reader.read();

        if (done) {
          options.onComplete?.();
          requestState.isProcessing = false;
          requestState.currentReader = null;
          return;
        }

        const chunk = decoder.decode(value, { stream: true });
        buffer += chunk;

        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        lines.forEach((line) => {
          if (line.trim()) {
            try {
              // handle the special [DONE] signal
              const cleanedLine = line.replace(/^data: /, '').trim();
              if (cleanedLine === '[DONE]') {
                return; // skip the [DONE] signal processing
              }

              if (cleanedLine) {
                const parsedLine = JSON.parse(cleanedLine);
                options.onMessage?.(parsedLine);
              }
            } catch (error) {
              console.debug('Error parsing line:', line);
            }
          }
        });

        // check if it has been cancelled
        if (combinedSignal.aborted) {
          requestState.isProcessing = false;
          requestState.currentReader = null;
          throw new Error('Request was aborted');
        }

        await processStream();
      } catch (error) {
        if ((error as Error).message === 'Request was aborted') {
          options.onComplete?.();
        } else {
          throw error; // rethrow other errors
        }
      }
    };

    await processStream();
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';

    // if the error is caused by cancellation, do not treat it as an error
    if (
      errorMessage !== 'The user aborted a request' &&
      errorMessage !== 'Request was aborted'
    ) {
      console.error('Request AI Error:', errorMessage);
      options.onError?.(new Error(errorMessage));
    } else {
      console.log('Request was cancelled by user');
      options.onComplete?.(); // cancellation is also considered complete
    }
  } finally {
    requestState.isProcessing = false;
    requestState.currentReader = null;
  }
};

// add a function to cancel the current request
const cancelCurrentRequest = () => {
  if (requestState.abortController) {
    requestState.abortController.abort();
    console.log('AI request cancelled by user');
    return true;
  }
  return false;
};

export { cancelCurrentRequest };
export default requestAi;

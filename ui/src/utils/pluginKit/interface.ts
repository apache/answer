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

import { NamedExoticComponent, FC, RefObject } from 'react';

import type * as Type from '@/common/interface';
import Request from '@/utils/request';

export enum PluginType {
  Connector = 'connector',
  Search = 'search',
  Editor = 'editor',
  EditorReplacement = 'editor_replacement',
  Route = 'route',
  Captcha = 'captcha',
  Render = 'render',
  Sidebar = 'sidebar',
}

export interface PluginInfo {
  slug_name: string;
  type: PluginType;
  name?: string;
  description?: string;
  route?: string;
  registrationMode?: 'multiple' | 'singleton';
}

export interface Plugin {
  info: PluginInfo;
  component: NamedExoticComponent | FC;
  i18nConfig?;
  hooks?: {
    useRender?: Array<
      (
        element: HTMLElement | RefObject<HTMLElement> | null,
        request?: typeof Request,
      ) => void
    >;
    useCaptcha?: (props: { captchaKey: Type.CaptchaKey; commonProps: any }) => {
      getCaptcha: () => Record<string, any>;
      check: (t: () => void) => void;
      handleCaptchaError: (error) => any;
      close: () => Promise<void>;
      resolveCaptchaReq: (data) => void;
    };
  };
  activated?: boolean;
}

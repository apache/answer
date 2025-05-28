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

import { create } from 'zustand';

import { AdminSettingsTheme } from '@/common/interface';
import { DEFAULT_THEME_COLOR } from '@/common/constants';

interface IType {
  theme: AdminSettingsTheme['theme'];
  theme_config: AdminSettingsTheme['theme_config'];
  theme_options: AdminSettingsTheme['theme_options'];
  color_scheme: AdminSettingsTheme['color_scheme'];
  update: (params: AdminSettingsTheme) => void;
}

const store = create<IType>((set) => ({
  theme: 'default',
  color_scheme: 'system',
  theme_options: [{ label: 'Default', value: 'default' }],
  theme_config: {
    default: {
      navbar_style: DEFAULT_THEME_COLOR,
      primary_color: DEFAULT_THEME_COLOR,
    },
  },
  update: (params) =>
    set((state) => {
      // Compatibility default value is colored or light before v1.5.1
      if (!params.theme_config.default.navbar_style.startsWith('#')) {
        params.theme_config.default.navbar_style = DEFAULT_THEME_COLOR;
      }
      return {
        ...state,
        ...params,
      };
    }),
}));

export default store;

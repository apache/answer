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

import { AdminSettingsInterface } from '@/common/interface';
import { DEFAULT_LANG } from '@/common/constants';

interface InterfaceType {
  interface: AdminSettingsInterface;
  update: (params: AdminSettingsInterface) => void;
}

const interfaceSetting = create<InterfaceType>((set) => ({
  interface: {
    language: DEFAULT_LANG,
    time_zone: '',
    default_avatar: 'system',
    gravatar_base_url: '',
  },
  update: (params) =>
    set((state) => {
      return {
        interface: { ...state.interface, ...params },
      };
    }),
}));

export default interfaceSetting;

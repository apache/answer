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

import { AdminSettingsSeo } from '@/common/interface';

interface IProps {
  seo: AdminSettingsSeo;
  update: (params: AdminSettingsSeo) => void;
}

const Index = create<IProps>((set) => ({
  seo: {
    robots: '',
    permalink: 1,
  },
  update: (params) =>
    set((state) => {
      const o = { ...state.seo, ...params };
      // @ts-ignore
      if (!/[1234]/.test(o.permalink)) {
        o.permalink = 1;
      }
      return {
        seo: o,
      };
    }),
}));

export default Index;

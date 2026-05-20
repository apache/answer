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

interface IType {
  custom_css: string;
  custom_head: string;
  custom_header: string;
  custom_footer: string;
  custom_sidebar: string;
  update: (params: {
    custom_css?: string;
    custom_head?: string;
    custom_header?: string;
    custom_footer?: string;
    custom_sidebar?: string;
  }) => void;
}

const loginSetting = create<IType>((set) => ({
  custom_css: '',
  custom_head: '',
  custom_header: '',
  custom_footer: '',
  custom_sidebar: '',
  update: (params) =>
    set((state) => {
      return {
        ...state,
        ...params,
      };
    }),
}));

export default loginSetting;

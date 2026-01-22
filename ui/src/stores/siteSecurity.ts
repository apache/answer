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

interface SecurityStore {
  login_required: boolean;
  check_update: boolean;
  external_content_display: string;
  update: (params: {
    external_content_display: string;
    check_update: boolean;
    login_required: boolean;
  }) => void;
}

const siteSecurityStore = create<SecurityStore>((set) => ({
  login_required: false,
  check_update: true,
  external_content_display: 'always_display',
  update: (params) =>
    set((state) => {
      return {
        ...state,
        ...params,
      };
    }),
}));

export default siteSecurityStore;

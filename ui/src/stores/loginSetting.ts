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

import { AdminSettingsLogin } from '@/common/interface';

interface IType {
  login: AdminSettingsLogin;
  update: (params: AdminSettingsLogin) => void;
}

const loginSetting = create<IType>((set) => ({
  login: {
    allow_new_registrations: true,
    allow_email_registrations: true,
    allow_email_domains: [],
    allow_password_login: true,
  },
  update: (params) =>
    set(() => {
      return {
        login: params,
      };
    }),
}));

export default loginSetting;

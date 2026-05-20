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

import type { UserInfoRes } from '@/common/interface';
import Storage from '@/utils/storage';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';

interface UserInfoStore {
  user: UserInfoRes;
  update: (params: UserInfoRes) => void;
  clear: (removeToken?: boolean) => void;
}

const initUser: UserInfoRes = {
  access_token: '',
  username: '',
  avatar: '',
  rank: 0,
  bio: '',
  bio_html: '',
  display_name: '',
  location: '',
  website: '',
  status: 'normal',
  mail_status: 1,
  language: 'Default',
  color_scheme: 'default',
  is_admin: false,
  have_password: true,
  role_id: 1,
};

const loggedUserInfo = create<UserInfoStore>((set) => ({
  user: initUser,
  update: (params) => {
    if (typeof params !== 'object' || !params) {
      return;
    }
    if (!params?.language) {
      params.language = 'Default';
    }
    if (!params?.color_scheme) {
      params.color_scheme = 'default';
    }
    set(() => {
      Storage.set(LOGGED_TOKEN_STORAGE_KEY, params.access_token);
      return { user: params };
    });
  },
  clear: (removeToken = true) =>
    set(() => {
      if (removeToken) {
        Storage.remove(LOGGED_TOKEN_STORAGE_KEY);
      }
      return { user: initUser };
    }),
}));

export default loggedUserInfo;

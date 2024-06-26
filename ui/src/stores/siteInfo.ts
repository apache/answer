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

import create from 'zustand';

import { AdminSettingsGeneral, AdminSettingsUsers } from '@/common/interface';
import { DEFAULT_SITE_NAME } from '@/common/constants';

interface SiteInfoType {
  siteInfo: AdminSettingsGeneral;
  version: string;
  revision: string;
  update: (params: AdminSettingsGeneral) => void;
  updateVersion: (ver: string, revision: string) => void;
  users: AdminSettingsUsers;
  updateUsers: (users: SiteInfoType['users']) => void;
}

const defaultUsersConf: AdminSettingsUsers = {
  allow_update_avatar: false,
  allow_update_bio: false,
  allow_update_display_name: false,
  allow_update_location: false,
  allow_update_username: false,
  allow_update_website: false,
  default_avatar: 'system',
  gravatar_base_url: 'https://www.gravatar.com/avatar/',
};

const siteInfo = create<SiteInfoType>((set) => ({
  siteInfo: {
    name: DEFAULT_SITE_NAME,
    description: '',
    short_description: '',
    site_url: '',
    contact_email: '',
    check_update: true,
    permalink: 1,
  },
  users: defaultUsersConf,
  version: '',
  revision: '',
  update: (params) =>
    set((_) => {
      const o = { ..._.siteInfo, ...params };
      if (!o.name) {
        o.name = DEFAULT_SITE_NAME;
      }
      return {
        siteInfo: o,
      };
    }),
  updateVersion: (ver, revision) => {
    set(() => {
      return { version: ver, revision };
    });
  },
  updateUsers: (users) => {
    set(() => {
      users ||= defaultUsersConf;
      return { users };
    });
  },
}));

export default siteInfo;

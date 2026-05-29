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

import React, { FC } from 'react';
import { Nav } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { NavLink, useMatch } from 'react-router-dom';

import { useGetUserPluginList } from '@/services';

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'settings.nav' });
  const settingMatch = useMatch('/users/settings/:setting');
  const { data } = useGetUserPluginList();

  return (
    <Nav variant="pills" className="flex-column">
      <NavLink
        className={({ isActive }) =>
          isActive || !settingMatch ? 'nav-link active' : 'nav-link'
        }
        to="/users/settings/profile">
        {t('profile')}
      </NavLink>
      <NavLink className="nav-link" to="/users/settings/notify">
        {t('notification')}
      </NavLink>
      <NavLink className="nav-link" to="/users/settings/account">
        {t('account')}
      </NavLink>
      <NavLink className="nav-link" to="/users/settings/interface">
        {t('interface')}
      </NavLink>
      <NavLink className="nav-link" to="/users/settings/anonymity">
        {t('anonymity')}
      </NavLink>
      {data?.map((item) => {
        return (
          <NavLink
            key={item.slug_name}
            className="nav-link w-100 text-truncate"
            to={`/users/settings/${item.slug_name}`}>
            {item.name}
          </NavLink>
        );
      })}
    </Nav>
  );
};

export default React.memo(Index);

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

import { useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import cloneDeep from 'lodash/cloneDeep';

import { AccordionNav, Icon } from '@/components';
import type { MenuItem } from '@/components/AccordionNav';
import { ADMIN_NAV_MENUS } from '@/common/constants';
import { useQueryPlugins } from '@/services';
import { interfaceStore } from '@/stores';

const AdminSideNav = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'btns' });
  const interfaceLang = interfaceStore((_) => _.interface.language);
  const { data: configurablePlugins, mutate: updateConfigurablePlugins } =
    useQueryPlugins({
      status: 'active',
      have_config: true,
    });

  const menus = cloneDeep(ADMIN_NAV_MENUS) as MenuItem[];
  if (configurablePlugins && configurablePlugins.length > 0) {
    menus.forEach((item) => {
      if (item.name === 'plugins' && item.children) {
        item.children = [
          ...item.children,
          ...configurablePlugins.map(
            (plugin): MenuItem => ({
              name: plugin.slug_name,
              displayName: plugin.name,
            }),
          ),
        ];
      }
    });
  }

  const observePlugins = (evt) => {
    if (evt.data.msgType === 'refreshConfigurablePlugins') {
      updateConfigurablePlugins();
    }
  };
  useEffect(() => {
    window.addEventListener('message', observePlugins);
    return () => {
      window.removeEventListener('message', observePlugins);
    };
  }, []);
  useEffect(() => {
    updateConfigurablePlugins();
  }, [interfaceLang]);

  return (
    <div id="adminSideNav">
      <NavLink to="/" className="pb-3 d-inline-block link-secondary">
        <Icon name="arrow-left" className="me-2" />
        <span>{t('back_sites')}</span>
      </NavLink>
      <AccordionNav menus={menus} path="/admin/" />
    </div>
  );
};

export default AdminSideNav;

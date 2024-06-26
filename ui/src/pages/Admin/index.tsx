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

import { FC, useEffect } from 'react';
import { Container, Row, Col } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Outlet, useMatch } from 'react-router-dom';

import cloneDeep from 'lodash/cloneDeep';

import { usePageTags } from '@/hooks';
import { AccordionNav } from '@/components';
import { ADMIN_NAV_MENUS } from '@/common/constants';
import { useQueryPlugins } from '@/services';
import { interfaceStore } from '@/stores';

import './index.scss';

const g10Paths = [
  'dashboard',
  'questions',
  'answers',
  'users',
  'flags',
  'installed-plugins',
];
const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_title' });
  const pathMatch = useMatch('/admin/:path');
  const curPath = pathMatch?.params.path || 'dashboard';

  const interfaceLang = interfaceStore((_) => _.interface.language);
  const { data: configurablePlugins, mutate: updateConfigurablePlugins } =
    useQueryPlugins({
      status: 'active',
      have_config: true,
    });

  const menus = cloneDeep(ADMIN_NAV_MENUS);
  if (configurablePlugins && configurablePlugins.length > 0) {
    menus.forEach((item) => {
      if (item.name === 'plugins' && item.children) {
        item.children = [
          ...item.children,
          ...configurablePlugins.map((plugin) => ({
            name: plugin.slug_name,
            displayName: plugin.name,
          })),
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

  usePageTags({
    title: t('admin'),
  });
  return (
    <>
      <div className="bg-light py-2">
        <Container className="py-1">
          <h6 className="mb-0 fw-bold lh-base">
            {t('title', { keyPrefix: 'admin.admin_header' })}
          </h6>
        </Container>
      </div>
      <Container className="admin-container">
        <Row>
          <Col lg={2}>
            <AccordionNav menus={menus} path="/admin/" />
          </Col>
          <Col lg={g10Paths.find((v) => curPath === v) ? 10 : 6}>
            <Outlet />
          </Col>
        </Row>
      </Container>
    </>
  );
};

export default Index;

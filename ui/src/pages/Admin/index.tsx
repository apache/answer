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

import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { Row, Col } from 'react-bootstrap';
import { Outlet, useMatch } from 'react-router-dom';

import { usePageTags } from '@/hooks';
import { AdminSideNav, Footer } from '@/components';

import '@/common/sideNavLayout.scss';
import './index.scss';

const g10Paths = [
  'dashboard',
  'questions',
  'answers',
  'users',
  'badges',
  'flags',
  'installed-plugins',
];
const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_title' });
  const pathMatch = useMatch('/admin/:path');
  const curPath = pathMatch?.params.path || 'dashboard';

  usePageTags({
    title: t('admin'),
  });
  return (
    <div className="admin-container d-flex">
      <div
        className="position-sticky px-3 border-end pt-4 d-none d-xl-block"
        id="pcSideNav">
        <AdminSideNav />
      </div>
      <div className="flex-fill w-100">
        <div className="d-flex justify-content-center px-0 px-md-4">
          <div className="answer-container main-mx-with">
            <Row className="py-4">
              <Col className="page-main flex-auto">
                <Outlet />
              </Col>
              {g10Paths.find((v) => curPath === v) ? null : (
                <Col className="page-right-side" />
              )}
            </Row>
          </div>
        </div>
        <div className="d-flex justify-content-center px-0 px-md-4">
          <div className="main-mx-with">
            <Footer />
          </div>
        </div>
      </div>
    </div>
  );
};

export default Index;

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
import { Nav } from 'react-bootstrap';
import {
  NavLink,
  useLocation,
  useNavigate,
  useSearchParams,
} from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { loggedUserInfoStore, sideNavStore, aiControlStore } from '@/stores';
import { Icon, PluginRender, CampusSectionNav } from '@/components';
import { PluginType } from '@/utils/pluginKit';
import request from '@/utils/request';

import './index.scss';

const Index: FC = () => {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const [searchParams] = useSearchParams();
  const { user: userInfo } = loggedUserInfoStore();
  const { can_revision, revision } = sideNavStore();
  const { ai_enabled } = aiControlStore();
  const navigate = useNavigate();

  return (
    <div className="side-nav-content d-flex h-100 flex-column" id="sideNav">
      <Nav variant="pills" className="flex-column">
        <NavLink
          to="/"
          end
          className={() =>
            pathname === '/' ||
            (pathname === '/questions' && !searchParams.has('section'))
              ? 'nav-link active'
              : 'nav-link'
          }>
          <Icon name="house-door-fill" className="me-2" />
          <span>{t('header.nav.question')}</span>
        </NavLink>
      </Nav>

      <CampusSectionNav />

      <Nav variant="pills" className="flex-column mt-3">
        {ai_enabled && (
          <NavLink
            to="/ai-assistant"
            className={() =>
              pathname === '/ai-assistant' ? 'nav-link active' : 'nav-link'
            }>
            <Icon name="chat-square-text-fill" className="me-2" />
            <span>{t('ai_assistant', { keyPrefix: 'page_title' })}</span>
          </NavLink>
        )}

        <PluginRender
          slug_name="quick_links"
          type={PluginType.Sidebar}
          request={request}
          navigate={navigate}
        />
      </Nav>

      {can_revision || userInfo?.role_id === 2 ? (
        <Nav
          variant="pills"
          className="side-nav-admin flex-column mt-auto pt-4">
          <div className="px-3 pb-2 small fw-bold text-secondary">
            {t('header.nav.moderation')}
          </div>
          {can_revision && (
            <NavLink to="/review" className="nav-link">
              <Icon name="shield-fill-check" className="me-2" />
              <span>{t('header.nav.review')}</span>
              {revision > 0 ? (
                <span className="badge rounded-pill bg-danger float-end mt-1">
                  {revision > 99 ? '99+' : revision}
                </span>
              ) : null}
            </NavLink>
          )}

          {userInfo?.role_id === 2 ? (
            <NavLink to="/admin" className="nav-link">
              <Icon name="gear-fill" className="me-2" />
              <span>{t('header.nav.admin')}</span>
            </NavLink>
          ) : null}
        </Nav>
      ) : null}
    </div>
  );
};

export default Index;

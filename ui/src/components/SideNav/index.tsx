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
import { NavLink, useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { loggedUserInfoStore, sideNavStore, aiControlStore } from '@/stores';
import { Icon, PluginRender } from '@/components';
import { PluginType } from '@/utils/pluginKit';
import request from '@/utils/request';

import './index.scss';

const Index: FC = () => {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const { user: userInfo } = loggedUserInfoStore();
  const { can_revision, revision } = sideNavStore();
  const { ai_enabled } = aiControlStore();
  const navigate = useNavigate();

  return (
    <Nav variant="pills" className="flex-column" id="sideNav">
      <NavLink
        to="/questions"
        className={({ isActive }) =>
          isActive || pathname === '/' ? 'nav-link active' : 'nav-link'
        }>
        <Icon name="question-circle-fill" className="me-2" />
        <span>{t('header.nav.question')}</span>
      </NavLink>

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

      <NavLink
        to="/tags"
        className={() =>
          pathname === '/tags' ? 'nav-link active' : 'nav-link'
        }>
        <Icon name="tags-fill" className="me-2" />
        <span>{t('header.nav.tag')}</span>
      </NavLink>

      <NavLink to="/users" className="nav-link">
        <Icon name="people-fill" className="me-2" />
        <span>{t('header.nav.user')}</span>
      </NavLink>

      <NavLink to="/badges" className="nav-link">
        <Icon name="award-fill" className="me-2" />
        <span>{t('header.nav.badges')}</span>
      </NavLink>

      <PluginRender
        slug_name="quick_links"
        type={PluginType.Sidebar}
        request={request}
        navigate={navigate}
      />

      {can_revision || userInfo?.role_id === 2 ? (
        <>
          <div className="py-2 px-3 mt-3 small fw-bold">
            {t('header.nav.moderation')}
          </div>
          {can_revision && (
            <NavLink to="/review" className="nav-link">
              <Icon name="shield-fill-check" className="me-2" />
              <span>{t('header.nav.review')}</span>
              <span className="float-end">
                {revision > 99 ? '99+' : revision > 0 ? revision : ''}
              </span>
            </NavLink>
          )}

          {userInfo?.role_id === 2 ? (
            <NavLink to="/admin" className="nav-link">
              <Icon name="gear-fill" className="me-2" />
              <span>{t('header.nav.admin')}</span>
            </NavLink>
          ) : null}
        </>
      ) : null}
    </Nav>
  );
};

export default Index;

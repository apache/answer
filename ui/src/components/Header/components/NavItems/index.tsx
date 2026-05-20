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

import { FC, memo } from 'react';
import { Nav, Dropdown } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { NavLink, useNavigate } from 'react-router-dom';

import type * as Type from '@/common/interface';
import { Avatar, Icon } from '@/components';
import { floppyNavigation, isDarkTheme } from '@/utils';
import { userCenterStore } from '@/stores';
import { REACT_BASE_PATH } from '@/router/alias';

interface Props {
  redDot: Type.NotificationStatus | undefined;
  userInfo: Type.UserInfoRes;
  logOut: (e) => void;
}

const Index: FC<Props> = ({ redDot, userInfo, logOut }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { agent: ucAgent } = userCenterStore();
  const handleLinkClick = (evt) => {
    if (floppyNavigation.shouldProcessLinkClick(evt)) {
      evt.preventDefault();
      const href = evt.currentTarget.getAttribute('href');
      floppyNavigation.navigate(href, {
        handler: navigate,
      });
    }
  };
  return (
    <>
      <Nav className="flex-row">
        <NavLink
          to="/users/notifications/inbox"
          title={t('inbox', { keyPrefix: 'notifications' })}
          className="icon-link nav-link d-flex align-items-center justify-content-center p-0 me-2 position-relative">
          <Icon name="bell-fill" className="fs-4" />
          {(redDot?.inbox || 0) > 0 && (
            <div className="unread-dot bg-danger">
              <span className="visually-hidden">
                {t('new_alerts', { keyPrefix: 'notifications' })}
              </span>
            </div>
          )}
        </NavLink>

        <NavLink
          to="/users/notifications/achievement"
          title={t('achievement', { keyPrefix: 'notifications' })}
          className="icon-link nav-link d-flex align-items-center justify-content-center p-0 me-2 position-relative">
          <Icon name="trophy-fill" className="fs-4" />
          {(redDot?.achievement || 0) > 0 && (
            <div className="unread-dot bg-danger">
              <span className="visually-hidden">
                {t('new_alerts', { keyPrefix: 'notifications' })}
              </span>
            </div>
          )}
        </NavLink>
      </Nav>

      <Dropdown align="end" data-bs-theme={isDarkTheme() ? 'dark' : 'light'}>
        <Dropdown.Toggle
          variant="success"
          id="dropdown-basic"
          as="a"
          role="button"
          className="no-toggle pointer">
          <Avatar
            size="36px"
            avatar={userInfo?.avatar}
            alt={userInfo?.display_name}
            searchStr="s=96"
          />
        </Dropdown.Toggle>

        <Dropdown.Menu className="position-absolute">
          <Dropdown.Item
            href={`${REACT_BASE_PATH}/users/${userInfo.username}`}
            onClick={handleLinkClick}>
            {t('header.nav.profile')}
          </Dropdown.Item>
          <Dropdown.Item
            href={`${REACT_BASE_PATH}/users/${userInfo.username}/bookmarks`}
            onClick={handleLinkClick}>
            {t('header.nav.bookmark')}
          </Dropdown.Item>
          <Dropdown.Item
            href={`${REACT_BASE_PATH}/users/settings/profile`}
            onClick={handleLinkClick}>
            {t('header.nav.setting')}
          </Dropdown.Item>
          <Dropdown.Divider />
          <Dropdown.Item
            href={`${REACT_BASE_PATH}/users/logout`}
            onClick={(e) => logOut(e)}>
            {t('header.nav.logout')}
          </Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown>
      {/* Dropdown for user center agent info */}
      {ucAgent?.enabled &&
      (ucAgent?.agent_info?.url ||
        ucAgent?.agent_info?.control_center?.length) ? (
        <Dropdown align="end" data-bs-theme={isDarkTheme() ? 'dark' : 'light'}>
          <Dropdown.Toggle
            variant="success"
            id="dropdown-uca"
            as="span"
            className="no-toggle">
            <Nav>
              <Icon
                name="grid-3x3-gap-fill"
                className="nav-link pointer p-0 fs-4 ms-3"
              />
            </Nav>
          </Dropdown.Toggle>

          <Dropdown.Menu className="position-absolute">
            {ucAgent.agent_info.url ? (
              <Dropdown.Item href={ucAgent.agent_info.url}>
                {ucAgent.agent_info.name}
              </Dropdown.Item>
            ) : null}
            {ucAgent.agent_info.url &&
            ucAgent.agent_info.control_center?.length ? (
              <Dropdown.Divider />
            ) : null}
            {ucAgent.agent_info.control_center?.map((ctrl) => {
              return (
                <Dropdown.Item key={ctrl.name} href={ctrl.url}>
                  {ctrl.label}
                </Dropdown.Item>
              );
            })}
          </Dropdown.Menu>
        </Dropdown>
      ) : null}
    </>
  );
};

export default memo(Index);

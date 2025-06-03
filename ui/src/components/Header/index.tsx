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

import { FC, memo, useState, useEffect } from 'react';
import { Navbar, Nav, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link, NavLink, useLocation, useMatch } from 'react-router-dom';

import classnames from 'classnames';

import { userCenter, floppyNavigation, isLight } from '@/utils';
import {
  loggedUserInfoStore,
  siteInfoStore,
  brandingStore,
  loginSettingStore,
  themeSettingStore,
  sideNavStore,
} from '@/stores';
import { logout, useQueryNotificationStatus } from '@/services';
import { Icon, MobileSideNav } from '@/components';

import NavItems from './components/NavItems';
import SearchInput from './components/SearchInput';

import './index.scss';

const Header: FC = () => {
  const location = useLocation();
  const { user, clear: clearUserStore } = loggedUserInfoStore();
  const { t } = useTranslation();
  const siteInfo = siteInfoStore((state) => state.siteInfo);
  const brandingInfo = brandingStore((state) => state.branding);
  const loginSetting = loginSettingStore((state) => state.login);
  const { updateReview } = sideNavStore();
  const { data: redDot } = useQueryNotificationStatus();
  const [showMobileSideNav, setShowMobileSideNav] = useState(false);

  const [showMobileSearchInput, setShowMobileSearchInput] = useState(false);
  /**
   * Automatically append `tag` information when creating a question
   */
  const tagMatch = useMatch('/tags/:slugName');
  let askUrl = '/questions/ask';
  if (tagMatch && tagMatch.params.slugName) {
    askUrl = `${askUrl}?tags=${encodeURIComponent(tagMatch.params.slugName)}`;
  }

  useEffect(() => {
    updateReview({
      can_revision: Boolean(redDot?.can_revision),
      revision: Number(redDot?.revision),
    });
  }, [redDot]);

  const handleLogout = async (evt) => {
    evt.preventDefault();
    await logout();
    clearUserStore();
    window.location.replace(window.location.href);
  };

  useEffect(() => {
    setShowMobileSearchInput(false);
    setShowMobileSideNav(false);
  }, [location.pathname]);

  let navbarStyle = 'theme-light';
  let themeMode = 'light';
  const { theme, theme_config } = themeSettingStore((_) => _);
  if (theme_config?.[theme]?.navbar_style) {
    // const color = theme_config[theme].navbar_style.startsWith('#')
    themeMode = isLight(theme_config[theme].navbar_style) ? 'light' : 'dark';
    navbarStyle = `theme-${themeMode}`;
  }

  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1199.9) {
        setShowMobileSideNav(false);
        setShowMobileSearchInput(false);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  return (
    <Navbar
      data-bs-theme={themeMode}
      expand="xl"
      className={classnames('sticky-top', navbarStyle)}
      style={{
        backgroundColor: theme_config[theme].navbar_style,
      }}
      id="header">
      <div className="w-100 d-flex align-items-center px-3">
        <Navbar.Toggle
          className="answer-navBar me-2"
          onClick={() => {
            setShowMobileSideNav(!showMobileSideNav);
            setShowMobileSearchInput(false);
          }}
        />

        <Navbar.Brand
          to="/"
          as={Link}
          className="lh-1 me-0 me-sm-5 p-0 nav-text">
          {brandingInfo.logo ? (
            <>
              <img
                className="d-none d-xl-block logo me-0"
                src={brandingInfo.logo}
                alt={siteInfo.name}
              />

              <img
                className="xl-none logo me-0"
                src={brandingInfo.mobile_logo || brandingInfo.logo}
                alt={siteInfo.name}
              />
            </>
          ) : (
            <span>{siteInfo.name}</span>
          )}
        </Navbar.Brand>

        <SearchInput className="d-none d-lg-block maxw-560" />

        <Nav className="d-block d-lg-none me-2 ms-auto">
          <Button
            variant="link"
            onClick={() => {
              setShowMobileSideNav(false);
              setShowMobileSearchInput(!showMobileSearchInput);
            }}
            className="p-0 btn-no-border icon-link nav-link d-flex align-items-center justify-content-center">
            <Icon name="search" className="lh-1 fs-4" />
          </Button>
        </Nav>

        {/* pc nav */}
        {user?.username ? (
          <Nav className="d-flex align-items-center flex-nowrap flex-row">
            <Nav.Item className="me-2 d-block d-xl-none">
              <NavLink
                to={askUrl}
                className="d-block icon-link nav-link text-center">
                <Icon name="plus-lg" className="lh-1 fs-4" />
              </NavLink>
            </Nav.Item>

            <Nav.Item className="me-2 d-none d-xl-block">
              <NavLink
                to={askUrl}
                className="nav-link d-flex align-items-center text-capitalize text-nowrap">
                <Icon name="plus-lg" className="me-2 lh-1 fs-4" />
                <span>{t('btns.create')}</span>
              </NavLink>
            </Nav.Item>

            <NavItems redDot={redDot} userInfo={user} logOut={handleLogout} />
          </Nav>
        ) : (
          <>
            <Link
              className={classnames('me-2 btn btn-link', {
                'link-light': navbarStyle === 'theme-dark',
                'link-primary': navbarStyle !== 'theme-dark',
              })}
              onClick={() => floppyNavigation.storageLoginRedirect()}
              to={userCenter.getLoginUrl()}>
              {t('btns.login')}
            </Link>
            {loginSetting.allow_new_registrations && (
              <Link
                className={classnames(
                  'btn',
                  navbarStyle === 'theme-dark' ? 'btn-light' : 'btn-primary',
                )}
                to={userCenter.getSignUpUrl()}>
                {t('btns.signup')}
              </Link>
            )}
          </>
        )}
      </div>

      {showMobileSearchInput && (
        <div className="w-100 px-3 mt-2 d-block d-lg-none">
          <SearchInput />
        </div>
      )}

      <MobileSideNav show={showMobileSideNav} onHide={setShowMobileSideNav} />
    </Navbar>
  );
};

export default memo(Header);

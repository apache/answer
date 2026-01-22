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

import { getLoggedUserInfo, getAppSettings } from '@/services';
import {
  loggedUserInfoStore,
  siteInfoStore,
  interfaceStore,
  brandingStore,
  loginSettingStore,
  customizeStore,
  themeSettingStore,
  seoSettingStore,
  loginToContinueStore,
  pageTagStore,
  writeSettingStore,
  siteSecurityStore,
} from '@/stores';
import { RouteAlias } from '@/router/alias';
import {
  LOGGED_TOKEN_STORAGE_KEY,
  REDIRECT_PATH_STORAGE_KEY,
} from '@/common/constants';
import Storage from '@/utils/storage';

import { setupAppLanguage, setupAppTimeZone, setupAppTheme } from './localize';
import { floppyNavigation, NavigateConfig } from './floppyNavigation';
import { pullUcAgent, getSignUpUrl } from './userCenter';

type TLoginState = {
  isLogged: boolean;
  isNotActivated: boolean;
  isActivated: boolean;
  isNormal: boolean;
  isAdmin: boolean;
  isModerator: boolean;
};

export type TGuardResult = {
  ok: boolean;
  redirect?: string;
  error?: {
    code?: number | string;
    msg?: string;
  };
};
export type TGuardFunc = (args: {
  loaderData?: any;
  path?: string;
  page?: string;
}) => TGuardResult;

export const deriveLoginState = (): TLoginState => {
  const ls: TLoginState = {
    isLogged: false,
    isNotActivated: false,
    isActivated: false,
    isNormal: false,
    isAdmin: false,
    isModerator: false,
  };
  const { user } = loggedUserInfoStore.getState();
  if (user.access_token) {
    ls.isLogged = true;
  }
  if (ls.isLogged && user.mail_status === 1) {
    ls.isActivated = true;
  }
  if (ls.isLogged && user.mail_status === 2) {
    ls.isNotActivated = true;
  }

  if (ls.isActivated) {
    ls.isNormal = true;
  }
  if (ls.isNormal && user.role_id === 2) {
    ls.isAdmin = true;
  }
  if (ls.isNormal && user.role_id === 3) {
    ls.isModerator = true;
  }
  return ls;
};

export const IGNORE_PATH_LIST = [
  RouteAlias.login,
  RouteAlias.signUp,
  RouteAlias.accountRecovery,
  RouteAlias.changeEmail,
  RouteAlias.passwordReset,
  RouteAlias.accountActivation,
  RouteAlias.confirmNewEmail,
  RouteAlias.confirmEmail,
  RouteAlias.authLanding,
  '/user-center/',
];

export const isIgnoredPath = (ignoredPath?: string | string[]) => {
  if (!ignoredPath) {
    ignoredPath = IGNORE_PATH_LIST;
  }
  if (!Array.isArray(ignoredPath)) {
    ignoredPath = [ignoredPath];
  }
  const matchingPath = ignoredPath.find((p) => {
    return floppyNavigation.matchToCurrentHref(p);
  });
  return !!matchingPath;
};

let pluTimestamp = 0;
export const pullLoggedUser = async (isInitPull = false) => {
  /**
   * WARN:
   * - dedupe pull requests in this time span in 10 seconds
   * - isInitPull:
   *   Requests sent by the initialisation method cannot be throttled
   *   and may cause Promise.allSettled to complete early in React development mode,
   *   resulting in inaccurate application data.
   */
  //
  if (!isInitPull && Date.now() - pluTimestamp < 1000 * 10) {
    return;
  }
  pluTimestamp = Date.now();
  const loggedUserInfo = await getLoggedUserInfo({
    passingError: true,
  }).catch(() => {
    pluTimestamp = 0;
    loggedUserInfoStore.getState().clear(false);
  });
  if (loggedUserInfo) {
    loggedUserInfoStore.getState().update(loggedUserInfo);
  }
};

export const logged = () => {
  const gr: TGuardResult = { ok: true };
  const us = deriveLoginState();
  if (!us.isLogged) {
    gr.ok = false;
    gr.redirect = RouteAlias.login;
  }
  return gr;
};

export const loggedRedirectHome = () => {
  const gr: TGuardResult = { ok: true };
  const us = deriveLoginState();
  if (!us.isLogged) {
    gr.ok = false;
    gr.redirect = RouteAlias.home;
  }
  return gr;
};

export const notLogged = () => {
  const gr: TGuardResult = { ok: true };
  const us = deriveLoginState();
  if (us.isLogged) {
    gr.ok = false;
    gr.redirect = RouteAlias.home;
  }
  return gr;
};

export const notActivated = () => {
  const gr: TGuardResult = { ok: true };
  const us = deriveLoginState();
  if (us.isActivated) {
    gr.ok = false;
    gr.redirect = RouteAlias.home;
  }
  return gr;
};

export const activated = () => {
  const gr = logged();
  const us = deriveLoginState();
  if (us.isNotActivated) {
    gr.ok = false;
    gr.redirect = RouteAlias.inactive;
  }
  return gr;
};

export const admin = () => {
  const gr = logged();
  const us = deriveLoginState();
  if (gr.ok && !us.isAdmin) {
    gr.ok = false;
    gr.error = {
      code: '403',
      msg: '',
    };
    gr.redirect = '';
  }
  return gr;
};

export const isAdminOrModerator = () => {
  const gr = logged();
  const us = deriveLoginState();
  if (gr.ok && !us.isAdmin && !us.isModerator) {
    gr.ok = false;
    gr.error = {
      code: '403',
      msg: '',
    };
    gr.redirect = '';
  }
  return gr;
};

export const isEditable = (args) => {
  const loaderData = args?.loaderData || {};
  const gr: TGuardResult = { ok: true };
  if (loaderData.code === 400) {
    gr.ok = false;
    gr.error = {
      code: '403',
      msg: loaderData.msg,
    };
  }
  return gr;
};

export const allowNewRegistration = () => {
  const gr: TGuardResult = { ok: true };
  const loginSetting = loginSettingStore.getState().login;
  if (!loginSetting.allow_new_registrations) {
    gr.ok = false;
    gr.redirect = RouteAlias.home;
  }
  return gr;
};

export const singUpAgent = () => {
  const gr: TGuardResult = { ok: true };
  const signUpUrl = getSignUpUrl();
  if (signUpUrl !== RouteAlias.signUp) {
    gr.ok = false;
    gr.redirect = signUpUrl;
  }
  return gr;
};

export const shouldLoginRequired = () => {
  const gr: TGuardResult = { ok: true };
  const { login_required } = siteSecurityStore.getState();
  if (!login_required) {
    return gr;
  }
  const us = deriveLoginState();
  if (us.isLogged) {
    return gr;
  }
  if (isIgnoredPath(IGNORE_PATH_LIST)) {
    return gr;
  }
  gr.ok = false;
  gr.redirect = RouteAlias.login;
  return gr;
};

/**
 * try user was logged and all state ok
 * @param canNavigate // if true, will navigate to login page if not logged
 */
export const tryNormalLogged = (canNavigate: boolean = false) => {
  const us = deriveLoginState();

  if (us.isNormal) {
    return true;
  }
  // must assert logged state first and return
  if (!us.isLogged) {
    if (canNavigate) {
      loginToContinueStore.getState().update({ show: true });
    }
    return false;
  }
  if (us.isNotActivated) {
    floppyNavigation.navigate(RouteAlias.inactive);
  }

  return false;
};

export const tryLoggedAndActivated = () => {
  const gr: TGuardResult = { ok: true };
  const us = deriveLoginState();

  if (!us.isLogged || !us.isActivated) {
    gr.ok = false;
  }
  return gr;
};

/**
 * Auto handling of page redirect logic after a successful login
 */
export const handleLoginRedirect = (handler?: NavigateConfig['handler']) => {
  const redirectUrl = Storage.get(REDIRECT_PATH_STORAGE_KEY) || RouteAlias.home;
  Storage.remove(REDIRECT_PATH_STORAGE_KEY);
  floppyNavigation.navigate(redirectUrl, {
    handler,
    options: { replace: true },
  });
};

/**
 * Unified processing of login logic after getting `access_token`
 */
export const handleLoginWithToken = (
  token: string | null,
  handler?: NavigateConfig['handler'],
) => {
  if (token) {
    Storage.set(LOGGED_TOKEN_STORAGE_KEY, token);
    setTimeout(() => {
      getLoggedUserInfo().then((res) => {
        loggedUserInfoStore.getState().update(res);
        const userStat = deriveLoginState();
        if (userStat.isNotActivated) {
          floppyNavigation.navigate(RouteAlias.inactive, {
            handler,
            options: {
              replace: true,
            },
          });
        } else {
          handleLoginRedirect(handler);
        }
      });
    });
  } else {
    floppyNavigation.navigate(RouteAlias.home, {
      handler,
      options: {
        replace: true,
      },
    });
  }
};

/**
 * Initialize app configuration
 */
export const initAppSettingsStore = async () => {
  const appSettings = await getAppSettings();
  if (appSettings) {
    siteInfoStore.getState().update(appSettings.general);
    siteInfoStore
      .getState()
      .updateVersion(appSettings.version, appSettings.revision);
    siteInfoStore.getState().updateUsers(appSettings.site_users);
    interfaceStore.getState().update(appSettings.interface);
    pageTagStore.getState().update({
      title: appSettings.general?.name,
      description: appSettings.general?.description,
    });
    brandingStore.getState().update(appSettings.branding);
    loginSettingStore.getState().update(appSettings.login);
    customizeStore.getState().update(appSettings.custom_css_html);
    themeSettingStore.getState().update(appSettings.theme);
    seoSettingStore.getState().update(appSettings.site_seo);
    writeSettingStore.getState().update({
      ...appSettings.site_advanced,
      ...appSettings.site_questions,
      ...appSettings.site_tags,
    });
    siteSecurityStore.getState().update(appSettings.site_security);
  }
};

export const askRedirect = () => {
  const queryString = window.location.search;

  return {
    ok: false,
    redirect: `/questions/add${queryString}`,
  };
};

export const linkedRedirect = () => {
  const queryString = window.location.search;
  const pathname = window.location.pathname.replace('/questions', '');

  return {
    ok: false,
    redirect: `${pathname}${queryString}`,
  };
};

let appInitialized = false;
export const setupApp = async () => {
  /**
   * This cannot be removed:
   * clicking on the current navigation link will trigger a call to the routing loader,
   * even though the page is not refreshed.
   */
  if (appInitialized) {
    return;
  }
  /**
   * WARN:
   * 1. must pre init logged user info for router guard
   * 2. must pre init app settings for app render
   */
  await Promise.allSettled([initAppSettingsStore(), pullLoggedUser(true)]);
  await Promise.allSettled([pullUcAgent()]);
  setupAppLanguage();
  setupAppTimeZone();
  setupAppTheme();
  /**
   * WARN:
   * Initialization must be completed after all initialization actions,
   * otherwise the problem of rendering twice in React development mode can lead to inaccurate data or flickering pages
   */
  appInitialized = true;
};

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

import dayjs from 'dayjs';
import i18next from 'i18next';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

import {
  interfaceStore,
  loggedUserInfoStore,
  themeSettingStore,
} from '@/stores';
import {
  CURRENT_LANG_STORAGE_KEY,
  DEFAULT_LANG,
  LANG_RESOURCE_STORAGE_KEY,
  DEFAULT_THEME,
} from '@/common/constants';
import {
  getAdminLanguageOptions,
  getInstallLanguageConfig,
  getLanguageConfig,
  getLanguageOptions,
} from '@/services';
import { changeTheme } from '@/utils/common';

import Storage from './storage';

/**
 * localize kit for i18n
 */
export const loadLanguageOptions = async (forAdmin = false) => {
  const languageOptions = forAdmin
    ? await getAdminLanguageOptions()
    : await getLanguageOptions();
  if (process.env.NODE_ENV === 'development') {
    const { default: optConf } = await import('@i18n/i18n.yaml');
    optConf?.language_options.forEach((opt) => {
      if (!languageOptions.find((_) => opt.value === _.value)) {
        languageOptions.push(opt);
      }
    });
  }
  return languageOptions;
};

const pullLanguageConf = (res) => {
  if (window.location.pathname === '/install') {
    return getInstallLanguageConfig(res.lng).then((langConf) => {
      if (langConf && langConf.ui) {
        res.resources = langConf.ui;
        Storage.set(LANG_RESOURCE_STORAGE_KEY, res);
      }
    });
  }

  return getLanguageConfig().then((langConf) => {
    if (langConf && langConf.ui) {
      res.resources = langConf.ui;
      Storage.set(LANG_RESOURCE_STORAGE_KEY, res);
      // The stale cached bundle was already added at boot; merge the fresh
      // one in place so new/changed keys apply without a second page load.
      i18next.addResourceBundle(res.lng, 'translation', langConf.ui, true, true);
    }
  });
};

const addI18nResource = async (langName) => {
  const res = { lng: langName, resources: undefined };
  const storageResource = Storage.get(LANG_RESOURCE_STORAGE_KEY);
  if (process.env.NODE_ENV === 'development') {
    try {
      const { default: resConf } = await import(`@i18n/${langName}.yaml`);
      res.resources = resConf.ui;
    } catch (ex) {
      console.error('addI18nResource error: ', ex);
    }
  } else if (storageResource && storageResource.lng === res.lng) {
    res.resources = storageResource.resources;
    pullLanguageConf(res);
  } else {
    await pullLanguageConf(res);
  }

  if (res.resources) {
    i18next.addResourceBundle(
      res.lng,
      'translation',
      res.resources,
      true,
      true,
    );
  }
};

export const getCurrentLang = () => {
  const loggedUser = loggedUserInfoStore.getState().user;
  const adminInterface = interfaceStore.getState().interface;
  const fallbackLang = Storage.get(CURRENT_LANG_STORAGE_KEY) || DEFAULT_LANG;
  let currentLang = loggedUser.language;
  // `default` mean use language value from admin interface
  if (/default/i.test(currentLang)) {
    currentLang = adminInterface.language;
  }
  currentLang ||= fallbackLang;
  return currentLang;
};

export const getCurrentTheme = () => {
  const loggedUser = loggedUserInfoStore.getState().user;
  const adminTheme = themeSettingStore.getState().color_scheme;
  const fallbackTheme = DEFAULT_THEME;
  let currentTheme = loggedUser.color_scheme;
  if (/default/i.test(currentTheme)) {
    currentTheme = adminTheme;
  }
  currentTheme ||= fallbackTheme;
  return currentTheme;
};

/**
 * localize for Day.js
 */
dayjs.extend(utc);
dayjs.extend(timezone);
const localeDayjs = (langName) => {
  langName = langName.replace('_', '-').toLowerCase();
  dayjs.locale(langName);
};

export const setupAppLanguage = async () => {
  const lang = getCurrentLang();
  await addI18nResource(lang);
  localeDayjs(lang);
  i18next.changeLanguage(lang);
};

export const setupAppTimeZone = () => {
  const adminInterface = interfaceStore.getState().interface;
  if (adminInterface.time_zone) {
    dayjs.tz.setDefault(adminInterface.time_zone);
  }
};

export const setupAppTheme = () => {
  const theme = getCurrentTheme();
  changeTheme(theme);
};

export const setupInstallLanguage = async (lang) => {
  await addI18nResource(lang);
  localeDayjs(lang);
  i18next.changeLanguage(lang);
};

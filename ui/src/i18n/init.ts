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

import { initReactI18next } from 'react-i18next';

import i18next from 'i18next';
import en_US from '@i18n/en_US.yaml';
import zh_CN from '@i18n/zh_CN.yaml';

import { DEFAULT_LANG, LANG_RESOURCE_STORAGE_KEY } from '@/common/constants';
import Storage from '@/utils/storage';

/**
 * Prevent i18n from re-initialising when the page is refreshed and switching to `fallbackLng`.
 */
const initLng = i18next.language || DEFAULT_LANG;
const initResources = {
  zh_CN: {
    translation: zh_CN.ui,
  },
  en_US: {
    translation: en_US.ui,
  },
};

const storageLang = Storage.get(LANG_RESOURCE_STORAGE_KEY);
if (
  storageLang &&
  storageLang.resources &&
  storageLang.lng &&
  storageLang.lng !== 'zh_CN' &&
  storageLang.lng !== 'en_US'
) {
  initResources[storageLang.lng] = {
    translation: storageLang.resources,
  };
}

i18next
  //  pass the i18n instance to react-i18next.
  .use(initReactI18next)
  .init({
    resources: initResources,
    lng: initLng,
    fallbackLng: DEFAULT_LANG,
    interpolation: {
      escapeValue: false,
    },
    react: {
      transSupportBasicHtmlNodes: true,
      // allow <br/> and simple html elements in translations
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i'],
    },
  });

export default i18next;

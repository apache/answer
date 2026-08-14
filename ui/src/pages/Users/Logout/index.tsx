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
import { Spinner } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { usePageTags } from '@/hooks';
import { logout } from '@/services';
import { loggedUserInfoStore } from '@/stores';
import Storage from '@/utils/storage';
import { RouteAlias, BASE_ORIGIN } from '@/router/alias';
import { REDIRECT_PATH_STORAGE_KEY } from '@/common/constants';

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_title' });
  const { user: loggedUserInfo, clear: clearUserStore } = loggedUserInfoStore();

  usePageTags({
    title: t('logout'),
  });

  useEffect(() => {
    if (loggedUserInfo.username) {
      logout().then(() => {
        clearUserStore();
        Storage.remove(REDIRECT_PATH_STORAGE_KEY);
        window.location.replace(`${BASE_ORIGIN}${RouteAlias.home}`);
      });
    }
    // auto height of container
    const pageWrap = document.querySelector('.page-wrap') as HTMLElement;
    if (pageWrap) {
      pageWrap.style.display = 'contents';
    }

    return () => {
      if (pageWrap) {
        pageWrap.style.display = 'block';
      }
    };
  }, []);
  return (
    <div className="d-flex flex-column flex-shrink-1 flex-grow-1 justify-content-center align-items-center">
      <Spinner variant="secondary" />
    </div>
  );
};

export default Index;

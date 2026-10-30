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

import { FC, memo, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { usePageTags } from '@/hooks';
import { guard } from '@/utils';
import { exchangeLoginHandoff } from '@/services/common';

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_title' });
  const navigate = useNavigate();
  const handled = useRef(false);
  useEffect(() => {
    if (handled.current) {
      return;
    }
    handled.current = true;
    const code = new URLSearchParams(window.location.hash.slice(1)).get('code');
    window.history.replaceState(
      null,
      document.title,
      `${window.location.pathname}${window.location.search}`,
    );
    if (!code) {
      guard.handleLoginWithToken(null, navigate);
      return;
    }
    exchangeLoginHandoff(code)
      .then((res) => guard.handleLoginWithToken(res.access_token, navigate))
      .catch(() => guard.handleLoginWithToken(null, navigate));
  }, []);
  usePageTags({
    title: t('oauth_callback'),
  });
  return null;
};

export default memo(Index);

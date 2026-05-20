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

import { FC, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import { usePageTags } from '@/hooks';
import { useLegalPrivacy } from '@/services';
import { htmlRender } from '@/components';

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'nav_menus' });
  usePageTags({
    title: t('privacy'),
  });
  const { data: privacy } = useLegalPrivacy();
  const contentText = privacy?.privacy_policy_original_text;
  let matchUrl: URL | undefined;

  useEffect(() => {
    const fmt = document.querySelector('.fmt') as HTMLElement;
    if (!fmt) {
      return;
    }
    htmlRender(fmt, {
      copySuccessText: t('copied', { keyPrefix: 'messages' }),
      copyText: t('copy', { keyPrefix: 'messages' }),
    });
  }, [privacy?.privacy_policy_parsed_text]);

  try {
    if (contentText) {
      matchUrl = new URL(contentText);
    }
    // eslint-disable-next-line no-empty
  } catch (ex) {}
  if (matchUrl) {
    window.location.replace(matchUrl.toString());
    return null;
  }
  return (
    <>
      <h3 className="mb-4">{t('privacy')}</h3>
      <div
        className="fmt"
        dangerouslySetInnerHTML={{
          __html: privacy?.privacy_policy_parsed_text || '',
        }}
      />
    </>
  );
};

export default Index;

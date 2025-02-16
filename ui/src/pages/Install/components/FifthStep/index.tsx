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
import { Button } from 'react-bootstrap';
import { useTranslation, Trans } from 'react-i18next';

import Progress from '../Progress';

interface Props {
  visible: boolean;
  siteUrl: string;
}
const Index: FC<Props> = ({ visible, siteUrl = '' }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'install' });

  if (!visible) return null;
  return (
    <div>
      <h5>{t('ready_title')}</h5>
      <p>
        <Trans i18nKey="install.ready_desc">
          If you ever feel like changing more settings, visit
          <a href={`${siteUrl}/users/login`}> admin section</a>; find it in the
          site menu.
        </Trans>
      </p>
      <p>{t('good_luck')}</p>

      <div className="d-flex align-items-center justify-content-between">
        <Progress step={5} />
        <Button href={siteUrl}>{t('done')}</Button>
      </div>
    </div>
  );
};

export default Index;

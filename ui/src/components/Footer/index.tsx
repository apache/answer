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

import React from 'react';
import { Link } from 'react-router-dom';
import { Trans, useTranslation } from 'react-i18next';

import Row from 'react-bootstrap/Row';
import dayjs from 'dayjs';

import { siteInfoStore } from '@/stores';

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'footer' }); // Scoped translations for footer
  const fullYear = dayjs().format('YYYY');
  const siteName = siteInfoStore((state) => state.siteInfo.name);
  const cc = `${siteName} © ${fullYear}`;

  return (
    <Row>
      <footer className="py-3 d-flex flex-wrap align-items-center justify-content-between text-secondary small">
        <div className="d-flex align-items-center">
          <div className="me-3">{cc}</div>

          <Link to="/tos" className="me-3 link-secondary">
            {t('terms', { keyPrefix: 'nav_menus' })}
          </Link>

          {/* Link to Privacy Policy with right margin for spacing */}
          <Link to="/privacy" className="link-secondary">
            {t('privacy', { keyPrefix: 'nav_menus' })}
          </Link>
        </div>
        <div>
          <Trans i18nKey="footer.build_on" values={{ cc }}>
            Powered by
            <a
              href="https://answer.apache.org"
              target="_blank"
              className="link-secondary"
              rel="noreferrer">
              Apache Answer
            </a>
          </Trans>
        </div>
      </footer>
    </Row>
  );
};

export default React.memo(Index);

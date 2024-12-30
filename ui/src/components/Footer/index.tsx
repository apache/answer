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
import { Container } from 'react-bootstrap';
import { Trans, useTranslation } from 'react-i18next';

import dayjs from 'dayjs';

import { siteInfoStore } from '@/stores';

const Footer = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'footer' }); // Scoped translations for footer
  const fullYear = dayjs().format('YYYY');
  const siteName = siteInfoStore((state) => state.siteInfo.name);
  const cc = `${fullYear} ${siteName}`;

  return (
    <footer className="bg-light">
      <Container className="py-3">
        <p className="text-center mb-0 p-3 small">
          {/* Link to Privacy Policy with right margin for spacing */}
          <Link to="/privacy" className="me-3">
            {t('privacy')}
          </Link>
          {/* Link to Terms of Service with right margin */}
          <Link to="/tos">{t('terms_of_service')}</Link>
        </p>
        <p className="text-center mb-0 small">
          <Trans i18nKey="build_on" values={{ cc }}>
            Powered by{' '}
            <a
              href="https://answer.apache.org"
              target="_blank"
              rel="noopener noreferrer">
              Apache Answer
            </a>
            - the open-source software that powers Q&A communities.
            <br />
            Made with love © 2022 Answer.
          </Trans>
        </p>
      </Container>
    </footer>
  );
};

export default React.memo(Footer);

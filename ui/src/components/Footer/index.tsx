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
import { Container } from 'react-bootstrap';
import { Trans } from 'react-i18next';

import dayjs from 'dayjs';

import { siteInfoStore } from '@/stores';

const Index = () => {
  const fullYear = dayjs().format('YYYY');
  const siteName = siteInfoStore((state) => state.siteInfo.name);
  const cc = `${fullYear} ${siteName}`;
  return (
    <footer className="bg-light">
      <Container className="py-3">
        <p className="text-center mb-0 small text-secondary">
          <Trans i18nKey="footer.build_on" values={{ cc }}>
            Powered by
            {/* eslint-disable-next-line react/jsx-no-target-blank */}
            <a href="https://zcareze.cn/" target="_blank">
              zcareu
            </a>
            - apache/answer the open-source software that powers Q&A communities.
            <br />
            Made with love. © 2022 Answer.
          </Trans>
        </p>
      </Container>
    </footer>
  );
};

export default React.memo(Index);

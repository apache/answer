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

import { Card, Row, Col } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

const AnswerLinks = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.dashboard' });

  return (
    <Card className="mb-4">
      <Card.Body>
        <h6 className="mb-3">{t('links')}</h6>
        <Row>
          <Col xs={6}>
            <a
              href="https://answer.apache.org/docs"
              target="_blank"
              rel="noreferrer">
              {t('documents')}
            </a>
          </Col>
          <Col xs={6}>
            <a
              href="https://answer.apache.org/plugins"
              target="_blank"
              rel="noreferrer">
              {t('plugins')}
            </a>
          </Col>
          <Col xs={6}>
            <a
              href="https://answer.apache.org/community/support"
              target="_blank"
              rel="noreferrer">
              {t('support')}
            </a>
          </Col>
          <Col xs={6}>
            <a href="https://meta.answer.dev" target="_blank" rel="noreferrer">
              {t('forum')}
            </a>
          </Col>
          <Col xs={6}>
            <a
              href="https://answer.apache.org/blog"
              target="_blank"
              rel="noreferrer">
              {t('blog')}
            </a>
          </Col>
          <Col xs={6}>
            <a
              href="https://github.com/apache/answer"
              target="_blank"
              rel="noreferrer">
              {t('github')}
            </a>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );
};

export default AnswerLinks;

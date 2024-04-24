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
import { Card, Row, Col } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import type * as Type from '@/common/interface';

interface IProps {
  data: Type.AdminDashboard['info'];
}
const Statistics: FC<IProps> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.dashboard' });

  return (
    <Card className="mb-4">
      <Card.Body>
        <h6 className="mb-3">{t('site_statistics')}</h6>
        <Row>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('questions')}</span>
            <strong>{data.question_count}</strong>
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('answers')}</span>
            <strong>{data.answer_count}</strong>
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('comments')}</span>
            <strong>{data.comment_count}</strong>
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('votes')}</span>
            <strong>{data.vote_count}</strong>
          </Col>
          <Col xs={6}>
            <span className="text-secondary me-1">{t('users')}</span>
            <strong>{data.user_count}</strong>
          </Col>
          <Col xs={6}>
            <span className="text-secondary me-1">{t('reviews')}</span>
            <strong>
              <Link to="/review" className="ms-2">
                {data.report_count}
              </Link>
            </strong>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );
};

export default Statistics;

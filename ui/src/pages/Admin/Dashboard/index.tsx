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
import { Row, Col } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { useDashBoard } from '@/services';

import {
  AnswerLinks,
  HealthStatus,
  Statistics,
  SystemInfo,
  DailyActivity,
} from './components';

const Dashboard: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.dashboard' });
  const { data } = useDashBoard();

  if (!data) {
    return null;
  }

  return (
    <>
      <h3 className="text-capitalize">{t('title')}</h3>
      <p className="mt-4">{t('welcome')}</p>
      <Row>
        <Col lg={6}>
          <Statistics data={data.info} />
        </Col>
        <Col lg={6}>
          <HealthStatus data={data.info} />
        </Col>
        <Col lg={6}>
          <SystemInfo data={data.info} />
        </Col>
        <Col lg={6}>
          <AnswerLinks />
        </Col>
        <Col lg={12}>
          <DailyActivity />
        </Col>
      </Row>
    </>
  );
};
export default Dashboard;

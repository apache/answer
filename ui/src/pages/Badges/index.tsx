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

import { useTranslation } from 'react-i18next';
import { Row, Col } from 'react-bootstrap';

import { CardBadge } from '@/components';
import { usePageTags } from '@/hooks';
import { useGetAllBadges } from '@/services';

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'badges' });

  const { data: badgesList } = useGetAllBadges();

  usePageTags({
    title: t('title'),
  });

  return (
    <div className="pt-4 mb-5">
      <h1 className="fs-3 mb-4">{t('title')}</h1>
      {badgesList?.map((item) => {
        return (
          <div key={item.group_name} className="mb-4">
            <h2 className="fs-5 mb-4">{item.group_name}</h2>
            <Row>
              {item.badges?.map((badge) => {
                return (
                  <Col sm={6} md={4} lg={3} key={badge.id} className="mb-4">
                    <CardBadge data={badge} showAwardedCount />
                  </Col>
                );
              })}
            </Row>
          </div>
        );
      })}
    </div>
  );
};

export default Index;

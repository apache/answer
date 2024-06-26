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
import { siteInfoStore } from '@/stores';

const { gt, gte } = require('semver');

interface IProps {
  data: Type.AdminDashboard['info'];
}

const HealthStatus: FC<IProps> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.dashboard' });
  const { version, remote_version } = data.version_info || {};
  const { siteInfo } = siteInfoStore();
  let isLatest = false;
  let hasNewerVersion = false;
  if (version && remote_version) {
    isLatest = gte(version, remote_version);
    hasNewerVersion = gt(remote_version, version);
  }
  return (
    <Card className="mb-4">
      <Card.Body>
        <h6 className="mb-3">{t('site_health')}</h6>
        <Row>
          <Col xs={6} className="mb-1 d-flex align-items-center">
            <span className="text-secondary me-1">{t('version')}</span>
            <strong>{version}</strong>
            {isLatest && (
              <a
                className="ms-1 badge rounded-pill text-bg-success"
                target="_blank"
                href="https://github.com/apache/incubator-answer/releases"
                rel="noreferrer">
                {t('latest')}
              </a>
            )}
            {!isLatest && hasNewerVersion && (
              <a
                className="ms-1 badge rounded-pill text-bg-warning"
                target="_blank"
                href="https://github.com/apache/incubator-answer/releases"
                rel="noreferrer">
                {t('update_to')} {remote_version}
              </a>
            )}
            {!isLatest && !remote_version && siteInfo.check_update && (
              <a
                className="ms-1 badge rounded-pill text-bg-danger"
                target="_blank"
                href="https://github.com/apache/incubator-answer/releases"
                rel="noreferrer">
                {t('check_failed')}
              </a>
            )}
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('run_mode')}</span>
            <strong>{data.login_required ? t('private') : t('public')}</strong>
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('upload_folder')}</span>
            <strong>
              {data.uploading_files ? t('writable') : t('not_writable')}
            </strong>
          </Col>
          <Col xs={6} className="mb-1">
            <span className="text-secondary me-1">{t('https')}</span>
            <strong>{data.https ? t('yes') : t('no')}</strong>
          </Col>
          <Col xs={6}>
            <span className="text-secondary me-1">{t('timezone')}</span>
            <strong>
              {data.time_zone.split('/')?.[1]?.replaceAll('_', ' ')}
            </strong>
          </Col>
          <Col xs={6}>
            <span className="text-secondary me-1">{t('smtp')}</span>
            {data.smtp !== 'not_configured' ? (
              <strong>{t(data.smtp)}</strong>
            ) : (
              <Link to="/admin/smtp" className="ms-2">
                {t('config')}
              </Link>
            )}
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );
};

export default HealthStatus;

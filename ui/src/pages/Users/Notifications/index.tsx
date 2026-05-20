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

import { useState, useEffect } from 'react';
import { Row, Col, ButtonGroup, Button, Nav } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate, Link } from 'react-router-dom';

import classNames from 'classnames';

import { usePageTags } from '@/hooks';
import {
  useQueryNotifications,
  clearUnreadNotification,
  clearNotificationStatus,
  readNotification,
} from '@/services';
import { floppyNavigation } from '@/utils';

import { Inbox, Achievements } from './components';
import './index.scss';

const PAGE_SIZE = 10;

const Notifications = () => {
  const [page, setPage] = useState(1);
  const [notificationData, setNotificationData] = useState<any>([]);
  const { t } = useTranslation('translation', { keyPrefix: 'notifications' });
  const inboxTypeNavs = ['all', 'posts', 'invites', 'votes'];
  const { type = 'inbox', subType = inboxTypeNavs[0] } = useParams();

  const queryParams: {
    type: string;
    inbox_type?: string;
    page: number;
    page_size: number;
  } = {
    type,
    page,
    page_size: PAGE_SIZE,
  };
  if (type === 'inbox') {
    queryParams.inbox_type = subType;
  }
  const { data, mutate } = useQueryNotifications(queryParams);

  useEffect(() => {
    clearNotificationStatus(type);
  }, []);

  useEffect(() => {
    if (!data) {
      return;
    }
    if (page > 1) {
      setNotificationData([...notificationData, ...(data?.list || [])]);
    } else {
      setNotificationData(data?.list);
    }
  }, [data]);
  const navigate = useNavigate();

  const handleTypeChange = (evt, val) => {
    if (!floppyNavigation.shouldProcessLinkClick(evt)) {
      return;
    }
    evt.preventDefault();
    if (type === val) {
      return;
    }
    setPage(1);
    setNotificationData([]);
    navigate(`/users/notifications/${val}`);
  };

  const handleLoadMore = () => {
    setPage(page + 1);
  };

  const handleUnreadNotification = async () => {
    await clearUnreadNotification(type);
    mutate();
  };

  const handleReadNotification = (id) => {
    readNotification(id);
  };
  usePageTags({
    title: t('notifications', { keyPrefix: 'page_title' }),
  });
  return (
    <Row className="pt-4 mb-5">
      <Col className="page-main flex-auto">
        <h3 className="mb-4">{t('title')}</h3>
        <div className="d-flex justify-content-between mb-3">
          <ButtonGroup size="sm">
            <Button
              as="a"
              href="/users/notifications/inbox"
              variant="outline-secondary"
              active={type === 'inbox'}
              onClick={(evt) => handleTypeChange(evt, 'inbox')}>
              {t('inbox')}
            </Button>
            <Button
              as="a"
              href="/users/notifications/achievement"
              variant="outline-secondary"
              active={type === 'achievement'}
              onClick={(evt) => handleTypeChange(evt, 'achievement')}>
              {t('achievement')}
            </Button>
          </ButtonGroup>
          <Button
            size="sm"
            variant="outline-secondary"
            onClick={handleUnreadNotification}>
            {t('all_read')}
          </Button>
        </div>
        {type === 'inbox' && (
          <>
            <Nav className="inbox-nav small">
              {inboxTypeNavs.map((nav) => {
                const navLinkHref = `/users/notifications/inbox/${nav}`;
                const navLinkName = t(`inbox_type.${nav}`);
                return (
                  <Nav.Item key={nav}>
                    <Link
                      to={navLinkHref}
                      onClick={() => {
                        setPage(1);
                      }}
                      className={classNames('nav-link', {
                        disabled: nav === subType,
                      })}>
                      {navLinkName}
                    </Link>
                  </Nav.Item>
                );
              })}
            </Nav>
            <Inbox
              data={notificationData}
              handleReadNotification={handleReadNotification}
            />
          </>
        )}
        {type === 'achievement' && (
          <Achievements
            data={notificationData}
            handleReadNotification={handleReadNotification}
          />
        )}
        {(data?.count || 0) > PAGE_SIZE * page && (
          <div className="d-flex justify-content-center align-items-center py-3">
            <Button
              variant="link"
              className="btn-no-border"
              onClick={handleLoadMore}>
              {t('show_more')}
            </Button>
          </div>
        )}
      </Col>
      <Col className="page-right-side" />
    </Row>
  );
};

export default Notifications;

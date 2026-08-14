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
import { useParams, useSearchParams, Link, Navigate } from 'react-router-dom';

import { usePageTags } from '@/hooks';
import { Pagination, FormatTime, Empty } from '@/components';
import { loggedUserInfoStore } from '@/stores';
import {
  usePersonalInfoByName,
  usePersonalTop,
  usePersonalListByTabName,
} from '@/services';
import type { UserInfoRes } from '@/common/interface';

import {
  UserInfo,
  NavBar,
  Overview,
  Alert,
  ListHead,
  DefaultList,
  Comments,
  Replies,
  Votes,
} from './components';

const Personal: FC = () => {
  const { tabName: routeTabName = 'overview', username = '' } = useParams();
  const legacyTabMap: Record<string, string> = {
    answers: 'comments',
    questions: 'posts',
  };
  const tabName =
    routeTabName === 'reputation' || routeTabName === 'badges'
      ? 'overview'
      : legacyTabMap[routeTabName] || routeTabName;
  const [searchParams] = useSearchParams();
  const page = searchParams.get('page') || 1;
  const order = searchParams.get('order') || 'newest';
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });
  const sessionUser = loggedUserInfoStore((state) => state.user);
  const isSelf = sessionUser?.username === username;

  const { data: userInfo } = usePersonalInfoByName(username);
  const { data: topData } = usePersonalTop(username, tabName);

  const { data: listData, isLoading = true } = usePersonalListByTabName(
    {
      username,
      page: Number(page),
      page_size: 30,
      order,
    },
    tabName,
  );
  const { count = 0, list = [] } = listData?.[tabName] || {};

  let pageTitle = '';
  if (userInfo?.username) {
    pageTitle = `${userInfo?.display_name} (${userInfo?.username})`;
  }
  usePageTags({
    title: pageTitle,
  });

  if (legacyTabMap[routeTabName]) {
    const query = searchParams.toString();
    const target = `/users/${username}/${legacyTabMap[routeTabName]}`;
    return <Navigate replace to={`${target}${query ? `?${query}` : ''}`} />;
  }

  return (
    <div className="pt-4 mb-5">
      <Row>
        <Col>
          {userInfo?.status !== 'normal' && userInfo?.status_msg && (
            <Alert data={userInfo?.status_msg} />
          )}
          <div className="d-md-flex d-block flex-wrap justify-content-between">
            <UserInfo data={userInfo as UserInfoRes} />
            {isSelf && (
              <div className="mb-3">
                <Link
                  className="btn btn-outline-secondary"
                  to="/users/settings/profile">
                  {t('edit_profile')}
                </Link>
              </div>
            )}
          </div>
          <NavBar tabName={tabName} slug={username} isSelf={isSelf} />

          <Overview
            visible={tabName === 'overview'}
            introduction={userInfo?.bio_html || ''}
            data={topData}
          />

          <ListHead
            count={count}
            sort={order}
            visible={tabName !== 'overview'}
            tabName={tabName}
          />
          <Comments data={list} visible={tabName === 'comments'} />
          <DefaultList
            data={list}
            tabName={tabName}
            visible={tabName === 'posts' || tabName === 'bookmarks'}
          />
          <Replies data={list} visible={tabName === 'replies'} />
          <Votes data={list} visible={tabName === 'votes'} />
          {!list?.length && !isLoading && <Empty />}

          {count > 0 && (
            <div className="d-flex justify-content-center py-4">
              <Pagination
                pageSize={30}
                totalSize={count || 0}
                currentPage={Number(page)}
              />
            </div>
          )}

          {tabName === 'overview' && (
            <>
              <h5 className="mb-3">{t('stats')}</h5>
              {userInfo?.created_at && (
                <div className="text-secondary">
                  <FormatTime time={userInfo.created_at} preFix={t('joined')} />
                  {t('comma')}{' '}
                  <FormatTime
                    time={userInfo.last_login_date}
                    preFix={t('last_login')}
                  />
                </div>
              )}
            </>
          )}
        </Col>
      </Row>
    </div>
  );
};
export default Personal;

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
import { Form, Table, Stack } from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import { Empty, Icon, Pagination, QueryGroup } from '@/components';
import * as Type from '@/common/interface';
import { useQueryBadges, updateBadgeStatus } from '@/services/admin/badges';
import { useToast } from '@/hooks';

import Action from './components/Action';

const BadgeFilterKeys: Type.BadgeFilterBy[] = ['all', 'active', 'inactive'];

const bgMap = {
  active: 'text-bg-success',
  inactive: 'text-bg-secondary',
};

const PAGE_SIZE = 10;

const Badges: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.badges' });

  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const curPage = Number(urlSearchParams.get('page') || '1');
  const curFilter = urlSearchParams.get('filter') || BadgeFilterKeys[0];
  const curQuery = urlSearchParams.get('query') || '';
  const Toast = useToast();

  const { data, isLoading, mutate } = useQueryBadges({
    page: curPage,
    page_size: PAGE_SIZE,
    ...(curQuery ? { q: curQuery } : {}),
    ...(curFilter === 'all' ? {} : { status: curFilter }),
  });

  const handleFilter = (e) => {
    urlSearchParams.set('query', e.target.value);
    urlSearchParams.delete('page');
    setUrlSearchParams(urlSearchParams);
  };

  const handleBadgeStatus = (badgeId, status) => {
    updateBadgeStatus({ id: badgeId, status }).then(() => {
      Toast.onShow({
        msg:
          status === 'inactive'
            ? t('badge_inactivated', { keyPrefix: 'messages' })
            : t('badge_activated', { keyPrefix: 'messages' }),
        variant: 'success',
      });
      mutate();
    });
  };

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <div className="d-flex flex-wrap justify-content-between align-items-center">
        <Stack direction="horizontal" gap={3} className="mb-3">
          <QueryGroup
            data={BadgeFilterKeys}
            currentSort={curFilter}
            sortKey="filter"
            i18nKeyPrefix="admin.badges"
          />
        </Stack>

        <Form.Control
          size="sm"
          type="search"
          value={curQuery}
          onChange={handleFilter}
          placeholder={t('filter.placeholder')}
          style={{ width: '12.25rem' }}
          className="mb-3"
        />
      </div>
      <Table responsive="md">
        <thead>
          <tr>
            <th className="min-w-15">{t('name')}</th>
            <th style={{ width: '19%' }}>{t('group')}</th>
            <th style={{ width: '14%' }}>{t('awards')}</th>
            <th style={{ width: '12%' }}>{t('status')}</th>
            <th className="text-end" style={{ width: '9%' }}>
              {t('action')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {data?.list?.map((badge) => (
            <tr key={badge.id}>
              <td className="d-flex align-items-center">
                {badge.icon?.startsWith('http') ? (
                  <img
                    src={badge.icon}
                    width={32}
                    height={32}
                    alt={badge.name}
                    className="me-3"
                  />
                ) : (
                  <Icon
                    name={badge?.icon}
                    size="32px"
                    className={classNames(
                      'lh-1 me-3',
                      badge?.level === 1 && 'bronze',
                      badge?.level === 2 && 'silver',
                      badge?.level === 3 && 'gold',
                    )}
                  />
                )}
                <div>
                  <Link to={`/badges/${badge.id}`}>{badge.name}</Link>
                  <div
                    className="text-body small"
                    dangerouslySetInnerHTML={{
                      __html: badge.description,
                    }}
                  />
                </div>
              </td>

              <td>{badge.group_name}</td>
              <td>
                <Link to={`/badges/${badge.id}`}>{badge.award_count}</Link>
              </td>
              <td>
                <span className={classNames('badge', bgMap[badge.status])}>
                  {t(badge.status)}
                </span>
              </td>
              <Action
                status={badge.status}
                onSelect={(status) => handleBadgeStatus(badge.id, status)}
              />
            </tr>
          ))}
        </tbody>
      </Table>
      {Number(data?.count) <= 0 && !isLoading && <Empty />}
      <div className="mt-4 mb-2 d-flex justify-content-center">
        <Pagination
          currentPage={curPage}
          totalSize={data?.count || 0}
          pageSize={PAGE_SIZE}
        />
      </div>
    </>
  );
};

export default Badges;

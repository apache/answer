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
import { Card, Table } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';

import { useQueryActivityLogDaily } from '@/services';

// What happened on each of the last 14 days; every number opens the
// activity log in a new tab, already filtered to that day and those actions
const COLUMNS: { key: string; actions: string }[] = [
  { key: 'questions', actions: 'question.create' },
  { key: 'answers', actions: 'answer.create' },
  { key: 'comments', actions: 'comment.create,comment.reply' },
  { key: 'reviews', actions: 'review.queued' },
  { key: 'badges', actions: 'badge.award' },
  { key: 'views', actions: 'page.view' },
];

const DailyActivity: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.dashboard.daily_activity',
  });
  const { data } = useQueryActivityLogDaily(14);
  const today = dayjs().format('YYYY-MM-DD');

  const logLink = (date: string, actions: string) =>
    `/admin/activity-log?range=custom&from=${date}&to=${date}&action=${encodeURIComponent(actions)}`;

  return (
    <Card className="mb-4">
      <Card.Body>
        <h6 className="mb-3">{t('title')}</h6>
        <Table size="sm" responsive className="mb-0 align-middle">
          <thead>
            <tr>
              <th>{t('day')}</th>
              {COLUMNS.map((c) => (
                <th key={c.key} className="text-end">
                  {t(c.key)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {(data || []).map((row) => (
              <tr key={row.date}>
                <td className="text-nowrap">
                  <a
                    href={logLink(row.date, '')}
                    target="_blank"
                    rel="noreferrer"
                    className={
                      row.date === today ? 'fw-bold' : 'link-secondary'
                    }>
                    {row.date}
                    {row.date === today ? ` (${t('today')})` : ''}
                  </a>
                </td>
                {COLUMNS.map((c) => {
                  const n = Number(row[c.key] || 0);
                  return (
                    <td key={c.key} className="text-end">
                      {n > 0 ? (
                        <a
                          href={logLink(row.date, c.actions)}
                          target="_blank"
                          rel="noreferrer"
                          title={t('open_log')}>
                          {n}
                        </a>
                      ) : (
                        <span className="text-secondary">0</span>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </Table>
      </Card.Body>
    </Card>
  );
};

export default DailyActivity;

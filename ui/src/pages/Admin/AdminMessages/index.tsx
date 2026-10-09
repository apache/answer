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

import { FC, useEffect, useState } from 'react';
import { Form, Table } from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';
import qs from 'qs';

import {
  BaseUserCard,
  Empty,
  Pagination,
  MessageUserButton,
} from '@/components';
import request from '@/utils/request';
import { useQueryAdminMessages, AdminMessageItem } from '@/services';

const PAGE_SIZE = 50;

// who got a message from an admin / moderator, when, and when they read it
const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.admin_messages',
  });
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const page = Number(urlSearchParams.get('page') || 1);
  const username = urlSearchParams.get('username') || '';
  const from = urlSearchParams.get('from') || '';
  const to = urlSearchParams.get('to') || '';
  const q = urlSearchParams.get('q') || '';

  const [qInput, setQInput] = useState(q);
  const [userInput, setUserInput] = useState(username);
  const [userOptions, setUserOptions] = useState<
    { username: string; display_name: string }[]
  >([]);
  const [expanded, setExpanded] = useState<string | null>(null);

  useEffect(() => setQInput(q), [q]);
  useEffect(() => setUserInput(username), [username]);

  const { data, isLoading, mutate } = useQueryAdminMessages({
    page,
    page_size: PAGE_SIZE,
    from: from ? dayjs(from).startOf('day').unix() : undefined,
    to: to ? dayjs(to).add(1, 'day').startOf('day').unix() : undefined,
    username,
    q,
  });

  const setParams = (patch: Record<string, string | undefined>) => {
    const next: Record<string, string> = {};
    urlSearchParams.forEach((v, k) => {
      next[k] = v;
    });
    Object.entries(patch).forEach(([k, v]) => {
      if (v === undefined || v === '') delete next[k];
      else next[k] = v;
    });
    delete next.page;
    setUrlSearchParams(next);
  };

  useEffect(() => {
    const term = userInput.trim();
    if (!term || term === username) {
      setUserOptions([]);
      return undefined;
    }
    const timer = setTimeout(() => {
      request
        .get(
          `/answer/admin/api/users/page?${qs.stringify({ page: 1, page_size: 8, query: term })}`,
        )
        .then((res: any) => setUserOptions(res?.list || []))
        .catch(() => setUserOptions([]));
    }, 300);
    return () => clearTimeout(timer);
  }, [userInput, username]);

  const contextLink = (m: AdminMessageItem) => {
    if (!m.question_id) return '';
    const base = `/questions/${m.question_id}/${m.url_title || ''}`;
    if (m.object_type === 'answer' && m.answer_id)
      return `${base}/${m.answer_id}`;
    if (m.object_type === 'comment')
      return `${base}${m.answer_id ? `/${m.answer_id}` : ''}?commentId=${m.object_id}`;
    return base;
  };

  return (
    <>
      <div className="d-flex flex-wrap justify-content-between align-items-center mb-4">
        <h3 className="mb-0">{t('title')}</h3>
        <MessageUserButton variant="button" onSent={() => mutate()} />
      </div>
      <div className="d-flex flex-wrap gap-3 align-items-end mb-3">
        <div>
          <Form.Label className="small text-secondary mb-1">
            {t('filter.user')}
          </Form.Label>
          <Form.Control
            size="sm"
            type="search"
            list="admin-messages-users"
            value={userInput}
            placeholder={t('filter.user_placeholder')}
            style={{ width: '12rem' }}
            onChange={(e) => setUserInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') setParams({ username: userInput.trim() });
            }}
            onBlur={() => {
              if (userInput.trim() !== username)
                setParams({ username: userInput.trim() });
            }}
          />
          <datalist id="admin-messages-users">
            {userOptions.map((u) => (
              <option key={u.username} value={u.username}>
                {u.display_name}
              </option>
            ))}
          </datalist>
        </div>
        <div>
          <Form.Label className="small text-secondary mb-1">
            {t('filter.range')}
          </Form.Label>
          <div className="d-flex gap-2 align-items-center">
            <Form.Control
              size="sm"
              type="date"
              value={from}
              max={to || undefined}
              onChange={(e) => setParams({ from: e.target.value })}
              style={{ width: '10rem' }}
            />
            <span className="text-secondary">–</span>
            <Form.Control
              size="sm"
              type="date"
              value={to}
              min={from || undefined}
              onChange={(e) => setParams({ to: e.target.value })}
              style={{ width: '10rem' }}
            />
          </div>
        </div>
        <div className="flex-grow-1">
          <Form.Label className="small text-secondary mb-1">
            {t('filter.search')}
          </Form.Label>
          <Form.Control
            size="sm"
            type="search"
            value={qInput}
            placeholder={t('filter.search_placeholder')}
            style={{ maxWidth: '20rem' }}
            onChange={(e) => setQInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') setParams({ q: qInput.trim() });
            }}
            onBlur={() => {
              if (qInput.trim() !== q) setParams({ q: qInput.trim() });
            }}
          />
        </div>
      </div>

      <div className="small text-secondary mb-2">
        {t('total', { count: Number(data?.count || 0) })}
      </div>

      <Table responsive="md" className="align-middle">
        <thead>
          <tr>
            <th style={{ width: '12%' }}>{t('col.sent')}</th>
            <th style={{ width: '14%' }}>{t('col.from')}</th>
            <th style={{ width: '18%' }}>{t('col.to')}</th>
            <th>{t('col.message')}</th>
            <th style={{ width: '12%' }}>{t('col.read')}</th>
          </tr>
        </thead>
        <tbody>
          {data?.list?.map((m) => {
            const link = contextLink(m);
            const open = expanded === m.id;
            return (
              <tr key={m.id}>
                <td className="text-nowrap small">
                  {dayjs.unix(m.created_at).format('YYYY-MM-DD HH:mm')}
                </td>
                <td>
                  <BaseUserCard
                    data={m.sender}
                    showReputation={false}
                    nameMaxWidth="140px"
                  />
                </td>
                <td>
                  <BaseUserCard
                    data={m.receiver}
                    showReputation={false}
                    nameMaxWidth="160px"
                  />
                  {m.receiver?.email ? (
                    <div className="small text-secondary text-break">
                      {m.receiver.email}
                    </div>
                  ) : null}
                </td>
                <td>
                  <div className="fw-bold text-break">{m.title}</div>
                  <div
                    className="text-break small"
                    style={open ? { whiteSpace: 'pre-wrap' } : undefined}
                    role="button"
                    tabIndex={0}
                    onClick={() => setExpanded(open ? null : m.id)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') setExpanded(open ? null : m.id);
                    }}>
                    {open || m.body.length <= 160
                      ? m.body
                      : `${m.body.slice(0, 160)}…`}
                  </div>
                  {link ? (
                    <div className="small">
                      <Link to={link}>
                        {t('context')}: {m.object_title || m.object_type}
                      </Link>
                    </div>
                  ) : null}
                </td>
                <td className="text-nowrap small">
                  {m.read_at ? (
                    <span className="text-success">
                      {dayjs.unix(m.read_at).format('YYYY-MM-DD HH:mm')}
                    </span>
                  ) : (
                    <span className="text-secondary">{t('unread')}</span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </Table>
      {Number(data?.count) <= 0 && !isLoading && <Empty />}
      <div className="mt-4 mb-2 d-flex justify-content-center">
        <Pagination
          currentPage={page}
          totalSize={data?.count || 0}
          pageSize={PAGE_SIZE}
        />
      </div>
    </>
  );
};

export default Index;

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

import { FC, useEffect, useMemo, useState } from 'react';
import {
  Badge,
  Button,
  ButtonGroup,
  Dropdown,
  Form,
  Stack,
  Table,
} from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';
import qs from 'qs';

import { BaseUserCard, Empty, Pagination, Icon } from '@/components';
import { toastStore } from '@/stores';
import request from '@/utils/request';
import {
  ActivityLogItem,
  ActivityLogParams,
  exportActivityLog,
  useQueryActivityLog,
  useQueryActivityLogActions,
} from '@/services';

const PAGE_SIZE = 50;
const PRESETS = ['today', 'yesterday', '7d', '30d', 'custom'] as const;
type Preset = (typeof PRESETS)[number];

// which kind of thing an action is, for the colour of its badge
const actionKind = (action: string) => {
  if (action === 'page.view') return 'view';
  if (action === 'user.login') return 'login';
  if (action.startsWith('badge.')) return 'badge';
  if (action.startsWith('review.') || action.endsWith('.flag'))
    return 'moderation';
  if (action === 'reputation.change') return 'reputation';
  if (
    action.includes('.vote') ||
    action.endsWith('.react') ||
    action.endsWith('.accept')
  )
    return 'vote';
  if (action.startsWith('user.')) return 'admin';
  return 'content';
};

const kindVariant: Record<string, string> = {
  view: 'light',
  login: 'secondary',
  badge: 'warning',
  moderation: 'danger',
  reputation: 'success',
  vote: 'info',
  admin: 'dark',
  content: 'primary',
};

// date range of a preset in the browser's time zone (from inclusive, to exclusive)
const presetRange = (preset: Preset): { from: number; to: number } | null => {
  const start = dayjs().startOf('day');
  switch (preset) {
    case 'today':
      return { from: start.unix(), to: start.add(1, 'day').unix() };
    case 'yesterday':
      return { from: start.subtract(1, 'day').unix(), to: start.unix() };
    case '7d':
      return {
        from: start.subtract(6, 'day').unix(),
        to: start.add(1, 'day').unix(),
      };
    case '30d':
      return {
        from: start.subtract(29, 'day').unix(),
        to: start.add(1, 'day').unix(),
      };
    default:
      return null;
  }
};

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.activity_log',
  });
  const { t: tGlobal } = useTranslation('translation');
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();

  const preset = (urlSearchParams.get('range') || 'today') as Preset;
  const customFrom = urlSearchParams.get('from') || '';
  const customTo = urlSearchParams.get('to') || '';
  const username = urlSearchParams.get('username') || '';
  const action = urlSearchParams.get('action') || '';
  const q = urlSearchParams.get('q') || '';
  const page = Number(urlSearchParams.get('page') || 1);

  const [qInput, setQInput] = useState(q);
  const [userInput, setUserInput] = useState(username);
  const [userOptions, setUserOptions] = useState<
    { username: string; display_name: string }[]
  >([]);
  const [exporting, setExporting] = useState(false);

  useEffect(() => setQInput(q), [q]);
  useEffect(() => setUserInput(username), [username]);

  const range = useMemo(() => {
    if (preset === 'custom') {
      const from = customFrom
        ? dayjs(customFrom).startOf('day').unix()
        : undefined;
      const to = customTo
        ? dayjs(customTo).add(1, 'day').startOf('day').unix()
        : undefined;
      return { from, to };
    }
    return presetRange(preset) || {};
  }, [preset, customFrom, customTo]);

  const params: ActivityLogParams = {
    page,
    page_size: PAGE_SIZE,
    from: range.from,
    to: range.to,
    username,
    action,
    q,
  };
  const { data, isLoading } = useQueryActivityLog(params);
  const { data: actions } = useQueryActivityLogActions({
    from: range.from,
    to: range.to,
  });

  const setParams = (
    patch: Record<string, string | undefined>,
    resetPage = true,
  ) => {
    const next: Record<string, string> = {};
    urlSearchParams.forEach((v, k) => {
      next[k] = v;
    });
    Object.entries(patch).forEach(([k, v]) => {
      if (v === undefined || v === '') delete next[k];
      else next[k] = v;
    });
    if (resetPage) delete next.page;
    setUrlSearchParams(next);
  };

  // username suggestions from the admin user list
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

  const selectedActions = action ? action.split(',') : [];
  const toggleAction = (key: string) => {
    const set = new Set(selectedActions);
    if (set.has(key)) set.delete(key);
    else set.add(key);
    setParams({ action: Array.from(set).join(',') });
  };

  const actionLabel = (key: string) =>
    t(`action.${key.replace(/\./g, '_')}`, { defaultValue: key });

  const handleExport = async () => {
    setExporting(true);
    try {
      await exportActivityLog({
        ...params,
        page: undefined,
        page_size: undefined,
      });
    } catch (e) {
      toastStore
        .getState()
        .show({ msg: t('export_failed'), variant: 'danger' });
    } finally {
      setExporting(false);
    }
  };

  const objectLink = (item: ActivityLogItem) => {
    const d = item.detail || {};
    const slug = item.url_title || '';
    switch (item.object_type) {
      case 'question':
        return item.question_id ? `/questions/${item.question_id}/${slug}` : '';
      case 'answer':
        return item.question_id
          ? `/questions/${item.question_id}/${slug}/${item.answer_id || item.object_id}`
          : '';
      case 'comment':
        return item.question_id
          ? `/questions/${item.question_id}/${slug}${item.answer_id ? `/${item.answer_id}` : ''}?commentId=${item.object_id}`
          : '';
      case 'page':
        return typeof d.path === 'string' ? d.path : '';
      case 'user':
        return item.target?.username
          ? `/users/${item.target.username}`
          : item.user?.username
            ? `/users/${item.user.username}`
            : '';
      case 'badge_award':
        return item.target?.username
          ? `/users/${item.target.username}/badges`
          : '';
      default:
        return '';
    }
  };

  const objectText = (item: ActivityLogItem) => {
    const d = item.detail || {};
    if (item.title) return item.title;
    if (item.object_type === 'page')
      return `${d.title || ''} ${d.path ? `(${d.path})` : ''}`.trim();
    if (item.object_type === 'badge_award' && d.badge)
      return tGlobal(d.badge, { defaultValue: d.badge });
    if (item.object_type === 'user')
      return item.target?.display_name || item.user?.display_name || '';
    return item.object_id && item.object_id !== '0' ? `#${item.object_id}` : '';
  };

  // the short facts shown under the object: reason, direction, method …
  const detailText = (item: ActivityLogItem) => {
    const d = item.detail || {};
    const parts: string[] = [];
    if (d.excerpt) parts.push(String(d.excerpt));
    if (d.reason) parts.push(`${t('detail.reason')}: ${d.reason}`);
    if (d.direction) parts.push(`${t('detail.direction')}: ${d.direction}`);
    if (d.method) parts.push(`${t('detail.method')}: ${d.method}`);
    if (d.activity) parts.push(String(d.activity));
    if (d.award_key && d.award_key !== 'admin' && d.award_key !== '0')
      parts.push(String(d.award_key));
    if (d.role_id) parts.push(`${t('detail.role')}: ${d.role_id}`);
    if (d.suspend_duration) parts.push(String(d.suspend_duration));
    if (d.reply_to) parts.push(`${t('detail.reply_to')} #${d.reply_to}`);
    if (item.ip) parts.push(item.ip);
    return parts.join(' · ');
  };

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <div className="d-flex flex-wrap gap-3 align-items-end mb-3">
        <div>
          <Form.Label className="small text-secondary mb-1">
            {t('filter.range')}
          </Form.Label>
          <div className="d-flex flex-wrap gap-2 align-items-center">
            <ButtonGroup size="sm">
              {PRESETS.map((p) => (
                <Button
                  key={p}
                  variant={preset === p ? 'primary' : 'outline-secondary'}
                  onClick={() =>
                    setParams({ range: p === 'today' ? undefined : p })
                  }>
                  {t(`range.${p}`)}
                </Button>
              ))}
            </ButtonGroup>
            {preset === 'custom' && (
              <>
                <Form.Control
                  size="sm"
                  type="date"
                  value={customFrom}
                  max={customTo || undefined}
                  onChange={(e) => setParams({ from: e.target.value })}
                  style={{ width: '10rem' }}
                />
                <span className="text-secondary">–</span>
                <Form.Control
                  size="sm"
                  type="date"
                  value={customTo}
                  min={customFrom || undefined}
                  onChange={(e) => setParams({ to: e.target.value })}
                  style={{ width: '10rem' }}
                />
              </>
            )}
          </div>
        </div>
        <div>
          <Form.Label className="small text-secondary mb-1">
            {t('filter.user')}
          </Form.Label>
          <Form.Control
            size="sm"
            type="search"
            list="activity-log-users"
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
          <datalist id="activity-log-users">
            {userOptions.map((u) => (
              <option key={u.username} value={u.username}>
                {u.display_name}
              </option>
            ))}
          </datalist>
        </div>
        <div>
          <Form.Label className="small text-secondary mb-1">
            {t('filter.action')}
          </Form.Label>
          <Dropdown autoClose="outside">
            <Dropdown.Toggle size="sm" variant="outline-secondary">
              {selectedActions.length
                ? t('filter.action_selected', { count: selectedActions.length })
                : t('filter.action_all')}
            </Dropdown.Toggle>
            <Dropdown.Menu
              style={{
                maxHeight: '22rem',
                overflowY: 'auto',
                minWidth: '18rem',
              }}>
              {selectedActions.length > 0 && (
                <Dropdown.Item
                  as="button"
                  onClick={() => setParams({ action: undefined })}>
                  {t('filter.action_clear')}
                </Dropdown.Item>
              )}
              {(actions || []).map((a) => (
                <div key={a.action} className="dropdown-item">
                  <Form.Check
                    type="checkbox"
                    id={`act-${a.action}`}
                    checked={selectedActions.includes(a.action)}
                    onChange={() => toggleAction(a.action)}
                    label={
                      <span>
                        {actionLabel(a.action)}{' '}
                        <span className="text-secondary small">
                          ({a.count})
                        </span>
                      </span>
                    }
                  />
                </div>
              ))}
              {!actions?.length && (
                <div className="dropdown-item text-secondary small">
                  {t('filter.action_none')}
                </div>
              )}
            </Dropdown.Menu>
          </Dropdown>
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
        <div>
          <Button
            size="sm"
            variant="outline-primary"
            disabled={exporting}
            onClick={handleExport}>
            <Icon name="download" className="me-1" />
            {exporting ? t('exporting') : t('export')}
          </Button>
        </div>
      </div>

      <div className="small text-secondary mb-2">
        {t('total', { count: Number(data?.count || 0) })}
      </div>

      <Table responsive="md" className="align-middle">
        <thead>
          <tr>
            <th style={{ width: '11%' }}>{t('col.time')}</th>
            <th style={{ width: '16%' }}>{t('col.user')}</th>
            <th style={{ width: '13%' }}>{t('col.action')}</th>
            <th>{t('col.object')}</th>
            <th style={{ width: '14%' }}>{t('col.target')}</th>
            <th style={{ width: '7%' }} className="text-end">
              {t('col.points')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {data?.list?.map((item) => {
            const link = objectLink(item);
            const text = objectText(item);
            const facts = detailText(item);
            return (
              <tr key={item.id}>
                <td className="text-nowrap small">
                  {dayjs.unix(item.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </td>
                <td>
                  {item.user?.id === '0' ? (
                    <span className="small text-secondary">
                      <Icon name="robot" className="me-1" />
                      {t('system')}
                    </span>
                  ) : (
                    <BaseUserCard
                      data={item.user}
                      showReputation={false}
                      nameMaxWidth="160px"
                    />
                  )}
                </td>
                <td>
                  <Badge
                    bg={kindVariant[actionKind(item.action)]}
                    text={
                      ['light', 'warning', 'info'].includes(
                        kindVariant[actionKind(item.action)],
                      )
                        ? 'dark'
                        : undefined
                    }
                    className="fw-normal">
                    {actionLabel(item.action)}
                  </Badge>
                </td>
                <td>
                  {link ? (
                    <Link to={link} className="text-break">
                      {text}
                    </Link>
                  ) : (
                    <span className="text-break">{text}</span>
                  )}
                  {facts && (
                    <div className="small text-secondary text-break">
                      {facts}
                    </div>
                  )}
                </td>
                <td>
                  {item.target ? (
                    <BaseUserCard
                      data={item.target}
                      showReputation={false}
                      nameMaxWidth="140px"
                    />
                  ) : null}
                </td>
                <td
                  className={`text-end fw-bold ${item.rank_delta > 0 ? 'text-success' : item.rank_delta < 0 ? 'text-danger' : 'text-secondary'}`}>
                  {item.rank_delta
                    ? item.rank_delta > 0
                      ? `+${item.rank_delta}`
                      : item.rank_delta
                    : ''}
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
      <Stack direction="horizontal" className="small text-secondary">
        {t('hint')}
      </Stack>
    </>
  );
};

export default Index;

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
import { Form, Table, Stack, Button } from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import {
  FormatTime,
  Icon,
  Pagination,
  BaseUserCard,
  Empty,
  QueryGroup,
  Modal,
} from '@/components';
import { ADMIN_LIST_STATUS } from '@/common/constants';
import * as Type from '@/common/interface';
import { deletePermanently, useQuestionSearch } from '@/services';
import { pathFactory } from '@/router/pathFactory';
import { toastStore } from '@/stores';

import Action from './components/Action';

const questionFilterItems: Type.AdminContentsFilterBy[] = [
  'normal',
  'pending',
  'closed',
  'deleted',
];

const PAGE_SIZE = 20;
const Questions: FC = () => {
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const curFilter = urlSearchParams.get('status') || questionFilterItems[0];
  const curPage = Number(urlSearchParams.get('page')) || 1;
  const curQuery = urlSearchParams.get('query') || '';
  const { t } = useTranslation('translation', { keyPrefix: 'admin.questions' });

  const {
    data: listData,
    isLoading,
    mutate: refreshList,
  } = useQuestionSearch({
    page_size: PAGE_SIZE,
    page: curPage,
    status: curFilter as Type.AdminContentsFilterBy,
    query: curQuery,
  });
  const count = listData?.count || 0;

  const handleDeletePermanently = () => {
    Modal.confirm({
      title: t('title', { keyPrefix: 'delete_permanently' }),
      content: t('content', { keyPrefix: 'delete_permanently' }),
      cancelBtnVariant: 'link',
      confirmText: t('ok', { keyPrefix: 'btns' }),
      onConfirm: () => {
        deletePermanently('questions').then(() => {
          toastStore.getState().show({
            msg: t('posts_deleted', { keyPrefix: 'messages' }),
            variant: 'success',
          });
          refreshList();
        });
      },
    });
  };

  const handleFilter = (e) => {
    urlSearchParams.set('query', e.target.value);
    urlSearchParams.delete('page');
    setUrlSearchParams(urlSearchParams);
  };
  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <div className="d-flex flex-wrap justify-content-between align-items-center mb-3">
        <Stack direction="horizontal" gap={3}>
          <QueryGroup
            data={questionFilterItems}
            currentSort={curFilter}
            sortKey="status"
            i18nKeyPrefix="btns"
          />
          {curFilter === 'deleted' ? (
            <Button
              variant="outline-danger"
              size="sm"
              onClick={() => handleDeletePermanently()}>
              {t('deleted_permanently', { keyPrefix: 'btns' })}
            </Button>
          ) : null}
        </Stack>

        <Form.Control
          value={curQuery}
          size="sm"
          type="search"
          placeholder={t('filter.placeholder')}
          onChange={handleFilter}
          style={{ width: '12.25rem' }}
          className="mt-3 mt-sm-0"
        />
      </div>
      <Table responsive="md">
        <thead>
          <tr>
            <th className="min-w-15">{t('post')}</th>
            <th style={{ width: '8%' }}>{t('votes')}</th>
            <th style={{ width: '8%' }}>{t('answers')}</th>
            <th style={{ width: '15%' }}>{t('created')}</th>
            <th style={{ width: '14%' }}>{t('status')}</th>
            <th style={{ width: '10%' }} className="text-end">
              {t('action')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {listData?.list?.map((li) => {
            return (
              <tr key={li.id}>
                <td>
                  <Link
                    to={pathFactory.questionLanding(li.id, li.url_title)}
                    target="_blank"
                    className="text-break text-wrap"
                    rel="noreferrer">
                    {li.title}
                  </Link>
                  {li.accepted_answer_id > 0 && (
                    <Icon
                      name="check-circle-fill"
                      className="ms-2 text-success"
                    />
                  )}
                </td>
                <td>{li.vote_count}</td>
                <td>
                  <Link
                    to={`/admin/answers?questionId=${li.id}`}
                    rel="noreferrer">
                    {li.answer_count}
                  </Link>
                </td>
                <td>
                  <Stack>
                    <BaseUserCard data={li.user_info} nameMaxWidth="130px" />
                    <FormatTime
                      className="small text-secondary"
                      time={li.create_time}
                    />
                  </Stack>
                </td>
                <td>
                  <span
                    className={classNames(
                      'badge',
                      'me-1',
                      'mb-1',
                      ADMIN_LIST_STATUS[curFilter]?.variant,
                    )}>
                    {t(ADMIN_LIST_STATUS[curFilter]?.name, {
                      keyPrefix: 'btns',
                    })}
                  </span>
                  {li.show === 2 && (
                    <span
                      className={classNames(
                        'badge',
                        ADMIN_LIST_STATUS.unlisted.variant,
                      )}>
                      {t(ADMIN_LIST_STATUS.unlisted.name, {
                        keyPrefix: 'btns',
                      })}
                    </span>
                  )}
                </td>

                <td className="text-end">
                  <Action
                    itemData={{ id: li.id, answer_count: li.answer_count }}
                    refreshList={refreshList}
                    curFilter={curFilter}
                    show={li.show}
                    pin={li.pin}
                  />
                </td>
              </tr>
            );
          })}
        </tbody>
      </Table>
      {Number(count) <= 0 && !isLoading && <Empty />}
      <div className="mt-4 mb-2 d-flex justify-content-center">
        <Pagination
          currentPage={curPage}
          totalSize={count}
          pageSize={PAGE_SIZE}
        />
      </div>
    </>
  );
};

export default Questions;

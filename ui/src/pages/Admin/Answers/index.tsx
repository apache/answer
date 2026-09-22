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
import { Form, Table, Stack, Button } from 'react-bootstrap';
import { useSearchParams, Link } from 'react-router-dom';
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
  TabNav,
} from '@/components';
import { ADMIN_LIST_STATUS, ADMIN_QA_NAV_MENUS } from '@/common/constants';
import * as Type from '@/common/interface';
import { deleteAnswers, deletePermanently, useAnswerSearch } from '@/services';
import { escapeRemove } from '@/utils';
import { pathFactory } from '@/router/pathFactory';
import { toastStore } from '@/stores';

import AnswerAction from './components/Action';

const answerFilterItems: Type.AdminContentsFilterBy[] = [
  'normal',
  'pending',
  'deleted',
];

const Answers: FC = () => {
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const curFilter = urlSearchParams.get('status') || answerFilterItems[0];
  const PAGE_SIZE = 20;
  const curPage = Number(urlSearchParams.get('page')) || 1;
  const curQuery = urlSearchParams.get('query') || '';
  const questionId = urlSearchParams.get('questionId') || '';
  const [selectedAnswerIDs, setSelectedAnswerIDs] = useState<string[]>([]);
  const { t } = useTranslation('translation', { keyPrefix: 'admin.answers' });

  const {
    data: listData,
    isLoading,
    mutate: refreshList,
  } = useAnswerSearch({
    page_size: PAGE_SIZE,
    page: curPage,
    status: curFilter as Type.AdminContentsFilterBy,
    query: curQuery,
    question_id: questionId,
  });
  const count = listData?.count || 0;
  const canBulkDelete = curFilter === 'normal';
  const pageAnswerIDs = listData?.list?.map((item) => item.id) || [];
  const allAnswersSelected =
    pageAnswerIDs.length > 0 &&
    pageAnswerIDs.every((answerID) => selectedAnswerIDs.includes(answerID));

  useEffect(() => {
    setSelectedAnswerIDs([]);
  }, [curFilter, curPage, curQuery, questionId]);

  const handleDeletePermanently = () => {
    Modal.confirm({
      title: t('title', { keyPrefix: 'delete_permanently' }),
      content: t('content', { keyPrefix: 'delete_permanently' }),
      cancelBtnVariant: 'link',
      confirmText: t('delete', { keyPrefix: 'btns' }),
      confirmBtnVariant: 'danger',
      onConfirm: () => {
        deletePermanently('answers').then(() => {
          toastStore.getState().show({
            msg: t('answers_deleted', { keyPrefix: 'messages' }),
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

  const toggleAnswer = (answerID: string) => {
    setSelectedAnswerIDs((selectedIDs) =>
      selectedIDs.includes(answerID)
        ? selectedIDs.filter((id) => id !== answerID)
        : [...selectedIDs, answerID],
    );
  };

  const toggleAllAnswers = () => {
    setSelectedAnswerIDs(allAnswersSelected ? [] : pageAnswerIDs);
  };

  const handleBulkDelete = () => {
    Modal.confirm({
      title: t('bulk_delete.title', { count: selectedAnswerIDs.length }),
      content: t('bulk_delete.content', { count: selectedAnswerIDs.length }),
      cancelBtnVariant: 'link',
      confirmText: t('delete', { keyPrefix: 'btns' }),
      confirmBtnVariant: 'danger',
      onConfirm: () => {
        deleteAnswers(selectedAnswerIDs).then((result) => {
          const failedCount = result.failed_ids.length;
          toastStore.getState().show({
            msg:
              failedCount > 0
                ? t('bulk_delete.partial', {
                    succeeded: result.succeeded_ids.length,
                    failed: failedCount,
                  })
                : t('bulk_delete.success', {
                    count: result.succeeded_ids.length,
                  }),
            variant: failedCount > 0 ? 'warning' : 'success',
          });
          setSelectedAnswerIDs([]);
          refreshList();
        });
      },
    });
  };

  return (
    <>
      <h3 className="mb-4">
        {t('page_title', { keyPrefix: 'admin.questions' })}
      </h3>
      <TabNav menus={ADMIN_QA_NAV_MENUS} />
      <div className="d-flex flex-wrap justify-content-between align-items-center">
        <Stack direction="horizontal" gap={3} className="mb-3">
          <QueryGroup
            data={answerFilterItems}
            currentSort={curFilter}
            sortKey="status"
            i18nKeyPrefix="btns"
          />
          {curFilter === 'deleted' && count > 0 ? (
            <Button
              variant="outline-danger"
              size="sm"
              onClick={() => handleDeletePermanently()}>
              {t('deleted_permanently', { keyPrefix: 'btns' })}
            </Button>
          ) : null}
          {canBulkDelete ? (
            <Button
              variant="outline-danger"
              size="sm"
              disabled={selectedAnswerIDs.length === 0}
              onClick={handleBulkDelete}>
              {t('bulk_delete.action', { count: selectedAnswerIDs.length })}
            </Button>
          ) : null}
        </Stack>

        <Form.Control
          value={curQuery}
          onChange={handleFilter}
          size="sm"
          type="search"
          placeholder={t('filter.placeholder')}
          style={{ width: '12.25rem' }}
          className="mb-3"
        />
      </div>
      <Table responsive="md">
        <thead>
          <tr>
            {canBulkDelete ? (
              <th style={{ width: '1%' }}>
                <Form.Check.Input
                  type="checkbox"
                  checked={allAnswersSelected}
                  onChange={toggleAllAnswers}
                  aria-label={t('bulk_delete.select_all')}
                />
              </th>
            ) : null}
            <th className="min-w-15">{t('post')}</th>
            <th style={{ width: '11%' }}>{t('votes')}</th>
            <th style={{ width: '14%' }}>{t('created')}</th>
            <th style={{ width: '11%' }}>{t('status')}</th>
            <th style={{ width: '11%' }} className="text-end">
              {t('action')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {listData?.list?.map((li) => {
            return (
              <tr key={li.id}>
                {canBulkDelete ? (
                  <td>
                    <Form.Check.Input
                      type="checkbox"
                      checked={selectedAnswerIDs.includes(li.id)}
                      onChange={() => toggleAnswer(li.id)}
                      aria-label={t('bulk_delete.select')}
                    />
                  </td>
                ) : null}
                <td>
                  <Link
                    to={pathFactory.answerLanding({
                      questionId: li.question_id,
                      slugTitle: li.question_info.url_title,
                      answerId: li.id,
                    })}
                    target="_blank"
                    className="text-break text-wrap"
                    rel="noreferrer">
                    {li.question_info.title}
                  </Link>
                  {li.accepted === 2 && (
                    <Icon
                      name="check-circle-fill"
                      className="ms-2 text-success"
                    />
                  )}
                  <div className="text-truncate-2 small max-w-30">
                    {escapeRemove(li.description)}
                  </div>
                </td>
                <td>{li.vote_count}</td>
                <td>
                  <Stack>
                    <BaseUserCard
                      avatarSize="20"
                      data={li.user_info}
                      nameMaxWidth="200px"
                    />

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
                      ADMIN_LIST_STATUS[curFilter]?.variant,
                    )}>
                    {t(ADMIN_LIST_STATUS[curFilter]?.name, {
                      keyPrefix: 'btns',
                    })}
                  </span>
                </td>
                <td className="text-end">
                  <AnswerAction
                    itemData={{ id: li.id, accepted: li.accepted }}
                    curFilter={curFilter}
                    refreshList={refreshList}
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

export default Answers;

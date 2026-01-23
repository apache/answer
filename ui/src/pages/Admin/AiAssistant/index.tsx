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

import { useState } from 'react';
import { Table, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { BaseUserCard, FormatTime, Pagination, Empty } from '@/components';
import { useQueryAdminConversationList } from '@/services';

import DetailModal from './components/DetailModal';
import Action from './components/Action';

const Index = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.conversations',
  });
  const [urlSearchParams] = useSearchParams();
  const curPage = Number(urlSearchParams.get('page') || '1');
  const PAGE_SIZE = 20;
  const [detailModal, setDetailModal] = useState({
    visible: false,
    id: '',
  });
  const {
    data: conversations,
    isLoading,
    mutate: refreshList,
  } = useQueryAdminConversationList({
    page: curPage,
    page_size: PAGE_SIZE,
  });

  const handleShowDetailModal = (data) => {
    setDetailModal({
      visible: true,
      id: data.id,
    });
  };

  const handleHideDetailModal = () => {
    setDetailModal({
      visible: false,
      id: '',
    });
  };

  return (
    <div className="d-flex flex-column flex-grow-1 position-relative">
      <h3 className="mb-4">{t('ai_assistant', { keyPrefix: 'nav_menus' })}</h3>
      <Table responsive="md">
        <thead>
          <tr>
            <th className="min-w-15">{t('topic')}</th>
            <th style={{ width: '10%' }}>{t('helpful')}</th>
            <th style={{ width: '10%' }}>{t('unhelpful')}</th>
            <th style={{ width: '20%' }}>{t('created')}</th>
            <th style={{ width: '10%' }} className="text-end">
              {t('action')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {conversations?.list.map((item) => {
            return (
              <tr key={item.id}>
                <td>
                  <Button
                    variant="link"
                    className="p-0 text-decoration-none text-truncate max-w-30"
                    onClick={() => handleShowDetailModal(item)}>
                    {item.topic}
                  </Button>
                </td>
                <td>{item.helpful_count}</td>
                <td>{item.unhelpful_count}</td>
                <td>
                  <div className="vstack">
                    <BaseUserCard data={item.user_info} avatarSize="20px" />
                    <FormatTime
                      className="small text-secondary"
                      time={item.created_at}
                    />
                  </div>
                </td>
                <td className="text-end">
                  <Action id={item.id} refreshList={refreshList} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </Table>
      {!isLoading && Number(conversations?.count) <= 0 && (
        <Empty>{t('empty')}</Empty>
      )}

      <div className="mt-4 mb-2 d-flex justify-content-center">
        <Pagination
          currentPage={curPage}
          totalSize={conversations?.count || 0}
          pageSize={PAGE_SIZE}
        />
      </div>
      <DetailModal
        visible={detailModal.visible}
        id={detailModal.id}
        onClose={handleHideDetailModal}
      />
    </div>
  );
};
export default Index;

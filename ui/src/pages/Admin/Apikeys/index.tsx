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
import { useTranslation } from 'react-i18next';
import { Button, Table } from 'react-bootstrap';

import dayjs from 'dayjs';

import { useQueryApiKeys } from '@/services';

import { Action, AddOrEditModal, CreatedModal } from './components';

const Index = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.apikeys',
  });
  const [showModal, setShowModal] = useState({
    visible: false,
    item: null,
  });
  const [showCreatedModal, setShowCreatedModal] = useState({
    visible: false,
    api_key: '',
  });
  const { data: apiKeysList, mutate: refreshList } = useQueryApiKeys();

  const handleAddModalState = (bol, item) => {
    setShowModal({
      visible: bol,
      item,
    });
  };

  const handleCreatedModalState = (visible, api_key) => {
    setShowCreatedModal({
      visible,
      api_key,
    });
  };

  const addOrEditCallback = (type, key) => {
    handleAddModalState(false, null);
    refreshList();
    if (type === 'add') {
      handleCreatedModalState(true, key);
    }
  };

  return (
    <div>
      <h3 className="mb-4">{t('title')}</h3>
      <Button
        variant="outline-primary mb-3"
        size="sm"
        onClick={() => handleAddModalState(true, null)}>
        {t('add_api_key')}
      </Button>
      <Table responsive="md">
        <thead className="c-table">
          <tr>
            <th style={{ width: '20%' }}>{t('desc')}</th>
            <th style={{ width: '11%' }}>{t('scope')}</th>
            <th style={{ minWidth: '200px' }}>{t('key')}</th>
            <th style={{ width: '18%' }}>{t('created')}</th>
            <th style={{ width: '18%' }}>{t('last_used')}</th>
            <th className="text-end" style={{ width: '10%' }}>
              {t('action', { keyPrefix: 'admin.questions' })}
            </th>
          </tr>
          {apiKeysList?.map((item) => {
            return (
              <tr key={item.id}>
                <td>{item.description}</td>
                <td>
                  {t(item.scope, {
                    keyPrefix: 'admin.apikeys.add_or_edit_modal',
                  })}
                </td>
                <td>{item.access_key}</td>
                <td>
                  {dayjs
                    .unix(item?.created_at)
                    .tz()
                    .format(t('long_date_with_time', { keyPrefix: 'dates' }))}
                </td>
                <td>
                  {item?.last_used_at &&
                    dayjs
                      .unix(item?.last_used_at)
                      .tz()
                      .format(t('long_date_with_time', { keyPrefix: 'dates' }))}
                </td>
                <td className="text-end">
                  <Action
                    itemData={item}
                    showModal={() => handleAddModalState(true, item)}
                    refreshList={refreshList}
                  />
                </td>
              </tr>
            );
          })}
        </thead>
      </Table>

      <AddOrEditModal
        data={showModal.item}
        visible={showModal.visible}
        onClose={handleAddModalState}
        callback={addOrEditCallback}
      />
      <CreatedModal
        visible={showCreatedModal.visible}
        api_key={showCreatedModal.api_key}
        onClose={() => handleCreatedModalState(false, '')}
      />
    </div>
  );
};

export default Index;

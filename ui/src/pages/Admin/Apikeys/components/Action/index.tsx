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

import { Dropdown } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Icon, Modal } from '@/components';
import { deleteApiKey } from '@/services';
import { toastStore } from '@/stores';

const ApiActions = ({ itemData, refreshList, showModal }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.apikeys.delete_modal',
  });

  const handleAction = (type) => {
    if (type === 'delete') {
      Modal.confirm({
        title: t('title'),
        content: t('content'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          deleteApiKey(itemData.id).then(() => {
            toastStore.getState().show({
              msg: t('api_key_deleted', { keyPrefix: 'messages' }),
              variant: 'success',
            });
            refreshList();
          });
        },
      });
    }

    if (type === 'edit') {
      showModal(true);
    }
  };

  return (
    <Dropdown>
      <Dropdown.Toggle variant="link" className="no-toggle p-0">
        <Icon
          name="three-dots-vertical"
          title={t('action', { keyPrefix: 'admin.answers' })}
        />
      </Dropdown.Toggle>
      <Dropdown.Menu align="end">
        <Dropdown.Item onClick={() => handleAction('edit')}>
          {t('edit', { keyPrefix: 'btns' })}
        </Dropdown.Item>
        <Dropdown.Item onClick={() => handleAction('delete')}>
          {t('delete', { keyPrefix: 'btns' })}
        </Dropdown.Item>
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default ApiActions;

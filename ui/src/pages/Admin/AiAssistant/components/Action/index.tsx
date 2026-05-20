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

import { Modal, Icon } from '@/components';
import { deleteAdminConversation } from '@/services';
import { useToast } from '@/hooks';

interface Props {
  id: string;
  refreshList?: () => void;
}
const ConversationsOperation = ({ id, refreshList }: Props) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.conversations',
  });
  const toast = useToast();

  const handleAction = (eventKey: string | null) => {
    if (eventKey === 'delete') {
      Modal.confirm({
        title: t('delete_modal.title'),
        content: t('delete_modal.content'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          deleteAdminConversation(id).then(() => {
            refreshList?.();
            toast.onShow({
              variant: 'success',
              msg: t('delete_modal.delete_success'),
            });
          });
        },
      });
    }
  };

  return (
    <Dropdown onSelect={handleAction}>
      <Dropdown.Toggle variant="link" className="no-toggle p-0 lh-1">
        <Icon name="three-dots-vertical" title={t('action')} />
      </Dropdown.Toggle>
      <Dropdown.Menu align="end">
        <Dropdown.Item eventKey="delete">
          {t('delete', { keyPrefix: 'btns' })}
        </Dropdown.Item>
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default ConversationsOperation;

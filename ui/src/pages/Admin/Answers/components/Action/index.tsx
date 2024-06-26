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
import { Link } from 'react-router-dom';

import { Icon, Modal } from '@/components';
import { changeAnswerStatus } from '@/services';

const AnswerActions = ({ itemData, curFilter, refreshList }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'delete' });

  const handleAction = (type) => {
    if (type === 'delete') {
      Modal.confirm({
        title: t('title'),
        content: itemData.accepted === 2 ? t('answer_accepted') : t('other'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          changeAnswerStatus(itemData.id, 'deleted').then(() => {
            refreshList();
          });
        },
      });
    }

    if (type === 'undelete') {
      Modal.confirm({
        title: t('undelete_title'),
        content: t('undelete_desc'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('undelete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          changeAnswerStatus(itemData.id, 'available').then(() => {
            refreshList();
          });
        },
      });
    }
  };

  if (curFilter === 'pending') {
    return (
      <Link
        to={`/review?type=queued_post&objectId=${itemData.id}`}
        className="btn btn-link p-0"
        title={t('review', { keyPrefix: 'header.nav' })}>
        <Icon name="three-dots-vertical" />
      </Link>
    );
  }

  return (
    <Dropdown>
      <Dropdown.Toggle variant="link" className="no-toggle p-0">
        <Icon
          name="three-dots-vertical"
          title={t('action', { keyPrefix: 'admin.answers' })}
        />
      </Dropdown.Toggle>
      <Dropdown.Menu>
        {curFilter === 'deleted' ? (
          <Dropdown.Item onClick={() => handleAction('undelete')}>
            {t('undelete', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        ) : (
          <Dropdown.Item onClick={() => handleAction('delete')}>
            {t('delete', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        )}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default AnswerActions;

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
import {
  changeQuestionStatus,
  questionOperation,
  reopenQuestion,
} from '@/services';
import { useReportModal, useToast } from '@/hooks';

const AnswerActions = ({ itemData, refreshList, curFilter, show, pin }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'delete' });
  const closeModal = useReportModal(refreshList);
  const toast = useToast();

  const handleAction = (type) => {
    if (type === 'delete') {
      Modal.confirm({
        title: t('title', { keyPrefix: 'delete' }),
        content:
          itemData.answer_count > 0
            ? t('question', { keyPrefix: 'delete' })
            : t('other', { keyPrefix: 'delete' }),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          changeQuestionStatus(itemData.id, 'deleted').then(() => {
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
          changeQuestionStatus(itemData.id, 'available').then(() => {
            refreshList();
          });
        },
      });
    }

    if (type === 'close') {
      closeModal.onShow({
        type: 'question',
        id: itemData.id,
        action: 'close',
      });
    }

    if (type === 'reopen') {
      Modal.confirm({
        title: t('title', { keyPrefix: 'question_detail.reopen' }),
        content: t('content', { keyPrefix: 'question_detail.reopen' }),
        cancelBtnVariant: 'link',
        confirmText: t('confirm_btn', { keyPrefix: 'question_detail.reopen' }),
        onConfirm: () => {
          reopenQuestion({
            question_id: itemData.id,
          }).then(() => {
            toast.onShow({
              msg: t('post_reopen', { keyPrefix: 'messages' }),
              variant: 'success',
            });
            refreshList();
          });
        },
      });
    }

    if (type === 'list' || type === 'unlist') {
      const keyPrefix =
        type === 'list' ? 'question_detail.list' : 'question_detail.unlist';
      Modal.confirm({
        title: t('title', { keyPrefix }),
        content: t('content', { keyPrefix }),
        cancelBtnVariant: 'link',
        confirmText: t('confirm_btn', { keyPrefix }),
        onConfirm: () => {
          questionOperation({
            id: itemData.id,
            operation: type === 'list' ? 'show' : 'hide',
          }).then(() => {
            toast.onShow({
              msg: t(`post_${type}`, { keyPrefix: 'messages' }),
              variant: 'success',
            });
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
        {curFilter === 'normal' && (
          <Dropdown.Item onClick={() => handleAction('close')}>
            {t('close', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        )}
        {curFilter === 'closed' && (
          <Dropdown.Item onClick={() => handleAction('reopen')}>
            {t('reopen', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        )}
        {curFilter !== 'deleted' ? (
          <Dropdown.Item onClick={() => handleAction('delete')}>
            {t('delete', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        ) : (
          <Dropdown.Item onClick={() => handleAction('undelete')}>
            {t('undelete', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        )}
        {show === 2 ? (
          <Dropdown.Item onClick={() => handleAction('list')}>
            {t('list', { keyPrefix: 'btns' })}
          </Dropdown.Item>
        ) : (
          pin !== 2 && (
            <Dropdown.Item onClick={() => handleAction('unlist')}>
              {t('unlist', { keyPrefix: 'btns' })}
            </Dropdown.Item>
          )
        )}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default AnswerActions;

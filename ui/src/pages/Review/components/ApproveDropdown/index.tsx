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

import { FC, useState } from 'react';
import { Dropdown, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Modal } from '@/components';
import { putFlagReviewAction } from '@/services';
import { useCaptchaModal, useReportModal, useToast } from '@/hooks';
import type * as Type from '@/common/interface';
import EditPostModal from '../EditPostModal';

interface IProps {
  itemData: Type.FlagReviewItem | null;
  curFilter: string;
  objectType: Type.FlagReviewItem['object_type'] | '';
  approveCallback: () => void;
}

const Index: FC<IProps> = ({
  itemData,
  objectType,
  curFilter,
  approveCallback,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_review' });

  const [isLoading, setIsLoading] = useState(false);
  const [showEditPostModal, setShowEditPostModal] = useState(false);
  const closeModal = useReportModal(approveCallback);
  const toast = useToast();
  const dCaptcha = useCaptchaModal('delete');

  const handleEditPostModalState = () => {
    setShowEditPostModal(!showEditPostModal);
  };

  const handleDelete = () => {
    let content = '';

    setIsLoading(true);

    if (objectType === 'question') {
      content =
        Number(itemData?.answer_count) > 0
          ? t('question', { keyPrefix: 'delete' })
          : t('other', { keyPrefix: 'delete' });
    }
    if (objectType === 'answer') {
      content = itemData?.answer_accepted
        ? t('answer_accepted', { keyPrefix: 'delete' })
        : t('other', { keyPrefix: 'delete' });
    }
    if (objectType === 'comment') {
      content = t('other', { keyPrefix: 'delete' });
    }
    Modal.confirm({
      title: t('title', { keyPrefix: 'delete' }),
      content,
      cancelBtnVariant: 'link',
      confirmBtnVariant: 'danger',
      confirmText: t('delete', { keyPrefix: 'btns' }),
      onConfirm: () => {
        dCaptcha.check(() => {
          const req: Type.PutFlagReviewParams = {
            operation_type: 'delete_post',
            flag_id: String(itemData?.flag_id),
            captcha_code: undefined,
            captcha_id: undefined,
          };
          dCaptcha.resolveCaptchaReq(req);

          delete req.captcha_code;
          delete req.captcha_id;

          putFlagReviewAction(req)
            .then(async () => {
              await dCaptcha.close();
              let msg = '';
              if (objectType === 'question') {
                msg = t('post_deleted', { keyPrefix: 'messages' });
              }
              if (objectType === 'answer') {
                msg = t('tip_answer_deleted');
              }
              if (objectType === 'answer' || objectType === 'question') {
                toast.onShow({
                  msg,
                  variant: 'success',
                });
              }
              approveCallback();
            })
            .catch((ex) => {
              if (ex.isError) {
                dCaptcha.handleCaptchaError(ex.list);
              }
            })
            .finally(() => {
              setIsLoading(false);
            });
        });
      },
      onCancel: () => {
        setIsLoading(false);
      },
    });
  };

  const handleAction = (type) => {
    if (type === 'delete') {
      handleDelete();
    }

    if (type === 'close') {
      closeModal.onShow({
        type: 'question',
        id: itemData?.flag_id || '',
        action: 'close',
        source: 'review',
        content: itemData?.reason_content,
        reportType: itemData?.reason.reason_type || -1,
      });
    }

    if (type === 'unlist') {
      const keyPrefix = 'question_detail.unlist';
      Modal.confirm({
        title: t('title', { keyPrefix }),
        content: t('content', { keyPrefix }),
        cancelBtnVariant: 'link',
        confirmText: t('confirm_btn', { keyPrefix }),
        onConfirm: () => {
          putFlagReviewAction({
            operation_type: 'unlist_post',
            flag_id: itemData?.flag_id || '',
          }).then(() => {
            toast.onShow({
              msg: t(`post_${type}`, { keyPrefix: 'messages' }),
              variant: 'success',
            });
            approveCallback();
          });
        },
      });
    }
  };

  const handleActionEdit = () => {
    handleEditPostModalState();
  };

  return (
    <div>
      <Dropdown>
        <Dropdown.Toggle
          as={Button}
          disabled={isLoading}
          variant="outline-primary"
          id="dropdown-basic">
          {t('approve', { keyPrefix: 'btns' })}
        </Dropdown.Toggle>

        <Dropdown.Menu>
          <Dropdown.Item onClick={() => handleActionEdit()}>
            {t('edit_post')}
          </Dropdown.Item>
          {curFilter === 'normal' && objectType === 'question' && (
            <Dropdown.Item onClick={() => handleAction('close')}>
              {t('close', { keyPrefix: 'btns' })}
            </Dropdown.Item>
          )}
          {curFilter !== 'deleted' && (
            <Dropdown.Item onClick={() => handleAction('delete')}>
              {t('delete', { keyPrefix: 'btns' })}
            </Dropdown.Item>
          )}
          {objectType === 'question' && (
            <>
              <Dropdown.Divider />
              {itemData?.object_show_status !== 2 && (
                <Dropdown.Item onClick={() => handleAction('unlist')}>
                  {t('unlist_post')}
                </Dropdown.Item>
              )}
            </>
          )}
        </Dropdown.Menu>
      </Dropdown>
      <EditPostModal
        visible={showEditPostModal}
        handleClose={handleEditPostModalState}
        objectType={objectType}
        originalData={{
          flag_id: itemData?.flag_id || '',
          id: itemData?.object_id || '',
          title: itemData?.title || '',
          content: itemData?.original_text || '',
          tags: itemData?.tags || [],
          question_id: itemData?.question_id || '',
          answer_id: itemData?.answer_id || '',
        }}
        callback={approveCallback}
      />
    </div>
  );
};

export default Index;

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

import { memo, FC } from 'react';
import { Button, Dropdown } from 'react-bootstrap';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { Icon, Modal } from '@/components';
import { useReportModal, useToast } from '@/hooks';
import { useCaptchaPlugin } from '@/utils/pluginKit';
import { QuestionOperationReq } from '@/common/interface';
import Share from '../Share';
import {
  deleteQuestion,
  deleteAnswer,
  editCheck,
  reopenQuestion,
  questionOperation,
  unDeleteAnswer,
  unDeleteQuestion,
} from '@/services';
import { tryNormalLogged } from '@/utils/guard';
import { floppyNavigation } from '@/utils';
import { toastStore } from '@/stores';

import '@/components/QueryGroup/index.scss';

interface IProps {
  type: 'answer' | 'question';
  qid: string;
  aid?: string;
  title: string;
  hasAnswer?: boolean;
  isAccepted: boolean;
  callback: (type: string) => void;
  memberActions;
}
const Index: FC<IProps> = ({
  type,
  qid,
  aid = '',
  title,
  isAccepted = false,
  hasAnswer = false,
  memberActions = [],
  callback,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'delete' });
  const toast = useToast();
  const navigate = useNavigate();
  const reportModal = useReportModal();
  const dCaptcha = useCaptchaPlugin('delete');

  const refreshQuestion = () => {
    callback?.('default');
  };
  const closeModal = useReportModal(refreshQuestion);
  const editUrl =
    type === 'answer' ? `/posts/${qid}/${aid}/edit` : `/posts/${qid}/edit`;

  const handleReport = () => {
    reportModal.onShow({
      type,
      id: type === 'answer' ? aid : qid,
      action: 'flag',
    });
  };

  const handleClose = () => {
    closeModal.onShow({
      type,
      id: qid,
      action: 'close',
    });
  };

  const submitDeleteQuestion = () => {
    const req = {
      id: qid,
      captcha_code: undefined,
      captcha_id: undefined,
    };
    dCaptcha?.resolveCaptchaReq(req);

    deleteQuestion(req)
      .then(async () => {
        await dCaptcha?.close();
        toast.onShow({
          msg: t('post_deleted', { keyPrefix: 'messages' }),
          variant: 'success',
        });
        callback?.('delete_question');
      })
      .catch((ex) => {
        if (ex.isError) {
          dCaptcha?.handleCaptchaError(ex.list);
        }
      });
  };

  const submitDeleteAnswer = () => {
    const req = {
      id: aid,
      captcha_code: undefined,
      captcha_id: undefined,
    };
    dCaptcha?.resolveCaptchaReq(req);

    deleteAnswer(req)
      .then(async () => {
        await dCaptcha?.close();
        // refresh page
        toast.onShow({
          msg: t('tip_answer_deleted'),
          variant: 'success',
        });
        callback?.('delete_answer');
      })
      .catch((ex) => {
        if (ex.isError) {
          dCaptcha?.handleCaptchaError(ex.list);
        }
      });
  };

  const handleDelete = () => {
    if (type === 'question') {
      Modal.confirm({
        title: t('title'),
        content: hasAnswer ? t('question') : t('other'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          if (!dCaptcha) {
            submitDeleteQuestion();
            return;
          }
          dCaptcha.check(() => {
            submitDeleteQuestion();
          });
        },
      });
    }

    if (type === 'answer' && aid) {
      Modal.confirm({
        title: t('title'),
        content: isAccepted ? t('answer_accepted') : t('other'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('delete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          if (!dCaptcha) {
            submitDeleteAnswer();
            return;
          }
          dCaptcha.check(() => {
            submitDeleteAnswer();
          });
        },
      });
    }
  };

  const handleUndelete = () => {
    Modal.confirm({
      title: t('undelete_title'),
      content: t('undelete_desc'),
      cancelBtnVariant: 'link',
      confirmBtnVariant: 'danger',
      confirmText: t('undelete', { keyPrefix: 'btns' }),
      onConfirm: () => {
        if (type === 'question') {
          unDeleteQuestion(qid).then(() => {
            callback?.('default');
          });
        }

        if (type === 'answer') {
          unDeleteAnswer(aid).then(() => {
            callback?.('all');
          });
        }
      },
    });
  };

  const handleEdit = (evt, targetUrl) => {
    if (!floppyNavigation.shouldProcessLinkClick(evt)) {
      return;
    }
    evt.preventDefault();
    let checkObjectId = qid;
    if (type === 'answer') {
      checkObjectId = aid;
    }
    editCheck(checkObjectId).then(() => {
      navigate(targetUrl);
    });
  };

  const handleReopen = () => {
    Modal.confirm({
      title: t('title', { keyPrefix: 'question_detail.reopen' }),
      content: t('content', { keyPrefix: 'question_detail.reopen' }),
      cancelBtnVariant: 'link',
      confirmText: t('confirm_btn', { keyPrefix: 'question_detail.reopen' }),
      onConfirm: () => {
        reopenQuestion({
          question_id: qid,
        }).then(() => {
          toast.onShow({
            msg: t('post_reopen', { keyPrefix: 'messages' }),
            variant: 'success',
          });
          refreshQuestion();
        });
      },
    });
  };

  const handleCommon = async (params) => {
    await questionOperation(params);
    let msg = '';
    if (params.operation === 'pin') {
      msg = t('post_pin', { keyPrefix: 'messages' });
    }
    if (params.operation === 'unpin') {
      msg = t('post_unpin', { keyPrefix: 'messages' });
    }
    if (params.operation === 'hide') {
      msg = t('post_hide_list', { keyPrefix: 'messages' });
    }
    if (params.operation === 'show') {
      msg = t('post_show_list', { keyPrefix: 'messages' });
    }
    toastStore.getState().show({
      msg,
      variant: 'success',
    });
    setTimeout(() => {
      refreshQuestion();
    }, 100);
  };

  const handlOtherActions = (action) => {
    const params: QuestionOperationReq = {
      id: qid,
      operation: action,
    };

    if (action === 'pin') {
      Modal.confirm({
        title: t('title', { keyPrefix: 'question_detail.pin' }),
        content: t('content', { keyPrefix: 'question_detail.pin' }),
        cancelBtnVariant: 'link',
        confirmText: t('confirm_btn', { keyPrefix: 'question_detail.pin' }),
        onConfirm: () => {
          handleCommon(params);
        },
      });
    } else {
      handleCommon(params);
    }
  };

  const handleAction = (action) => {
    if (!tryNormalLogged(true)) {
      return;
    }
    if (action === 'delete') {
      handleDelete();
    }

    if (action === 'undelete') {
      handleUndelete();
    }

    if (action === 'report') {
      handleReport();
    }

    if (action === 'close') {
      handleClose();
    }

    if (action === 'reopen') {
      handleReopen();
    }

    if (
      action === 'pin' ||
      action === 'unpin' ||
      action === 'hide' ||
      action === 'show'
    ) {
      handlOtherActions(action);
    }
  };

  const firstAction =
    memberActions?.filter(
      (v) =>
        v.action === 'report' ||
        v.action === 'edit' ||
        v.action === 'delete' ||
        v.action === 'undelete',
    ) || [];
  const secondAction =
    memberActions?.filter(
      (v) =>
        v.action === 'close' ||
        v.action === 'reopen' ||
        v.action === 'pin' ||
        v.action === 'unpin' ||
        v.action === 'hide' ||
        v.action === 'show',
    ) || [];

  return (
    <>
      <div className="md-show align-items-center">
        <Share
          type={type}
          qid={qid}
          aid={aid}
          title={title}
          className="link-secondary small"
        />
        {firstAction?.map((item) => {
          if (item.action === 'edit') {
            return (
              <Link
                key={item.action}
                to={editUrl}
                className="link-secondary p-0 small ms-3"
                onClick={(evt) => handleEdit(evt, editUrl)}
                style={{ lineHeight: '23px' }}>
                {item.name}
              </Link>
            );
          }
          return (
            <Button
              key={item.action}
              variant="link"
              size="sm"
              className="link-secondary p-0 ms-3"
              onClick={() => handleAction(item.action)}>
              {item.name}
            </Button>
          );
        })}
        {secondAction.length > 0 && (
          <Dropdown className="ms-3 d-flex">
            <Dropdown.Toggle
              variant="link"
              size="sm"
              title={t('action', { keyPrefix: 'question_detail' })}
              className="link-secondary p-0 no-toggle">
              <Icon name="three-dots" />
            </Dropdown.Toggle>
            <Dropdown.Menu>
              {secondAction.map((item) => {
                return (
                  <Dropdown.Item
                    key={item.action}
                    onClick={() => handleAction(item.action)}>
                    {item.name}
                  </Dropdown.Item>
                );
              })}
            </Dropdown.Menu>
          </Dropdown>
        )}
      </div>
      <div className="md-hide">
        {memberActions.length > 0 && (
          <Dropdown className="d-flex">
            <Dropdown.Toggle
              variant="link"
              size="sm"
              title={t('action', { keyPrefix: 'question_detail' })}
              className="link-secondary no-toggle">
              <Icon name="three-dots" />
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Share
                type={type}
                qid={qid}
                aid={aid}
                title={title}
                className="inherit"
                mode="mobile"
              />
              {[...firstAction, ...secondAction].map((item) => {
                return (
                  <Dropdown.Item
                    key={item.action}
                    onClick={() => handleAction(item.action)}>
                    {item.name}
                  </Dropdown.Item>
                );
              })}
            </Dropdown.Menu>
          </Dropdown>
        )}
      </div>
    </>
  );
};

export default memo(Index);

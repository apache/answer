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

import { FC, useEffect, useState, useRef } from 'react';
import { Card, Alert, Stack, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link, useSearchParams } from 'react-router-dom';

import classNames from 'classnames';

import { getPendingReviewPostList, putPendingReviewAction } from '@/services';
import {
  BaseUserCard,
  Tag,
  FormatTime,
  Icon,
  ImgViewer,
  htmlRender,
} from '@/components';
import { scrollToDocTop } from '@/utils';
import type * as Type from '@/common/interface';
import { ADMIN_LIST_STATUS } from '@/common/constants';
import generateData from '../../utils/generateData';

interface IProps {
  refreshCount: () => void;
}

const Index: FC<IProps> = ({ refreshCount }) => {
  const [urlSearch, setUrlSearchParams] = useSearchParams();
  const objectId = urlSearch.get('objectId') || '';
  const { t } = useTranslation('translation', { keyPrefix: 'page_review' });
  const ref = useRef<HTMLDivElement>(null);
  const [noTasks, setNoTasks] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [reviewResp, setReviewResp] = useState<Type.QuestionDetailRes>();
  const flagItemData = reviewResp?.list[0] as Type.QueuedReviewItem;

  const resolveNextOne = (resp, pageNumber) => {
    const { count, list = [] } = resp;
    // auto rollback
    if (!list.length && count && page !== 1) {
      pageNumber = 1;
      setPage(pageNumber);
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      queryNextOne(pageNumber, '');
      return;
    }
    if (pageNumber !== page) {
      setPage(pageNumber);
    }
    setReviewResp(resp);
    if (!list.length) {
      setNoTasks(true);
    }
    setTimeout(() => {
      scrollToDocTop();
    }, 150);
  };

  const queryNextOne = (pageNumber, id) => {
    getPendingReviewPostList(pageNumber, id).then((resp) => {
      resolveNextOne(resp, pageNumber);
    });
  };

  useEffect(() => {
    if (!ref.current) {
      return;
    }

    setTimeout(() => {
      htmlRender(ref.current, {
        copySuccessText: t('copied', { keyPrefix: 'messages' }),
        copyText: t('copy', { keyPrefix: 'messages' }),
      });
    }, 70);
  }, [ref.current]);

  useEffect(() => {
    queryNextOne(page, objectId);
  }, []);

  const handleAction = (type: 'approve' | 'reject') => {
    if (!flagItemData) {
      return;
    }
    setIsLoading(true);
    putPendingReviewAction({
      status: type,
      review_id: flagItemData?.review_id,
    })
      .then(() => {
        refreshCount();
        queryNextOne(page, '');
        if (objectId) {
          urlSearch.delete('objectId');
          setUrlSearchParams(urlSearch);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  const handlingSkip = () => {
    queryNextOne(page + 1, '');
    if (objectId) {
      urlSearch.delete('objectId');
      setUrlSearchParams(urlSearch);
    }
  };

  const { object_type, author_user_info, object_status, reason } =
    flagItemData || {
      object_type: '',
      author_user_info: null,
      reason: null,
      object_status: 0,
    };

  const { itemLink, itemId, itemTimePrefix } = generateData(flagItemData);

  if (noTasks) return null;
  return (
    <Card>
      <Card.Header>
        {object_type !== 'user' ? t('queued_post') : t('queued_post_user')}
      </Card.Header>
      <Card.Body className="p-0">
        <Alert variant="info" className="border-0 rounded-0 mb-0">
          <Stack
            direction="horizontal"
            gap={1}
            className="align-items-center mb-2">
            <div className="small d-flex align-items-center">
              <Icon type="bi" name="plugin" size="24px" className="me-2 lh-1" />
              <span>{flagItemData?.submitter_display_name}</span>
            </div>
            {flagItemData?.submit_at && (
              <FormatTime
                time={flagItemData.submit_at}
                className="small text-secondary"
                preFix={t('proposed')}
              />
            )}
          </Stack>
          <Stack className="align-items-start">
            <p className="mb-0">{reason}</p>
          </Stack>
        </Alert>
        <div className="p-3">
          <small className="d-block text-secondary mb-4">
            <span>{t(object_type, { keyPrefix: 'btns' })} </span>
            <Link to={itemLink} target="_blank" className="link-secondary">
              #{itemId}
            </Link>
          </small>
          {object_type === 'question' && (
            <>
              <h5 className="mb-3">{flagItemData?.title}</h5>
              <div className="mb-4">
                {flagItemData?.tags?.map((item) => {
                  return (
                    <Tag key={item.slug_name} className="me-1" data={item} />
                  );
                })}
              </div>
            </>
          )}
          <div className="small font-monospace">
            <ImgViewer>
              <article
                ref={ref}
                className="fmt text-break text-wrap"
                dangerouslySetInnerHTML={{ __html: flagItemData?.parsed_text }}
              />
            </ImgViewer>
          </div>
          <div className="d-flex flex-wrap align-items-center justify-content-between mt-4">
            <div>
              <span
                className={classNames(
                  'badge',
                  ADMIN_LIST_STATUS[object_status]?.variant,
                )}>
                {t(ADMIN_LIST_STATUS[object_status]?.name, {
                  keyPrefix: 'btns',
                })}
              </span>
              {flagItemData?.object_show_status === 2 && (
                <span
                  className={classNames(
                    'ms-1 badge',
                    ADMIN_LIST_STATUS.unlisted.variant,
                  )}>
                  {t(ADMIN_LIST_STATUS.unlisted.name, { keyPrefix: 'btns' })}
                </span>
              )}
            </div>
            <div className="d-flex align-items-center small">
              <BaseUserCard
                data={author_user_info}
                avatarSize="24"
                avatarClass="me-2"
              />
              <FormatTime
                time={Number(flagItemData?.created_at)}
                className="text-secondary ms-1 flex-shrink-0"
                preFix={t(itemTimePrefix, { keyPrefix: 'question_detail' })}
              />
            </div>
          </div>
        </div>
      </Card.Body>

      <Card.Footer className="p-3">
        <p>
          {object_type !== 'user'
            ? t('approve_post_tip')
            : t('approve_user_tip')}
        </p>
        <Stack direction="horizontal" gap={2}>
          <Button
            variant="outline-primary"
            disabled={isLoading}
            onClick={() => handleAction('approve')}>
            {t('approve', { keyPrefix: 'btns' })}
          </Button>
          <Button
            variant="outline-primary"
            disabled={isLoading}
            onClick={() => handleAction('reject')}>
            {t('reject', { keyPrefix: 'btns' })}
          </Button>

          <Button
            variant="outline-primary"
            disabled={isLoading}
            onClick={handlingSkip}>
            {t('skip', { keyPrefix: 'btns' })}
          </Button>
        </Stack>
      </Card.Footer>
    </Card>
  );
};

export default Index;

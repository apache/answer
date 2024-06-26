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
import { Alert, Stack, Button, Card } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { BaseUserCard, FormatTime, DiffContent } from '@/components';
import { getSuggestReviewList, revisionAudit } from '@/services';
import { pathFactory } from '@/router/pathFactory';
import { scrollToDocTop } from '@/utils';
import type * as Type from '@/common/interface';

interface IProps {
  refreshCount: () => void;
}

const Index: FC<IProps> = ({ refreshCount }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_review' });
  const [isLoading, setIsLoading] = useState(false);
  const [noTasks, setNoTasks] = useState(false);
  const [page, setPage] = useState(1);
  const [reviewResp, setReviewResp] = useState<Type.SuggestReviewResp>();
  const ro = reviewResp?.list[0];
  const { info, type, unreviewed_info } = ro || {
    info: null,
    type: '',
    unreviewed_info: null,
  };
  const resolveNextOne = (resp, pageNumber) => {
    const { count, list = [] } = resp;
    // auto rollback
    if (!list.length && count && page !== 1) {
      pageNumber = 1;
      setPage(pageNumber);
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      queryNextOne(pageNumber);
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
  const queryNextOne = (pageNumber) => {
    getSuggestReviewList(pageNumber).then((resp) => {
      resolveNextOne(resp, pageNumber);
    });
  };
  const reviewInfo = unreviewed_info?.content;

  const handlingApprove = () => {
    if (!unreviewed_info) {
      return;
    }
    setIsLoading(true);
    revisionAudit(unreviewed_info.id, 'approve')
      .then(() => {
        refreshCount();
        queryNextOne(page);
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  const handlingReject = () => {
    if (!unreviewed_info) {
      return;
    }
    setIsLoading(true);
    revisionAudit(unreviewed_info.id, 'reject')
      .then(() => {
        refreshCount();
        queryNextOne(page);
      })
      .catch((ex) => {
        console.error('revisionAudit reject error: ', ex);
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  const handlingSkip = () => {
    queryNextOne(page + 1);
  };

  let itemLink = '';
  let itemId = '';
  let editSummary = unreviewed_info?.reason;
  const editor = unreviewed_info?.user_info;
  const editTime = unreviewed_info?.create_at;
  if (type === 'question') {
    itemLink = pathFactory.questionLanding(info?.object_id, info?.url_title);
    itemId = info?.object_id;
    editSummary ||= t('edit_question');
  } else if (type === 'answer') {
    itemLink = pathFactory.answerLanding({
      // @ts-ignore
      questionId: unreviewed_info.content.question_id,
      slugTitle: info?.url_title,
      answerId: unreviewed_info.object_id,
    });
    itemId = unreviewed_info.object_id;
    editSummary ||= t('edit_answer');
  } else if (type === 'tag') {
    const tagInfo = unreviewed_info.content as Type.Tag;
    itemLink = pathFactory.tagLanding(tagInfo.slug_name);
    itemId = tagInfo?.tag_id || tagInfo.slug_name;
    editSummary ||= t('edit_tag');
  }
  useEffect(() => {
    queryNextOne(page);
  }, []);

  if (noTasks) return null;

  let newData: Record<string, any> = {};
  let oldData: Record<string, any> = {};
  let diffOpts: Partial<{
    showTitle: boolean;
    showTagUrlSlug: boolean;
  }> = {
    showTitle: true,
    showTagUrlSlug: true,
  };
  if (type === 'question' && info && reviewInfo && 'content' in reviewInfo) {
    newData = {
      title: reviewInfo.title,
      original_text: reviewInfo.content,
      tags: reviewInfo.tags,
    };
    oldData = {
      title: info.title,
      original_text: info.content,
      tags: info.tags,
    };
  }
  if (type === 'answer' && info && reviewInfo && 'content' in reviewInfo) {
    newData = {
      original_text: reviewInfo.content,
    };
    oldData = {
      original_text: info.content,
    };
  }

  if (type === 'tag' && info && reviewInfo) {
    newData = {
      original_text: reviewInfo.original_text,
    };
    oldData = {
      original_text: info.content,
    };
    diffOpts = { showTitle: false, showTagUrlSlug: false };
  }

  return (
    <Card>
      <Card.Header>{t('suggest_edits')}</Card.Header>
      <Card.Body className="p-0">
        <Alert variant="info" className="border-0 rounded-0 mb-0">
          <Stack
            direction="horizontal"
            gap={1}
            className="align-items-center mb-2">
            <BaseUserCard data={editor} avatarSize="24" avatarClass="me-2" />
            {editTime && (
              <FormatTime
                time={editTime}
                className="small text-secondary"
                preFix={t('proposed')}
              />
            )}
          </Stack>
          <Stack className="align-items-start">
            <p className="mb-0">{editSummary}</p>
          </Stack>
        </Alert>
        <div className="p-3">
          <small className="d-block text-secondary mb-4">
            <span>{t(type, { keyPrefix: 'btns' })} </span>
            <Link to={itemLink} target="_blank" className="link-secondary">
              #{itemId}
            </Link>
          </small>

          <DiffContent
            className="mt-2"
            objectType={type}
            newData={newData}
            oldData={oldData}
            opts={diffOpts}
          />
        </div>
      </Card.Body>
      <Card.Footer className="p-3">
        <p>{t('approve_revision_tip')}</p>
        <Stack direction="horizontal" gap={2}>
          <Button
            variant="outline-primary"
            disabled={isLoading}
            onClick={handlingApprove}>
            {t('approve', { keyPrefix: 'btns' })}
          </Button>
          <Button
            variant="outline-primary"
            disabled={isLoading}
            onClick={handlingReject}>
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

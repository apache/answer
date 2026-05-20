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

import { memo, FC, useState, useEffect, useRef } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button, OverlayTrigger, Tooltip } from 'react-bootstrap';

import {
  Tag,
  Actions,
  Operate,
  BaseUserCard,
  Comment,
  FormatTime,
  htmlRender,
  ImgViewer,
} from '@/components';
import { useRenderHtmlPlugin } from '@/utils/pluginKit';
import { formatCount, guard } from '@/utils';
import { following } from '@/services';
import { pathFactory } from '@/router/pathFactory';

interface Props {
  data: any;
  hasAnswer: boolean;
  initPage: (type: string) => void;
  isLogged: boolean;
}

const Index: FC<Props> = ({ data, initPage, hasAnswer, isLogged }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'question_detail',
  });
  const [searchParams] = useSearchParams();
  const [followed, setFollowed] = useState(data?.is_followed);
  const ref = useRef<HTMLDivElement>(null);

  useRenderHtmlPlugin(ref.current);

  const handleFollow = (e) => {
    e.preventDefault();
    if (!guard.tryNormalLogged(true)) {
      return;
    }
    following({
      object_id: data?.id,
      is_cancel: followed,
    }).then((res) => {
      setFollowed(res.is_followed);
    });
  };

  useEffect(() => {
    if (data) {
      setFollowed(data?.is_followed);
    }
  }, [data]);

  useEffect(() => {
    if (!ref.current) {
      return;
    }

    htmlRender(ref.current, {
      copySuccessText: t('copied', { keyPrefix: 'messages' }),
      copyText: t('copy', { keyPrefix: 'messages' }),
    });
  }, [ref.current]);

  if (!data?.id) {
    return null;
  }

  return (
    <div>
      <h1 className="h3 mb-2 text-wrap text-break pb-1">
        <Link
          className="link-dark"
          reloadDocument
          to={pathFactory.questionLanding(data.id, data.url_title)}>
          {data.title}
          {data.status === 2
            ? ` [${t('closed', { keyPrefix: 'question' })}]`
            : ''}
        </Link>
      </h1>

      <div className="d-flex flex-wrap align-items-center small mb-4 text-secondary border-bottom pb-3">
        <BaseUserCard data={data.user_info} className="me-3" />

        {isLogged ? (
          <>
            <Link to={`/posts/${data.id}/timeline`}>
              <FormatTime
                time={data.create_time}
                preFix={t('created')}
                className="me-3 link-secondary"
              />
            </Link>

            <Link to={`/posts/${data.id}/timeline`}>
              <FormatTime
                time={data.edit_time}
                preFix={t('Edited')}
                className="me-3 link-secondary"
              />
            </Link>
          </>
        ) : (
          <>
            <FormatTime
              time={data.create_time}
              preFix={t('created')}
              className="me-3 link-secondary"
            />

            <FormatTime
              time={data.edit_time}
              preFix={t('Edited')}
              className="me-3 link-secondary"
            />
          </>
        )}

        {data?.view_count > 0 && (
          <div className="me-3">
            {t('Views')} {formatCount(data.view_count)}
          </div>
        )}
        <OverlayTrigger
          placement="bottom"
          overlay={<Tooltip id="followTooltip">{t('follow_tip')}</Tooltip>}>
          <Button
            variant="link"
            size="sm"
            className="p-0 btn-no-border"
            onClick={(e) => handleFollow(e)}>
            {t(followed ? 'Following' : 'Follow')}
          </Button>
        </OverlayTrigger>
      </div>

      <ImgViewer>
        <article
          ref={ref}
          className="fmt text-break text-wrap last-p mb-4"
          dangerouslySetInnerHTML={{ __html: data?.html }}
        />
      </ImgViewer>

      <div className="m-n1">
        {data?.tags?.map((item: any) => {
          return <Tag className="m-1" key={item.slug_name} data={item} />;
        })}
      </div>

      <Actions
        className="mt-4"
        source="question"
        data={{
          id: data?.id,
          isHate: data?.vote_status === 'vote_down',
          isLike: data?.vote_status === 'vote_up',
          votesCount: data?.vote_count,
          collected: data?.collected,
          collectCount: data?.collection_count,
          username: data.user_info?.username,
        }}
      />

      <div className="mt-4">
        <Comment
          objectId={data?.id}
          mode="question"
          commentId={searchParams.get('commentId')}>
          <Operate
            qid={data?.id}
            type="question"
            memberActions={data?.member_actions}
            title={data.title}
            hasAnswer={hasAnswer}
            isAccepted={Boolean(data?.accepted_answer_id)}
            callback={initPage}
          />
        </Comment>
      </div>
    </div>
  );
};

export default memo(Index);

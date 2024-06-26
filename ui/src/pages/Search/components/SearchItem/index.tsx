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
import { Link, useSearchParams } from 'react-router-dom';
import { ListGroupItem } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { pathFactory } from '@/router/pathFactory';
import {
  Tag,
  FormatTime,
  BaseUserCard,
  Counts,
  HighlightText,
} from '@/components';
import Pattern from '@/common/pattern';
import type { SearchResItem } from '@/common/interface';
import { escapeRemove } from '@/utils';

interface Props {
  data: SearchResItem;
}
const Index: FC<Props> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'question' });
  if (!data?.object_type) {
    return null;
  }
  let itemUrl = pathFactory.questionLanding(
    data.object.id,
    data.object.url_title,
  );
  if (data.object_type === 'answer' && data.object.question_id) {
    itemUrl = pathFactory.answerLanding({
      questionId: data.object.question_id,
      slugTitle: data.object.url_title,
      answerId: data.object.id,
    });
  }

  const [searchParams] = useSearchParams();
  const q = searchParams.get('q');
  const keywords =
    q
      ?.replace(Pattern.search, '')
      ?.split(' ')
      ?.filter((v) => v !== '') || [];

  return (
    <ListGroupItem className="py-3 px-0 border-start-0 border-end-0 bg-transparent">
      <div className="mb-2 clearfix">
        <span
          className="float-start me-2 badge text-bg-dark"
          style={{ marginTop: '2px' }}>
          {data.object_type === 'question' ? 'Q' : 'A'}
        </span>
        <Link className="h5 mb-0 link-dark text-break" to={itemUrl}>
          <HighlightText text={data.object.title} keywords={keywords} />
          {data.object.status === 'closed'
            ? ` [${t('closed', { keyPrefix: 'question' })}]`
            : null}
        </Link>
      </div>
      <div className="d-flex flex-wrap align-items-center small text-secondary mb-2">
        <BaseUserCard data={data.object?.user_info} showAvatar={false} />

        <span className="split-dot" />
        <FormatTime
          time={data.object?.created_at}
          className="me-3"
          preFix={data.object_type === 'question' ? 'asked' : 'answered'}
        />

        <Counts
          className="my-2 my-sm-0"
          showViews={false}
          isAccepted={data.object?.accepted}
          showAnswers={data.object_type === 'question'}
          showAccepted={data.object?.accepted && data.object_type === 'answer'}
          data={{
            votes: data.object?.vote_count,
            answers: data.object?.answer_count,
            views: 0,
          }}
        />
      </div>

      {data.object?.excerpt && (
        <p className="small text-truncate-2 mb-2 last-p text-break">
          <HighlightText
            text={escapeRemove(data.object.excerpt) || ''}
            keywords={keywords}
          />
        </p>
      )}

      {data.object?.tags?.map((item) => {
        return <Tag key={item.slug_name} className="me-1" data={item} />;
      })}
    </ListGroupItem>
  );
};

export default memo(Index);

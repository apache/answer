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

import { FC, memo } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { pathFactory } from '@/router/pathFactory';
import { Icon } from '@/components';

interface Props {
  data: any[];
  type: 'comment' | 'post';
}
const Index: FC<Props> = ({ data, type }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });
  const visibleData = data.filter((item) =>
    type === 'comment'
      ? item.answer_id && item.question_id && item.question_info?.title
      : item.question_id && item.title,
  );
  return (
    <ol className="list-unstyled">
      {visibleData.map((item, index) => {
        return (
          <li
            className={`${index === visibleData.length - 1 ? '' : 'mb-2'}`}
            key={type === 'comment' ? item.answer_id : item.question_id}>
            <Link
              className="text-truncate-1"
              to={
                type === 'comment'
                  ? pathFactory.answerLanding({
                      questionId: item.question_id,
                      slugTitle: item.question_info?.url_title,
                      answerId: item.answer_id,
                    })
                  : pathFactory.questionLanding(
                      item.question_id,
                      item.url_title,
                    )
              }>
              {type === 'comment' ? item.question_info.title : item.title}
            </Link>

            <div className="text-secondary small">
              <Icon name="hand-thumbs-up-fill me-1" />
              <span>
                {item.vote_count} {t('votes', { keyPrefix: 'counts' })}
              </span>

              {type === 'post' && (
                <div className="d-inline-block text-secondary ms-3 small">
                  <Icon name="chat-square-text-fill" />

                  <span>
                    {' '}
                    {item.answer_count} {t('comments')}
                  </span>
                </div>
              )}
            </div>
          </li>
        );
      })}
    </ol>
  );
};

export default memo(Index);

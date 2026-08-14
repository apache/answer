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
import { ListGroup, ListGroupItem } from 'react-bootstrap';
import { Link } from 'react-router-dom';

import { pathFactory } from '@/router/pathFactory';
import { FormatTime } from '@/components';

interface Props {
  visible: boolean;
  data;
}

const Index: FC<Props> = ({ visible, data }) => {
  if (!visible || !data?.length) {
    return null;
  }
  const visibleData = data.filter(
    (item) =>
      item.title &&
      item.question_id &&
      (item.object_type === 'question' ||
        (item.object_type === 'answer' && item.answer_id)),
  );
  return (
    <ListGroup className="rounded-0">
      {visibleData.map((item) => {
        return (
          <ListGroupItem
            className="py-3 px-0 bg-transparent border-start-0 border-end-0"
            key={item.comment_id}>
            <Link
              className="text-break"
              to={
                item.object_type === 'question'
                  ? pathFactory.questionLanding(
                      item.question_id,
                      item.url_title,
                    )
                  : pathFactory.answerLanding({
                      questionId: item.question_id,
                      slugTitle: item.url_title,
                      answerId: item.answer_id,
                    })
              }>
              {item.title}
            </Link>
            <div
              className="small mb-2 last-p text-break text-truncate-2"
              dangerouslySetInnerHTML={{
                __html: item.content,
              }}
            />

            <FormatTime
              time={item.created_at}
              className="small text-secondary"
            />
          </ListGroupItem>
        );
      })}
    </ListGroup>
  );
};

export default memo(Index);

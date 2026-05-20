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

import { FormatTime, Tag, Counts } from '@/components';
import { pathFactory } from '@/router/pathFactory';

interface Props {
  visible: boolean;
  data: any[];
}
const Index: FC<Props> = ({ visible, data }) => {
  if (!visible || !data?.length) {
    return null;
  }
  return (
    <ListGroup className="rounded-0">
      {data.map((item) => {
        return (
          <ListGroupItem
            className="py-3 px-0 bg-transparent border-start-0 border-end-0"
            key={item.answer_id}>
            <h6 className="mb-2">
              <Link
                to={pathFactory.answerLanding({
                  questionId: item.question_id,
                  slugTitle: item.question_info?.url_title,
                  answerId: item.answer_id,
                })}
                className="text-break">
                {item.question_info?.title}
              </Link>
            </h6>
            <div className="d-flex align-items-center small text-secondary mb-2">
              <FormatTime time={item.create_time} className="me-3" />

              <Counts
                data={{ votes: item?.vote_count, views: 0, answers: 0 }}
                showAnswers={false}
                showViews={false}
                showAccepted={item.accepted === 2}
              />
            </div>
            <div>
              {item.question_info?.tags?.map((tag) => {
                return <Tag key={tag.slug_name} className="me-1" data={tag} />;
              })}
            </div>
          </ListGroupItem>
        );
      })}
    </ListGroup>
  );
};

export default memo(Index);

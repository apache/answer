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

import { FC } from 'react';
import { Card, ListGroup, ListGroupItem } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { pathFactory } from '@/router/pathFactory';
import { Icon } from '@/components';
import { useHotQuestions } from '@/services';

const HotQuestions: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'question' });
  const { data: questionRes } = useHotQuestions();
  if (!questionRes?.list?.length) {
    return null;
  }
  return (
    <Card>
      <Card.Header className="text-nowrap text-capitalize">
        {t('hot_questions')}
      </Card.Header>
      <ListGroup variant="flush">
        {questionRes?.list?.map((li) => {
          return (
            <ListGroupItem
              key={li.id}
              as={Link}
              to={pathFactory.questionLanding(li.id, li.url_title)}
              action>
              <div className="link-dark text-truncate-3">{li.title}</div>
              {li.answer_count > 0 ? (
                <div
                  className={`d-flex align-items-center small mt-1 ${
                    li.accepted_answer_id > 0
                      ? 'link-success'
                      : 'link-secondary'
                  }`}>
                  {li.accepted_answer_id >= 1 ? (
                    <Icon name="check-circle-fill" />
                  ) : (
                    <Icon name="chat-square-text-fill" />
                  )}
                  <span className="ms-1">
                    {t('x_answers', { count: li.answer_count })}
                  </span>
                </div>
              ) : null}
            </ListGroupItem>
          );
        })}
      </ListGroup>
    </Card>
  );
};
export default HotQuestions;

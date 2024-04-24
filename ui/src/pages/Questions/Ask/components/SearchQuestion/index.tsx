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

import { memo } from 'react';
import { Accordion, ListGroup } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Icon } from '@/components';
import { pathFactory } from '@/router/pathFactory';

import './index.scss';

const SearchQuestion = ({ similarQuestions }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ask' });
  // set max similar number
  if (similarQuestions && similarQuestions.length > 5) {
    similarQuestions.length = 5;
  }
  return (
    <Accordion defaultActiveKey="0" className="search-question-wrap mt-3">
      <Accordion.Item eventKey="0" className="overflow-hidden">
        <Accordion.Button className="px-3 py-2 bg-light text-body">
          {t('similar_questions')}
        </Accordion.Button>

        <Accordion.Body className="p-0">
          <ListGroup variant="flush">
            {similarQuestions.map((item) => {
              return (
                <ListGroup.Item
                  action
                  as="a"
                  className="link-dark text-wrap text-break"
                  key={item.id}
                  href={pathFactory.questionLanding(item.id, item.url_title)}
                  target="_blank">
                  <span
                    className={`${
                      item.accepted_answer || item.answer_count > 0
                        ? 'me-3'
                        : ''
                    }`}>
                    {item.title}
                    {item.status === 'closed'
                      ? ` [${t('closed', { keyPrefix: 'question' })}] `
                      : null}
                  </span>

                  {item.accepted_answer ? (
                    <span className="small text-success d-inline-block">
                      <Icon type="bi" name="check-circle-fill" />
                      <span className="ms-1">
                        {t('x_answers', {
                          keyPrefix: 'question',
                          count: item.answer_count,
                        })}
                      </span>
                    </span>
                  ) : (
                    item.answer_count > 0 && (
                      <span className="small text-secondary d-inline-block">
                        <Icon type="bi" name="chat-square-text-fill" />
                        <span className="ms-1">
                          {t('x_answers', {
                            keyPrefix: 'question',
                            count: item.answer_count,
                          })}
                        </span>
                      </span>
                    )
                  )}
                </ListGroup.Item>
              );
            })}
          </ListGroup>
        </Accordion.Body>
      </Accordion.Item>
    </Accordion>
  );
};

export default memo(SearchQuestion);

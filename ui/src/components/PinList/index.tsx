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
import { ListGroup, Stack } from 'react-bootstrap';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { Counts } from '@/components';
import { pathFactory } from '@/router/pathFactory';

interface IProps {
  data: any[];
}

const PinList: FC<IProps> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'question' });
  if (!data?.length) return null;

  return (
    <ListGroup.Item className="py-3 px-0 border-start-0 border-end-0">
      <Stack
        direction="horizontal"
        gap={3}
        className="overflow-x-auto align-items-stretch">
        {data.map((item) => {
          return (
            <ListGroup.Item
              action
              as="li"
              key={item.id}
              className="border-0 p-0"
              style={{
                minWidth: '238px',
                width: `${100 / data.length}%`,
              }}>
              <NavLink
                to={pathFactory.questionLanding(item.id, item.url_title)}
                className="border rounded h-100 d-flex flex-column justify-content-between p-3">
                <h6 className="text-wrap link-dark text-break text-truncate-2">
                  {item.title}
                  {item.status === 2 ? ` [${t('closed')}]` : ''}
                </h6>

                <Counts
                  data={{
                    votes: item.vote_count,
                    answers: item.answer_count,
                    views: item.view_count,
                  }}
                  isAccepted={item.accepted_answer_id >= 1}
                  showViews={false}
                  className="mt-2 mt-md-0 small text-secondary"
                />
              </NavLink>
            </ListGroup.Item>
          );
        })}
      </Stack>
    </ListGroup.Item>
  );
};

export default PinList;

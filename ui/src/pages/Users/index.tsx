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

import { Row, Col } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Fragment } from 'react';

import { usePageTags } from '@/hooks';
import { useQueryContributeUsers } from '@/services';
import { Avatar } from '@/components';

const Users = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'users' });

  const { data: users } = useQueryContributeUsers();

  usePageTags({
    title: t('users', { keyPrefix: 'page_title' }),
  });

  if (!users) {
    return null;
  }

  const keys = Object.keys(users);
  return (
    <Row className="py-4 mb-4 d-flex justify-content-center">
      <Col xxl={12}>
        <h1 className="fs-3 mb-4">{t('title')}</h1>
      </Col>

      <Col xxl={12}>
        {keys.map((key, index) => {
          if (users[key]?.length === 0) {
            return null;
          }
          return (
            <Fragment key={key}>
              <Row className="mb-4">
                <Col>
                  <h2 className="fs-6 mb-0">{t(key)}</h2>
                </Col>
              </Row>
              <Row className={index === keys.length - 1 ? '' : 'mb-4'}>
                {users[key]?.map((user) => (
                  <Col
                    key={user.username}
                    xl={3}
                    lg={4}
                    md={4}
                    sm={6}
                    xs={12}
                    className="mb-4">
                    <div className="d-flex">
                      <Link to={`/users/${user.username}`}>
                        <Avatar
                          size="48px"
                          avatar={user?.avatar}
                          searchStr="s=96"
                          alt={user.display_name}
                        />
                      </Link>
                      <div className="ms-2">
                        <Link
                          className="text-break"
                          to={`/users/${user.username}`}>
                          {user.display_name}
                        </Link>
                        <div className="text-secondary small">
                          {key === 'users_with_the_most_vote'
                            ? `${user.vote_count} ${t('votes')}`
                            : `${user.rank} ${t('reputation')}`}
                        </div>
                      </div>
                    </div>
                  </Col>
                ))}
              </Row>
            </Fragment>
          );
        })}
      </Col>
    </Row>
  );
};

export default Users;

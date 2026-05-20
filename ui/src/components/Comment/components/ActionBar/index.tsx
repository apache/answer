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
import { Button, Dropdown } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import classNames from 'classnames';

import { Icon, FormatTime } from '@/components';

const ActionBar = ({
  nickName,
  username,
  createdAt,
  isVote,
  voteCount = 0,
  memberActions,
  onReply,
  onVote,
  onAction,
  userStatus = '',
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'comment' });

  return (
    <div className="d-flex justify-content-between flex-wrap small">
      <div className="d-flex align-items-center flex-wrap link-secondary">
        {userStatus !== 'deleted' ? (
          <Link
            to={`/users/${username}`}
            className="name-ellipsis"
            style={{ maxWidth: '200px' }}>
            {nickName}
          </Link>
        ) : (
          <span>{nickName}</span>
        )}
        <span className="mx-1">•</span>
        <FormatTime time={createdAt} className="me-3 flex-shrink-0" />
        <Button
          title={t('tip_vote')}
          variant="link"
          size="sm"
          className={`flex-shrink-0 me-3 btn-no-border p-0 ${
            isVote ? '' : 'link-secondary'
          }`}
          onClick={onVote}>
          <Icon name="hand-thumbs-up-fill" />
          {voteCount > 0 && (
            <span className="ms-2 link-secondary">{voteCount}</span>
          )}
        </Button>
        <Button
          variant="link"
          size="sm"
          className="link-secondary m-0 p-0 btn-no-border"
          onClick={onReply}>
          {t('btn_reply')}
        </Button>
      </div>
      <div className="align-items-center control-area d-none">
        {memberActions.map((action, index) => {
          return (
            <Button
              key={action.name}
              variant="link"
              size="sm"
              className={classNames(
                'link-secondary btn-no-border m-0 p-0',
                index > 0 && 'ms-3',
              )}
              onClick={() => onAction(action)}>
              {action.name}
            </Button>
          );
        })}
      </div>
      <Dropdown className="d-block d-md-none">
        <Dropdown.Toggle
          as="div"
          variant="success"
          className="no-toggle"
          id="dropdown-comment">
          <Icon name="three-dots" className="text-secondary" />
        </Dropdown.Toggle>

        <Dropdown.Menu align="end">
          {memberActions.map((action) => {
            return (
              <Dropdown.Item key={action.name} onClick={() => onAction(action)}>
                {action.name}
              </Dropdown.Item>
            );
          })}
        </Dropdown.Menu>
      </Dropdown>
    </div>
  );
};

export default memo(ActionBar);

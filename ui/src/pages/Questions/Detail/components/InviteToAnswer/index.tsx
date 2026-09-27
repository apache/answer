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

import { memo, FC, useState, useEffect } from 'react';
import { Card, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import classNames from 'classnames';

import { Avatar } from '@/components';
import { getInviteUser, putInviteUser } from '@/services';
import type * as Type from '@/common/interface';
import { useCaptchaPlugin } from '@/utils/pluginKit';

import PeopleDropdown from './PeopleDropdown';

import './index.scss';

interface Props {
  questionId: string;
  readOnly?: boolean;
}
const MAX_INVITE_USERS = 20;

const Index: FC<Props> = ({ questionId, readOnly = false }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'invite_to_answer',
  });

  const [editing, setEditing] = useState(false);
  const [users, setUsers] = useState<Type.UserInfoBase[]>([]);
  const iaCaptcha = useCaptchaPlugin('invitation_answer');

  const initInviteUsers = () => {
    if (!questionId) {
      return;
    }
    getInviteUser(questionId)
      .then((resp) => {
        setUsers(resp);
      })
      .catch(() => {
        if (!users) {
          setUsers([]);
        }
      });
  };

  const updateInviteUsers = (user: Type.UserInfoBase) => {
    const userID = users?.find((_) => _.id === user.id);
    let userList: any = [...(users || [])];
    if (userID) {
      userList = userList?.filter((_) => {
        return _.id !== user.id;
      });
    } else {
      if (userList.length >= MAX_INVITE_USERS) {
        return;
      }
      userList.push(user);
    }
    setUsers(userList);
  };

  const submitInviteUser = () => {
    const names = users.map((_) => {
      return _.username;
    });
    const imgCode: Type.ImgCodeReq = {};
    iaCaptcha?.resolveCaptchaReq(imgCode);
    putInviteUser(questionId, names, imgCode)
      .then(async () => {
        await iaCaptcha?.close();
        setEditing(false);
      })
      .catch((ex) => {
        if (ex.isError) {
          iaCaptcha?.handleCaptchaError(ex.list);
        }
      });
  };

  const saveInviteUsers = () => {
    if (!users) {
      return;
    }

    if (!iaCaptcha) {
      submitInviteUser();
      return;
    }

    iaCaptcha.check(() => submitInviteUser());
  };

  useEffect(() => {
    initInviteUsers();
  }, [questionId]);

  const showEmpty = readOnly && users?.length === 0;

  if (showEmpty) {
    return null;
  }

  return (
    <Card className="invite-answer-card position-relative border-0 mb-4">
      <Card.Header className="text-nowrap d-flex justify-content-between text-capitalize">
        {t('title')}
        {!readOnly && (
          <Button
            onClick={() => setEditing(true)}
            variant="link"
            className="p-0">
            {t('edit', { keyPrefix: 'btns' })}
          </Button>
        )}
      </Card.Header>
      <Card.Body className={classNames('position-relative')}>
        <div className={classNames('d-flex align-items-center flex-wrap m-n1')}>
          {users?.map((user) => {
            return (
              <Link
                key={user.username}
                to={`/users/${user.username}`}
                className="mx-2 my-1 d-inline-flex flex-nowrap">
                <Avatar
                  avatar={user.avatar}
                  size="24"
                  alt={user.display_name}
                  className="rounded-1"
                />
                <small className="ms-2">{user.display_name}</small>
              </Link>
            );
          })}
          {users?.length === 0 ? (
            <div className="text-muted">{t('desc')}</div>
          ) : null}
        </div>
      </Card.Body>
      {editing && (
        <PeopleDropdown
          visible={editing}
          selectedPeople={users}
          onSelect={updateInviteUsers}
          saveInviteUsers={saveInviteUsers}
        />
      )}
    </Card>
  );
};

export default memo(Index);

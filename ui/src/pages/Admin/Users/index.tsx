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

import { FC, useEffect, useState } from 'react';
import { Form, Table, Button, Stack } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';
import dayjs from 'dayjs';

import {
  Pagination,
  FormatTime,
  BaseUserCard,
  Empty,
  QueryGroup,
  Modal,
  TabNav,
} from '@/components';
import * as Type from '@/common/interface';
import { useUserModal } from '@/hooks';
import { toastStore, loggedUserInfoStore, userCenterStore } from '@/stores';
import {
  useQueryUsers,
  addUsers,
  getAdminUcAgent,
  AdminUcAgent,
  changeUserStatus,
  deletePermanently,
} from '@/services';
import { formatCount } from '@/utils';
import { ADMIN_USERS_NAV_MENUS } from '@/common/constants';

import DeleteUserModal from './components/DeleteUserModal';
import Action from './components/Action';
import SuspenseUserModal from './components/SuspenseUserModal';

const UserFilterKeys: Type.UserFilterBy[] = [
  'normal',
  'staff',
  'inactive',
  'suspended',
  'deleted',
];

const bgMap = {
  normal: 'text-bg-success',
  suspended: 'text-bg-danger',
  deleted: 'text-bg-danger',
  inactive: 'text-bg-secondary',
};

const PAGE_SIZE = 10;
const Users: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.users' });
  const [deleteUserModalState, setDeleteUserModalState] = useState({
    show: false,
    userId: '',
  });
  const [suspenseUserModalState, setSuspenseUserModalState] = useState({
    show: false,
    userId: '',
  });
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const curFilter = urlSearchParams.get('filter') || UserFilterKeys[0];
  const curPage = Number(urlSearchParams.get('page') || '1');
  const curQuery = urlSearchParams.get('query') || '';
  const currentUser = loggedUserInfoStore((state) => state.user);
  const { agent: ucAgent } = userCenterStore();
  const [adminUcAgent, setAdminUcAgent] = useState<AdminUcAgent>({
    allow_create_user: true,
    allow_update_user_status: true,
    allow_update_user_password: true,
    allow_update_user_role: true,
  });

  const {
    data,
    isLoading,
    mutate: refreshUsers,
  } = useQueryUsers({
    page: curPage,
    page_size: PAGE_SIZE,
    query: curQuery,
    ...(curFilter === 'all'
      ? {}
      : curFilter === 'staff'
        ? { staff: true }
        : { status: curFilter }),
  });

  const userModal = useUserModal({
    onConfirm: (userModel) => {
      return new Promise((resolve, reject) => {
        addUsers(userModel)
          .then(() => {
            toastStore.getState().show({
              msg: t('user_added', { keyPrefix: 'messages' }),
              variant: 'success',
            });
            urlSearchParams.set('filter', 'normal');
            urlSearchParams.delete('page');
            setUrlSearchParams(urlSearchParams);
            refreshUsers();
            resolve(true);
          })
          .catch((e) => {
            reject(e);
          });
      });
    },
  });

  const handleFilter = (e) => {
    urlSearchParams.set('query', e.target.value);
    urlSearchParams.delete('page');
    setUrlSearchParams(urlSearchParams);
  };
  useEffect(() => {
    if (ucAgent?.enabled) {
      getAdminUcAgent().then((resp) => {
        setAdminUcAgent(resp);
      });
    }
  }, [ucAgent]);

  const changeDeleteUserModalState = (modalData: {
    show: boolean;
    userId: string;
  }) => {
    setDeleteUserModalState(modalData);
  };

  const handleDelete = (val) => {
    changeUserStatus({
      user_id: deleteUserModalState.userId,
      status: 'deleted',
      remove_all_content: val,
    }).then(() => {
      toastStore.getState().show({
        msg: t('user_deleted', { keyPrefix: 'messages' }),
        variant: 'success',
      });
      changeDeleteUserModalState({
        show: false,
        userId: '',
      });
      refreshUsers();
    });
  };

  const handleDeletePermanently = () => {
    Modal.confirm({
      title: t('title', { keyPrefix: 'delete_permanently' }),
      content: t('content', { keyPrefix: 'delete_permanently' }),
      cancelBtnVariant: 'link',
      confirmText: t('delete', { keyPrefix: 'btns' }),
      confirmBtnVariant: 'danger',
      onConfirm: () => {
        deletePermanently('users').then(() => {
          toastStore.getState().show({
            msg: t('users_deleted', { keyPrefix: 'messages' }),
            variant: 'success',
          });
          refreshUsers();
        });
      },
    });
  };

  const handleSuspenseUserModalState = (modalData: {
    show: boolean;
    userId: string;
  }) => {
    setSuspenseUserModalState(modalData);
  };

  const showAddUser =
    !ucAgent?.enabled || (ucAgent?.enabled && adminUcAgent?.allow_create_user);
  const showActionPassword =
    !ucAgent?.enabled ||
    (ucAgent?.enabled && adminUcAgent?.allow_update_user_password);

  const showActionRole =
    !ucAgent?.enabled ||
    (ucAgent?.enabled && adminUcAgent?.allow_update_user_role);

  const showActionStatus =
    !ucAgent?.enabled ||
    (ucAgent?.enabled && adminUcAgent?.allow_update_user_status);
  const showAction = showActionPassword || showActionRole || showActionStatus;

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <TabNav menus={ADMIN_USERS_NAV_MENUS} />
      <div className="d-flex flex-wrap justify-content-between align-items-center">
        <Stack direction="horizontal" gap={3} className="mb-3">
          <QueryGroup
            data={UserFilterKeys}
            currentSort={curFilter}
            sortKey="filter"
            i18nKeyPrefix="admin.users"
          />
          {curFilter === 'deleted' && Number(data?.count) > 0 ? (
            <Button
              variant="outline-danger"
              size="sm"
              onClick={() => handleDeletePermanently()}>
              {t('deleted_permanently', { keyPrefix: 'btns' })}
            </Button>
          ) : null}
          {showAddUser ? (
            <Button
              variant="outline-primary"
              size="sm"
              onClick={() => userModal.onShow()}>
              {t('add_user')}
            </Button>
          ) : null}
        </Stack>

        <Form.Control
          size="sm"
          type="search"
          value={curQuery}
          onChange={handleFilter}
          placeholder={t('filter.placeholder')}
          style={{ width: '12.25rem' }}
          className="mb-3"
        />
      </div>
      <Table responsive="md">
        <thead>
          <tr>
            <th>{t('name')}</th>
            <th style={{ width: '12%' }}>{t('reputation')}</th>
            <th style={{ width: '15%' }} className="min-w-15">
              {t('email')}
            </th>
            <th className="text-nowrap" style={{ width: '12%' }}>
              {t('created_at')}
            </th>
            {(curFilter === 'deleted' || curFilter === 'suspended') && (
              <th className="text-nowrap" style={{ width: '12%' }}>
                {curFilter === 'deleted' ? t('delete_at') : t('suspend_at')}
              </th>
            )}
            {curFilter === 'suspended' && (
              <th className="text-nowrap" style={{ width: '12%' }}>
                {t('suspend_until')}
              </th>
            )}

            <th style={{ width: '12%' }}>{t('status')}</th>
            {curFilter !== 'suspended' && curFilter !== 'deleted' && (
              <th style={{ width: '12%' }}>{t('role')}</th>
            )}
            {curFilter !== 'deleted' ? (
              <th style={{ width: '8%' }} className="text-end">
                {t('action')}
              </th>
            ) : null}
          </tr>
        </thead>
        <tbody className="align-middle">
          {data?.list.map((user) => {
            return (
              <tr key={user.user_id}>
                <td>
                  <BaseUserCard
                    data={user}
                    className="fs-6"
                    avatarSize="32px"
                    avatarSearchStr="s=48"
                    avatarClass="me-2"
                    showReputation={false}
                    nameMaxWidth="160px"
                  />
                </td>
                <td>{formatCount(user.rank)}</td>
                <td className="text-break">{user.e_mail}</td>
                <td>
                  <FormatTime time={user.created_at} />
                </td>
                {curFilter === 'suspended' && (
                  <>
                    <td className="text-nowrap">
                      <FormatTime time={user.suspended_at} />
                    </td>
                    <td className="text-nowrap">
                      {user.suspended_until <= 0 ||
                      Number(
                        dayjs(user.suspended_until * 1000).format('YYYY'),
                      ) > 2099
                        ? t('suspend_user.forever')
                        : dayjs(user.suspended_until * 1000).format(
                            t('long_date_with_time', { keyPrefix: 'dates' }),
                          )}
                    </td>
                  </>
                )}
                {curFilter === 'deleted' && (
                  <td className="text-nowrap">
                    <FormatTime time={user.deleted_at} />
                  </td>
                )}
                <td>
                  <span className={classNames('badge', bgMap[user.status])}>
                    {t(user.status)}
                  </span>
                </td>
                {curFilter !== 'suspended' && curFilter !== 'deleted' && (
                  <td>
                    <span className="badge text-bg-light">
                      {t(user.role_name)}
                    </span>
                  </td>
                )}
                {curFilter !== 'deleted' &&
                (showAction || user.status === 'inactive') ? (
                  <Action
                    userData={user}
                    showActionPassword={showActionPassword}
                    showActionRole={showActionRole}
                    showActionStatus={showActionStatus}
                    currentUser={currentUser}
                    refreshUsers={refreshUsers}
                    showDeleteModal={changeDeleteUserModalState}
                    showSuspenseModal={handleSuspenseUserModalState}
                  />
                ) : null}
              </tr>
            );
          })}
        </tbody>
      </Table>
      {Number(data?.count) <= 0 && !isLoading && <Empty />}
      <div className="mt-4 mb-2 d-flex justify-content-center">
        <Pagination
          currentPage={curPage}
          totalSize={data?.count || 0}
          pageSize={PAGE_SIZE}
        />
      </div>

      <DeleteUserModal
        show={deleteUserModalState.show}
        onClose={() => {
          changeDeleteUserModalState({
            show: false,
            userId: '',
          });
        }}
        onDelete={(val) => handleDelete(val)}
      />
      <SuspenseUserModal
        show={suspenseUserModalState.show}
        userId={suspenseUserModalState.userId}
        onClose={() => {
          handleSuspenseUserModalState({
            show: false,
            userId: '',
          });
        }}
        refreshUsers={refreshUsers}
      />
    </>
  );
};

export default Users;

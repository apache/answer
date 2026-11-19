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

import React, { useEffect, useRef, useState, FC } from 'react';
import { Dropdown } from 'react-bootstrap';

import { useSearchUserStaff } from '@/services';
import usePageUsers from '@/hooks/usePageUsers';
import * as Types from '@/common/interface';

import './index.scss';

interface IProps {
  children: React.ReactNode;
  pageUsers;
  onSelected: (val: string) => void;
}

const MAX_RECODE = 5;

const Mentions: FC<IProps> = ({ children, pageUsers, onSelected }) => {
  const menuRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [val, setValue] = useState('');
  const [users, setUsers] = useState<Types.PageUser[]>([]);
  const [cursor, setCursor] = useState(0);
  const [isRequested, setRequestedState] = useState(false);
  const { setUsers: savePageUsers } = usePageUsers();
  const { data: staffUserList = [] } = useSearchUserStaff(val);
  const mapStaffUsers =
    staffUserList
      ?.map((item) => ({
        displayName: item.display_name,
        userName: item.username,
      }))
      ?.filter(
        (item) =>
          users.findIndex((user) => user.userName === item.userName) < 0,
      ) || [];

  const searchUser = () => {
    const element = dropdownRef.current?.children[0];
    const { value, selectionStart = 0 } = element as HTMLTextAreaElement;

    if (value.indexOf('@') < 0) {
      setValue('');
    }
    if (!selectionStart) {
      return;
    }

    const str = value.substring(
      value.substring(0, selectionStart).lastIndexOf('@'),
      selectionStart,
    );

    if (str.substring(str.lastIndexOf(' '), selectionStart).indexOf('@') < 0) {
      return;
    }

    setValue(str.substring(1));

    if (!str.substring(1)) {
      return;
    }
    if (isRequested) {
      return;
    }
    setRequestedState(true);
  };

  useEffect(() => {
    const element = dropdownRef.current?.children[0] as HTMLTextAreaElement;

    if (element) {
      element.addEventListener('input', searchUser);
    }
    return () => {
      element.removeEventListener('input', searchUser);
    };
  }, [dropdownRef]);

  useEffect(() => {
    setUsers(pageUsers);
  }, [pageUsers, val]);

  const handleClick = (item) => {
    const element = dropdownRef.current?.children[0] as HTMLTextAreaElement;

    const { value, selectionStart = 0 } = element;

    if (!selectionStart) {
      return;
    }

    if (item?.userName) {
      savePageUsers(item);
    }
    const text = `@${item?.userName} `;
    onSelected(
      `${value.substring(
        0,
        value.substring(0, selectionStart).lastIndexOf('@'),
      )}${text}${value.substring(selectionStart)}`,
    );
    setUsers([]);
    setValue('');
  };
  const filterData = val
    ? [...users, ...mapStaffUsers].filter(
        (item) =>
          item.displayName?.indexOf(val) === 0 ||
          item.userName?.indexOf(val) === 0,
      )
    : [];
  const handleKeyDown = (e) => {
    const { keyCode } = e;

    if (keyCode === 38 && cursor > 0) {
      e.preventDefault();
      setCursor(cursor - 1);
    }
    if (keyCode === 40 && cursor < filterData.length - 1) {
      e.preventDefault();

      setCursor(cursor + 1);
    }
    if (keyCode === 13 && cursor > -1 && cursor <= filterData.length - 1) {
      e.preventDefault();

      const item = filterData[cursor];

      handleClick(item);
      setCursor(0);
    }
  };

  return (
    <Dropdown
      ref={dropdownRef}
      className="mentions-wrap"
      show={filterData.length > 0}
      onKeyDown={handleKeyDown}>
      {children}
      <Dropdown.Menu
        className={filterData.length > 0 ? 'visible' : 'invisible'}
        ref={menuRef}>
        {filterData
          .filter((_, index) => index < MAX_RECODE)
          .map((item, index) => {
            return (
              <Dropdown.Item
                className={`${cursor === index ? 'bg-gray-200' : ''}`}
                key={item.displayName}
                onClick={() => handleClick(item)}>
                <span className="link-dark me-1">{item.displayName}</span>
                <small className="link-secondary">@{item.userName}</small>
              </Dropdown.Item>
            );
          })}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default Mentions;

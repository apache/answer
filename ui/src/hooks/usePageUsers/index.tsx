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

import { useState } from 'react';

import uniqBy from 'lodash/uniqBy';

import * as Types from '@/common/interface';

let globalUsers: Types.PageUser[] = [];
const usePageUsers = () => {
  const [users, setUsers] = useState<Types.PageUser[]>(globalUsers);
  const getUsers = () => {
    return users;
  };
  return {
    getUsers,
    setUsers: (data: Types.PageUser | Types.PageUser[]) => {
      if (data instanceof Array) {
        if (data.length === 0) {
          return;
        }
        setUsers(uniqBy([...users, ...data], 'userName'));
        globalUsers = uniqBy([...globalUsers, ...data], 'userName');
      } else {
        setUsers(uniqBy([...users, data], 'userName'));
        globalUsers = uniqBy([...globalUsers, data], 'userName');
      }
    },
  };
};

export default usePageUsers;

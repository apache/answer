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
import { Nav } from 'react-bootstrap';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

interface Props {
  slug: string;
  isSelf: boolean;
  tabName: string;
}

const list = [
  {
    path: '',
    name: 'overview',
  },
  {
    path: '/answers',
    name: 'answers',
  },
  {
    path: '/questions',
    name: 'questions',
  },
  {
    role: 'self', // Only visible to author
    path: '/bookmarks',
    name: 'bookmarks',
  },
  {
    path: '/comments',
    name: 'comments',
  },
  {
    role: 'self', // Only visible to author
    path: '/votes',
    name: 'votes',
  },
  {
    path: '/badges',
    name: 'badges',
  },
];
const Index: FC<Props> = ({ slug, tabName = 'overview', isSelf }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });
  return (
    <Nav
      className="pt-2 mb-4 flex-nowrap"
      variant="pills"
      style={{ overflow: 'auto' }}>
      {list.map((item) => {
        if (item.role && !isSelf) {
          return null;
        }
        if (item.path) {
          return (
            <NavLink
              to={`/users/${slug}${item.path}`}
              key={item.name}
              className="nav-link">
              {t(item.name)}
            </NavLink>
          );
        }
        return (
          <NavLink
            key={item.name}
            to={`/users/${slug}`}
            className={({ isActive }) =>
              isActive && tabName === 'overview'
                ? 'nav-link active'
                : 'nav-link'
            }>
            {t(item.name)}
          </NavLink>
        );
      })}
    </Nav>
  );
};

export default memo(Index);

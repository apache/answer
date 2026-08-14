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
import { useTranslation } from 'react-i18next';

import { QueryGroup } from '@/components';

const sortBtns = ['newest', 'score'];

interface Props {
  tabName: string;
  count: number;
  sort: string;
  visible: boolean;
}
const Index: FC<Props> = ({
  tabName = 'comments',
  visible,
  sort,
  count = 0,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });

  if (!visible) {
    return null;
  }

  return (
    <div className="d-flex  align-items-center justify-content-between pb-3">
      <h5 className="mb-0">
        {count} {t(tabName)}
      </h5>
      {(tabName === 'comments' || tabName === 'posts') && (
        <QueryGroup
          data={sortBtns}
          currentSort={sort}
          i18nKeyPrefix="personal"
        />
      )}
    </div>
  );
};

export default memo(Index);

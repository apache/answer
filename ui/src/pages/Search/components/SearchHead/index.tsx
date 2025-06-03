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

const sortBtns = ['relevance', 'newest', 'active', 'score'];

interface Props {
  count: number;
  sort: string;
}
const Index: FC<Props> = ({ sort, count = 0 }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'search.sort_btns',
  });

  return (
    <div className="d-flex flex-wrap align-items-center justify-content-between pt-2 pb-3">
      <h5 className="mb-0">
        {count === -1
          ? t('counts_loading', { keyPrefix: 'search' })
          : t('counts', { count, keyPrefix: 'search' })}
      </h5>
      <QueryGroup
        data={sortBtns}
        currentSort={sort}
        sortKey="order"
        i18nKeyPrefix="search.sort_btns"
      />
    </div>
  );
};

export default memo(Index);

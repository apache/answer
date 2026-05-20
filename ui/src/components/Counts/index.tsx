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

import classname from 'classnames';

import { Icon } from '@/components';
import { formatCount } from '@/utils/common';

interface Props {
  data: {
    votes: number;
    answers: number;
    views: number;
  };
  showVotes?: boolean;
  showAnswers?: boolean;
  showViews?: boolean;
  showAccepted?: boolean;
  isAccepted?: boolean;
  className?: string;
}
const Index: FC<Props> = ({
  data,
  showVotes = true,
  showAnswers = true,
  showViews = true,
  isAccepted = false,
  showAccepted = false,
  className = '',
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'counts' });

  return (
    <div className={classname('d-flex align-items-center', className)}>
      {showVotes && (
        <div className="d-flex align-items-center flex-shrink-0 text-body">
          <Icon name="hand-thumbs-up-fill me-1" />
          <span className="fw-medium">{data.votes}</span>
          <span className="ms-1">{t('votes')}</span>
        </div>
      )}

      {showAccepted && (
        <div className="d-flex align-items-center ms-3 text-success flex-shrink-0">
          <Icon name="check-circle-fill me-1" />
          <span>{t('accepted')}</span>
        </div>
      )}

      {showAnswers && (
        <div
          className={`d-flex flex-shrink-0 align-items-center ms-3 ${
            isAccepted ? 'text-bg-success rounded-pill px-2 ' : ''
          }`}>
          {isAccepted ? (
            <Icon name="check-circle-fill me-1" />
          ) : (
            <Icon name="chat-square-text-fill me-1" />
          )}
          <span className="fw-medium">{data.answers}</span>
          <span className="ms-1">{t('answers')}</span>
        </div>
      )}
      {showViews && (
        <span
          className={classname(
            'summary-stat ms-3 flex-shrink-0',
            data.views >= 100 * 1000
              ? 'view-level3'
              : data.views >= 10000
                ? 'view-level2'
                : data.views >= 1000
                  ? 'view-level1'
                  : '',
          )}>
          <Icon name="bar-chart-fill" />
          <span className="fw-medium ms-1">{formatCount(data.views)}</span>
          <span className="ms-1">{t('views')}</span>
        </span>
      )}
    </div>
  );
};

export default memo(Index);

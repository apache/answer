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

import { memo, FC } from 'react';
import { Alert } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';

interface Props {
  data;
}
const Index: FC<Props> = ({ data }) => {
  const { t } = useTranslation();
  return (
    <Alert className="mb-4" variant={data.level}>
      {data.level === 'info' ? (
        <div>
          {data.msg.startsWith('http') ? (
            <p>
              {data.description}{' '}
              <a href={data.msg} className="alert-exist">
                <strong>{t('question_detail.show_exist')}</strong>
              </a>
            </p>
          ) : (
            <p>{data.msg ? data.msg : data.description}</p>
          )}
          <div className="small">
            {t('question_detail.closed_in')}{' '}
            <time
              dateTime={dayjs.unix(data.time).tz().toISOString()}
              title={dayjs
                .unix(data.time)
                .tz()
                .format(t('dates.long_date_with_time'))}>
              {dayjs
                .unix(data.time)
                .tz()
                .format(t('dates.long_date_with_year'))}
            </time>
            .
          </div>
        </div>
      ) : (
        data.msg
      )}
    </Alert>
  );
};

export default memo(Index);

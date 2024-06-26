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

import { memo, FC, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { following } from '@/services';
import { tryNormalLogged } from '@/utils/guard';
import { escapeRemove } from '@/utils';
import { pathFactory } from '@/router/pathFactory';
import { PluginRender } from '@/components';
import Pattern from '@/common/pattern';

interface Props {
  data;
}

const Index: FC<Props> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'search' });
  const [searchParams] = useSearchParams();
  const q = searchParams.get('q');
  const options = q?.match(Pattern.search);
  const [followed, setFollowed] = useState(data?.is_follower);

  const follow = () => {
    if (!tryNormalLogged(true)) {
      return;
    }
    following({
      object_id: data?.tag_id,
      is_cancel: followed,
    }).then((res) => {
      setFollowed(res.is_followed);
    });
  };

  return (
    <div className="mb-5">
      <div className="mb-3 d-flex align-items-center justify-content-between">
        <h3 className="mb-0">{t('title')}</h3>

        <PluginRender type="search" slug_name="serarch_info" />
      </div>
      <p>
        <span className="text-secondary me-1">{t('keywords')}</span>
        {q?.replace(Pattern.search, '')}
        <br />
        {options?.length && (
          <>
            <span className="text-secondary">{t('options')} </span>
            {options?.map((item) => {
              return <code key={item}>{item} </code>;
            })}
          </>
        )}
      </p>
      {data?.slug_name && (
        <>
          {data.excerpt && (
            <p className="text-break">
              {escapeRemove(data.excerpt)}
              <Link className="ms-1" to={pathFactory.tagInfo(data.slug_name)}>
                [{t('more')}]
              </Link>
            </p>
          )}

          <Button variant="outline-primary" onClick={follow}>
            {followed ? t('following') : t('follow')}
          </Button>
        </>
      )}
    </div>
  );
};

export default memo(Index);

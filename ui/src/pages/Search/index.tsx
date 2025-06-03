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

import { Row, Col, ListGroup } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { useEffect, useState } from 'react';

import { usePageTags, useSkeletonControl } from '@/hooks';
import { useCaptchaPlugin } from '@/utils/pluginKit';
import { Pagination } from '@/components';
import { getSearchResult } from '@/services';
import type { SearchParams, SearchRes } from '@/common/interface';

import {
  Head,
  SearchHead,
  SearchItem,
  Tips,
  Empty,
  ListLoader,
} from './components';

const Index = () => {
  const { t } = useTranslation('translation');
  const [searchParams] = useSearchParams();
  const page = searchParams.get('page') || 1;
  const q = searchParams.get('q') || '';
  const order = searchParams.get('order') || 'relevance';
  const [isLoading, setIsLoading] = useState(false);
  const { isSkeletonShow } = useSkeletonControl(isLoading);
  const [data, setData] = useState<SearchRes>({
    count: 0,
    list: [],
    extra: null,
  });
  const { count = 0, list = [], extra = null } = data || {};

  const searchCaptcha = useCaptchaPlugin('search');

  const doSearch = () => {
    setIsLoading(true);
    const params: SearchParams = {
      q,
      order,
      page: Number(page),
      size: 20,
    };

    const captcha = searchCaptcha?.getCaptcha();
    if (captcha?.verify) {
      params.captcha_id = captcha.captcha_id;
      params.captcha_code = captcha.captcha_code;
    }

    getSearchResult(params)
      .then(async (resp) => {
        await searchCaptcha?.close();
        setData(resp);
      })
      .catch((err) => {
        if (err.isError) {
          searchCaptcha?.handleCaptchaError(err.list);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  useEffect(() => {
    if (!searchCaptcha) {
      doSearch();
      return;
    }
    searchCaptcha.check(() => {
      doSearch();
    });
  }, [q, order, page]);

  let pageTitle = t('search', { keyPrefix: 'page_title' });
  if (q) {
    pageTitle = `${t('posts_containing', { keyPrefix: 'page_title' })} '${q}'`;
  }
  usePageTags({
    title: pageTitle,
  });

  return (
    <Row className="pt-4 mb-5">
      <Col className="page-main flex-auto">
        <Head data={extra} />
        <SearchHead sort={order} count={isLoading ? -1 : count} />
        <ListGroup className="rounded-0 mb-5">
          {isSkeletonShow ? (
            <ListLoader />
          ) : (
            list?.map((item) => {
              return <SearchItem key={item.object.id} data={item} />;
            })
          )}
        </ListGroup>

        {!isLoading && !list?.length && <Empty />}

        <div className="d-flex justify-content-center">
          <Pagination
            currentPage={Number(page)}
            pageSize={20}
            totalSize={count}
          />
        </div>
      </Col>
      <Col className="page-right-side mt-4 mt-xl-0">
        <Tips />
      </Col>
    </Row>
  );
};

export default Index;

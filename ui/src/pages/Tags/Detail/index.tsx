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
import { Row, Col, Button } from 'react-bootstrap';
import {
  useParams,
  Link,
  useNavigate,
  useSearchParams,
} from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { usePageTags } from '@/hooks';
import * as Type from '@/common/interface';
import { FollowingTags, CustomSidebar, Icon } from '@/components';
import {
  useTagInfo,
  useFollow,
  useQuerySynonymsTags,
  useQuestionList,
} from '@/services';
import QuestionList, { QUESTION_ORDER_KEYS } from '@/components/QuestionList';
import HotQuestions from '@/components/HotQuestions';
import { guard, pageTitleType } from '@/utils';
import { pathFactory } from '@/router/pathFactory';

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'tags' });
  const navigate = useNavigate();
  const routeParams = useParams();
  const curTagName = routeParams.tagName || '';
  const [urlSearchParams] = useSearchParams();
  const curOrder = (urlSearchParams.get('order') ||
    QUESTION_ORDER_KEYS[0]) as Type.QuestionOrderBy;
  const curPage = Number(urlSearchParams.get('page')) || 1;
  const reqParams: Type.QueryQuestionsReq = {
    page_size: 20,
    page: curPage,
    order: curOrder as Type.QuestionOrderBy,
    tag: routeParams.tagName,
  };
  const [tagInfo, setTagInfo] = useState<any>({});
  const [tagFollow, setTagFollow] = useState<Type.FollowParams>();
  const { data: tagResp, isLoading } = useTagInfo({ name: curTagName });
  const { data: listData, isLoading: listLoading } = useQuestionList(reqParams);
  const { data: followResp } = useFollow(tagFollow);
  const { data: synonymsRes } = useQuerySynonymsTags(
    tagInfo?.tag_id,
    tagInfo?.status,
  );
  const toggleFollow = () => {
    if (!guard.tryNormalLogged(true)) {
      return;
    }
    setTagFollow({
      is_cancel: tagInfo.is_follower,
      object_id: tagInfo.tag_id,
    });
  };

  useEffect(() => {
    if (tagResp) {
      const info = { ...tagResp };
      if (info.main_tag_slug_name) {
        navigate(pathFactory.tagLanding(info.main_tag_slug_name), {
          replace: true,
        });
        return;
      }
      if (followResp) {
        info.is_follower = followResp.is_followed;
      }

      if (info.excerpt) {
        info.excerpt =
          info.excerpt.length > 256
            ? [...info.excerpt].slice(0, 256).join('')
            : info.excerpt;
      }

      setTagInfo(info);
    }
  }, [tagResp, followResp]);
  let pageTitle = '';
  if (tagInfo?.display_name) {
    pageTitle = `'${tagInfo.display_name}' ${t(pageTitleType(), {
      keyPrefix: 'page_title',
    })}`;
  }
  const keywords: string[] = [];
  if (tagInfo?.slug_name) {
    keywords.push(tagInfo.slug_name);
  }
  synonymsRes?.synonyms.forEach((o) => {
    keywords.push(o.slug_name);
  });
  usePageTags({
    title: pageTitle,
    description: tagInfo?.description,
    keywords: keywords.join(','),
  });
  return (
    <Row className="pt-4 mb-5">
      <Col className="page-main flex-auto">
        {isLoading ? (
          <div className="tag-box mb-5 placeholder-glow">
            <div className="mb-3 h3 placeholder" style={{ width: '120px' }} />
            <p
              className="placeholder w-100 d-block align-top"
              style={{ height: '24px' }}
            />

            <div
              className="placeholder d-block align-top"
              style={{ height: '38px', width: '100px' }}
            />
          </div>
        ) : (
          <div className="tag-box mb-5">
            <h3 className="mb-3">
              <Link
                to={pathFactory.tagLanding(tagInfo.slug_name)}
                replace
                className="link-dark">
                {tagInfo.display_name}
              </Link>
            </h3>

            <div
              className="text-break"
              dangerouslySetInnerHTML={{ __html: tagInfo.excerpt }}
            />

            <div className="box-ft">
              {tagInfo.is_follower ? (
                <div>
                  <Button variant="primary" onClick={() => toggleFollow()}>
                    {t('button_following')}
                  </Button>
                  <Link
                    to={pathFactory.tagInfo(curTagName)}
                    className="btn btn-outline-secondary ms-2">
                    {t('wiki')}
                  </Link>
                  <Link
                    className="btn btn-outline-secondary ms-2"
                    to="/users/settings/notify">
                    <Icon name="bell-fill" />
                  </Link>
                </div>
              ) : (
                <div>
                  <Button
                    variant="outline-primary"
                    onClick={() => toggleFollow()}>
                    {t('button_follow')}
                  </Button>
                  <Link
                    to={pathFactory.tagInfo(curTagName)}
                    className="btn btn-outline-secondary ms-2">
                    {t('wiki')}
                  </Link>
                </div>
              )}
            </div>
          </div>
        )}
        <QuestionList
          source="tag"
          data={listData}
          order={curOrder}
          orderList={QUESTION_ORDER_KEYS.filter((k) => k !== 'recommend')}
          isLoading={listLoading}
        />
      </Col>
      <Col className="page-right-side mt-4 mt-xl-0">
        <CustomSidebar />
        <FollowingTags />
        <HotQuestions />
      </Col>
    </Row>
  );
};

export default Index;

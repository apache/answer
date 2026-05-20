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

import { FC } from 'react';
import { Row, Col } from 'react-bootstrap';
import { useMatch, Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { usePageTags } from '@/hooks';
import {
  FollowingTags,
  QuestionList,
  HotQuestions,
  CustomSidebar,
} from '@/components';
import {
  siteInfoStore,
  loggedUserInfoStore,
  loginSettingStore,
} from '@/stores';
import { useQuestionList, useQuestionRecommendList } from '@/services';
import * as Type from '@/common/interface';
import { userCenter, floppyNavigation } from '@/utils';
import { QUESTION_ORDER_KEYS } from '@/components/QuestionList';

const Questions: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'question' });
  const { t: t2 } = useTranslation('translation');
  const { user: loggedUser } = loggedUserInfoStore((_) => _);
  const [urlSearchParams] = useSearchParams();
  const curPage = Number(urlSearchParams.get('page')) || 1;
  const curOrder = (urlSearchParams.get('order') ||
    QUESTION_ORDER_KEYS[0]) as Type.QuestionOrderBy;
  const reqParams: Type.QueryQuestionsReq = {
    page_size: 20,
    page: curPage,
    order: curOrder as Type.QuestionOrderBy,
  };
  const { data: listData, isLoading: listLoading } =
    curOrder === 'recommend'
      ? useQuestionRecommendList(reqParams)
      : useQuestionList(reqParams);
  const isIndexPage = useMatch('/');
  let pageTitle = t('questions', { keyPrefix: 'page_title' });
  let slogan = '';
  const { siteInfo } = siteInfoStore();
  if (isIndexPage) {
    pageTitle = `${siteInfo.name}`;
    slogan = `${siteInfo.short_description}`;
  }
  const { login: loginSetting } = loginSettingStore();

  usePageTags({ title: pageTitle, subtitle: slogan });
  return (
    <Row className="pt-4 mb-5">
      <Col className="page-main flex-auto overflow-x-hidden">
        <QuestionList
          source="questions"
          data={listData}
          order={curOrder}
          orderList={
            loggedUser.username
              ? QUESTION_ORDER_KEYS
              : QUESTION_ORDER_KEYS.filter((key) => key !== 'recommend')
          }
          isLoading={listLoading}
        />
      </Col>
      <Col className="page-right-side mt-4 mt-xl-0">
        <CustomSidebar />
        {!loggedUser.username && (
          <div className="card mb-4">
            <div className="card-body">
              <h5 className="card-title">
                {t2('website_welcome', {
                  site_name: siteInfo.name,
                })}
              </h5>
              <p className="card-text">{siteInfo.description}</p>
              <Link
                to={userCenter.getLoginUrl()}
                className="btn btn-primary"
                onClick={floppyNavigation.handleRouteLinkClick}>
                {t('login', { keyPrefix: 'btns' })}
              </Link>
              {loginSetting.allow_new_registrations ? (
                <Link
                  to={userCenter.getSignUpUrl()}
                  className="btn btn-link ms-2"
                  onClick={floppyNavigation.handleRouteLinkClick}>
                  {t('signup', { keyPrefix: 'btns' })}
                </Link>
              ) : null}
            </div>
          </div>
        )}
        {loggedUser.access_token && <FollowingTags />}
        <HotQuestions />
      </Col>
    </Row>
  );
};

export default Questions;

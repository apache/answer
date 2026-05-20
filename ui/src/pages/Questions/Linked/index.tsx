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
import { Row, Col } from 'react-bootstrap';
import { useParams, useSearchParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { usePageTags } from '@/hooks';
import { useQuestionLink, questionDetail } from '@/services';
import * as Type from '@/common/interface';
import {
  QuestionList,
  CustomSidebar,
  HotQuestions,
  FollowingTags,
} from '@/components';
import { userCenter, floppyNavigation } from '@/utils';
import { QUESTION_ORDER_KEYS } from '@/components/QuestionList';
import {
  loggedUserInfoStore,
  siteInfoStore,
  loginSettingStore,
} from '@/stores';

const LinkedQuestions: FC = () => {
  const { qid } = useParams<{ qid: string }>();
  const { t } = useTranslation('translation', { keyPrefix: 'linked_question' });
  const { t: t2 } = useTranslation('translation');
  const { user: loggedUser } = loggedUserInfoStore((_) => _);
  const [urlSearchParams] = useSearchParams();
  const curPage = Number(urlSearchParams.get('page')) || 1;
  const curOrder = (urlSearchParams.get('order') ||
    QUESTION_ORDER_KEYS[0]) as Type.QuestionOrderBy;
  const pageSize = 10;
  const { siteInfo } = siteInfoStore();
  const { data: listData, isLoading: listLoading } = useQuestionLink({
    question_id: qid || '',
    page: curPage,
    page_size: pageSize,
    order: curOrder,
  });
  const { login: loginSetting } = loginSettingStore();
  const [questionTitle, setQuestionTitle] = useState('');

  useEffect(() => {
    questionDetail(qid || '')
      .then((res) => {
        setQuestionTitle(res.title);
      })
      .catch((err) => {
        console.error('get question detail failed:', err);
        setQuestionTitle(`#${qid}`);
      });
  }, []);
  usePageTags({
    title: t('title'),
  });

  console.log(
    'listData',
    QUESTION_ORDER_KEYS.filter((v) => v !== 'unanswered').slice(0, 5),
  );

  return (
    <Row className="pt-4 mb-5">
      <Col className="page-main flex-auto">
        <h3 className="mb-3">{t('title')}</h3>
        <div className="mb-5">
          {t('description')}&nbsp;
          <a href={`/questions/${qid}`}>{questionTitle}</a>
        </div>
        <QuestionList
          source="linked"
          data={listData}
          order={curOrder}
          orderList={QUESTION_ORDER_KEYS.filter(
            (v) => v !== 'unanswered',
          ).slice(0, 5)}
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

export default LinkedQuestions;

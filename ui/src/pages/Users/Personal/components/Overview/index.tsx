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
import { Row, Col } from 'react-bootstrap';

import TopList from '../TopList';

interface Props {
  visible: boolean;
  introduction: string;
  data;
}
const Index: FC<Props> = ({ visible, introduction, data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });
  if (!visible) {
    return null;
  }
  return (
    <div>
      <h5 className="mb-3">{t('about_me')}</h5>
      {introduction ? (
        <div
          className="mb-5 text-break fmt"
          dangerouslySetInnerHTML={{ __html: introduction }}
        />
      ) : (
        <div className="mb-5">{t('about_me_empty')}</div>
      )}

      <Row className="mb-4">
        <Col sm={12} md={6} className="mb-4">
          <h5 className="mb-3">{t('top_comments')}</h5>
          {data?.comments?.length > 0 ? (
            <TopList data={data.comments} type="comment" />
          ) : (
            <div className="mb-5">{t('content_empty')}</div>
          )}
        </Col>
        <Col sm={12} md={6}>
          <h5 className="mb-3">{t('top_posts')}</h5>
          {data?.posts?.length > 0 ? (
            <TopList data={data.posts} type="post" />
          ) : (
            <div className="mb-5">{t('content_empty')}</div>
          )}
        </Col>
      </Row>
    </div>
  );
};

export default memo(Index);

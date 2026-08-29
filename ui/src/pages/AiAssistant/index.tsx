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

import { useEffect, useState } from 'react';
import { Row, Col, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';

import classNames from 'classnames';

import * as Type from '@/common/interface';
import { Icon } from '@/components';
import { getConversationDetail, getConversationList } from '@/services';
import { usePageTags } from '@/hooks';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import { Storage } from '@/utils';

import ConversationsList from './components/ConversationList';
import AnswerChatBot from '@/components/AnswerChatBot';

interface ConversationListItem {
  conversation_id: string;
  topic: string;
}

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const [isShowConversationList, setIsShowConversationList] = useState(false);
  const [recentNewItem, setRecentNewItem] = useState<any>(null);
  const [conversions, setConversions] = useState<Type.ConversationDetail>({
    records: [],
    conversation_id: '',
    created_at: 0,
    topic: '',
    updated_at: 0,
  });
  const navigate = useNavigate();
  const { id = '' } = useParams<{ id: string }>();
  const [conversationsPage, setConversationsPage] = useState(1);

  const [conversationsList, setConversationsList] = useState<{
    count: number;
    list: ConversationListItem[];
  }>({
    count: 0,
    list: [],
  });

  const guest = !Storage.get(LOGGED_TOKEN_STORAGE_KEY);

  const resetPageState = () => {
    setConversions({
      records: [],
      conversation_id: '',
      created_at: 0,
      topic: '',
      updated_at: 0,
    });
    setRecentNewItem(null);
  };

  const handleNewConversation = (e) => {
    e.preventDefault();
    navigate('/ai-assistant', { replace: true });
  };

  const fetchDetail = () => {
    getConversationDetail(id).then((res) => {
      setConversions(res);
    });
  };

  usePageTags({
    title: conversions?.topic || t('ai_assistant', { keyPrefix: 'page_title' }),
  });

  useEffect(() => {
    if (id) {
      fetchDetail();
    } else {
      resetPageState();
    }
  }, [id]);

  const getList = (p) => {
    getConversationList({
      page: p,
      page_size: 10,
    }).then((res) => {
      setConversationsList({
        count: res.count,
        list: [...conversationsList.list, ...res.list],
      });
    });
  };

  const getMore = (e) => {
    e.preventDefault();
    setConversationsPage((prev) => prev + 1);
    getList(conversationsPage + 1);
  };

  useEffect(() => {
    getList(1);

    return () => {
      setConversationsList({
        count: 0,
        list: [],
      });
      setConversationsPage(1);
    };
  }, []);

  useEffect(() => {
    if (recentNewItem && recentNewItem.conversation_id) {
      setConversationsList((prev) => ({
        ...prev,
        list: [
          recentNewItem,
          ...prev.list.filter(
            (item) => item.conversation_id !== recentNewItem.conversation_id,
          ),
        ],
      }));
    }
  }, [recentNewItem]);

  return (
    <div className="pt-4 d-flex flex-column flex-grow-1 position-relative">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h3 className="mb-0">
          {t('ai_assistant', { keyPrefix: 'page_title' })}
        </h3>
        <div>
          <Button
            variant="outline-primary"
            href="/ai-assistant"
            className="me-2"
            size="sm"
            onClick={handleNewConversation}>
            {t('new')}
          </Button>
          <Button
            variant={isShowConversationList ? 'secondary' : 'outline-secondary'}
            size="sm"
            title={t('recent_conversations')}
            onClick={() => setIsShowConversationList(!isShowConversationList)}>
            <Icon name="clock-history" />
          </Button>
        </div>
      </div>
      <Row
        className={classNames(
          'flex-grow-1',
          !isShowConversationList ? 'justify-content-center' : '',
        )}>
        <Col
          className={classNames(
            'page-main flex-auto d-flex flex-column flex-grow-1',
          )}
          style={{ maxWidth: '772px' }}>
          <div
            style={{
              flex: '1 1 auto',
              minHeight: 0,
              display: 'flex',
              flexDirection: 'column',
            }}>
            <AnswerChatBot
              conversationId={id}
              initialRecords={conversions?.records || []}
              guest={guest}
              onConversationCreated={(cid) => {
                if (id !== cid) {
                  navigate(`/ai-assistant/${cid}`, { replace: true });
                }
              }}
              height="100%">
              <div
                slot="sender-footer-prefix"
                className="small text-secondary ps-2 text-truncate">
                {t('ai_disclaimer')}
              </div>
            </AnswerChatBot>
          </div>
        </Col>
        {isShowConversationList && (
          <Col className="page-right-side mt-4 mt-xl-0">
            <ConversationsList data={conversationsList} loadMore={getMore} />
          </Col>
        )}
      </Row>
    </div>
  );
};

export default Index;

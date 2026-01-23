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
import { Row, Col, Spinner, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';

import classNames from 'classnames';
import { v4 as uuidv4 } from 'uuid';

import * as Type from '@/common/interface';
import requestAi, { cancelCurrentRequest } from '@/utils/requestAi';
import { Sender, BubbleUser, BubbleAi, Icon } from '@/components';
import { getConversationDetail, getConversationList } from '@/services';
import { usePageTags } from '@/hooks';
import { Storage } from '@/utils';

import ConversationsList from './components/ConversationList';

interface ConversationListItem {
  conversation_id: string;
  topic: string;
}

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const [isShowConversationList, setIsShowConversationList] = useState(false);
  const [isGenerate, setIsGenerate] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
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
  const [temporaryBottomSpace, setTemporaryBottomSpace] = useState(0);
  const [conversationsPage, setConversationsPage] = useState(1);

  const [conversationsList, setConversationsList] = useState<{
    count: number;
    list: ConversationListItem[];
  }>({
    count: 0,
    list: [],
  });

  const calculateTemporarySpace = () => {
    const viewportHeight = window.innerHeight;
    const navHeight = 64;
    const senderHeight = (document.querySelector('.sender-wrap') as HTMLElement)
      ?.offsetHeight;
    const neededSpace = viewportHeight - senderHeight - navHeight - 120;
    const height = neededSpace;
    console.log('lasMsgHeight', height);

    setTemporaryBottomSpace(height);
  };

  const resetPageState = () => {
    setConversions({
      records: [],
      conversation_id: '',
      created_at: 0,
      topic: '',
      updated_at: 0,
    });
    setIsGenerate(false);
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

  const handleSubmit = async (userMsg) => {
    setIsLoading(true);
    if (conversions?.records.length === 0) {
      setRecentNewItem({
        conversation_id: id,
        topic: userMsg,
      });
    }
    const chatId = Date.now();
    setConversions((prev) => ({
      ...prev,
      topic: userMsg,
      conversation_id: id,
      records: [
        ...prev.records,
        {
          id: chatId,
          role: 'user',
          content: userMsg,
          chat_completion_id: String(chatId), // Add required properties
          helpful: 0,
          unhelpful: 0,
          created_at: chatId,
        },
      ],
    }));

    // scroll to user message after the page height is stable
    requestAnimationFrame(() => {
      const userBubbles = document.querySelectorAll('.bubble-user-wrap');
      const lastUserBubble = userBubbles[userBubbles.length - 1];
      if (lastUserBubble) {
        lastUserBubble.scrollIntoView({
          behavior: 'smooth',
          block: 'start',
        });
      }
    });

    calculateTemporarySpace();

    const params = {
      conversation_id: id,
      messages: [
        {
          role: 'user',
          content: userMsg,
        },
      ],
    };

    await requestAi('/answer/api/v1/chat/completions', {
      body: JSON.stringify(params),
      onMessage: (res) => {
        if (!res.choices[0].delta?.content) {
          return;
        }
        setIsLoading(false);
        setIsGenerate(true);
        setConversions((prev) => {
          const updatedRecords = [...prev.records];
          const lastConversion = updatedRecords[updatedRecords.length - 1];
          if (lastConversion?.chat_completion_id === res?.chat_completion_id) {
            updatedRecords[updatedRecords.length - 1] = {
              ...lastConversion,
              content: lastConversion.content + res.choices[0].delta.content,
            };
          } else {
            updatedRecords.push({
              chat_completion_id: res.chat_completion_id,
              role: res.choices[0].delta.role || 'assistant',
              content: res.choices[0].delta.content,
              helpful: 0,
              unhelpful: 0,
              created_at: Date.now(),
            });
          }
          return {
            ...prev,
            conversation_id: params.conversation_id,
            records: updatedRecords,
          };
        });
      },
      onError: (error) => {
        setIsLoading(false);
        setIsGenerate(false);
        console.error('Error:', error);
      },
      onComplete: () => {
        setIsGenerate(false);
        setIsLoading(false);
      },
    });
  };

  const handleSender = (userMsg) => {
    if (conversions?.records.length <= 0) {
      const newConversationId = uuidv4();
      navigate(`/ai-assistant/${newConversationId}`);
      Storage.set('_a_once_msg', userMsg);
    } else {
      handleSubmit(userMsg);
    }
  };

  const handleCancel = () => {
    if (cancelCurrentRequest()) {
      setIsGenerate(false);
    }
  };

  usePageTags({
    title: conversions?.topic || t('ai_assistant', { keyPrefix: 'page_title' }),
  });

  useEffect(() => {
    if (id) {
      const msg = Storage.get('_a_once_msg');
      Storage.remove('_a_once_msg');
      if (msg) {
        if (msg) {
          handleSubmit(msg);
        }
        return;
      }
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
            !conversions?.conversation_id ? 'justify-content-center' : '',
          )}
          style={{ maxWidth: '772px' }}>
          {conversions?.records.length > 0 && (
            <div className="flex-grow-1 pb-5">
              {conversions?.records.map((item, index) => {
                const isLastMessage =
                  index === Number(conversions?.records.length) - 1;
                return (
                  <div
                    key={`${item.chat_completion_id}-${item.role}`}
                    className={`${isLastMessage ? '' : 'mb-4'}`}>
                    {item.role === 'user' ? (
                      <BubbleUser content={item.content} />
                    ) : (
                      <BubbleAi
                        minHeight={isLastMessage ? temporaryBottomSpace : 0}
                        canType={isGenerate && isLastMessage}
                        chatId={item.chat_completion_id}
                        isLast={isLastMessage}
                        isCompleted={!isGenerate}
                        content={item.content}
                        actionData={{
                          helpful: item.helpful,
                          unhelpful: item.unhelpful,
                        }}
                      />
                    )}
                  </div>
                );
              })}

              {temporaryBottomSpace > 0 && isLoading && (
                <div
                  style={{
                    height: `${temporaryBottomSpace}px`,
                  }}>
                  {isLoading && (
                    <Spinner
                      animation="border"
                      size="sm"
                      variant="secondary"
                      className="mt-4"
                    />
                  )}
                </div>
              )}
            </div>
          )}
          {conversions?.conversation_id ? null : (
            <h5 className="text-center mb-3">{t('description')}</h5>
          )}
          <Sender
            onSubmit={handleSender}
            onCancel={handleCancel}
            isGenerate={isGenerate || isLoading}
            hasConversation={!!conversions?.conversation_id}
          />
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

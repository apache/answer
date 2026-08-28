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

import React, { useEffect, useMemo, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import { ChatBot } from '@tdesign-react/chat';
import type {
  AIMessageContent,
  ChatMessagesData,
  ChatServiceConfig,
  SSEChunkData,
} from '@tdesign-react/chat';
import 'tdesign-react/dist/tdesign.css';

import { voteConversation } from '@/services';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import Storage from '@/utils/storage';

export interface ConversationRecordLike {
  role: string;
  content: string;
  reasoning_content?: string;
  chat_completion_id?: string;
}

interface IProps {
  conversationId: string;
  initialRecords?: ConversationRecordLike[];
  guest?: boolean;
  onConversationCreated?: (id: string) => void;
  onSend?: (prompt: string) => void;
  height?: number | string;
  style?: React.CSSProperties;
}

type StreamChunk = {
  id?: string;
  choices?: Array<{
    delta?: { role?: string; content?: string; reasoning_content?: string };
    finish_reason?: string | null;
  }>;
};

const buildGuestHeaders = () => {
  const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
  return {
    Authorization: token,
    'X-Requested-With': 'XMLHttpRequest',
  };
};

/**
 * AnswerChatBot wraps TDesign React Chat with the Answer backend contract:
 * custom SSE framing, reasoning_content -> thinking blocks, guest mode and
 * conversation voting.
 */
const AnswerChatBot: React.FC<IProps> = (props) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const chatRef = useRef<HTMLElement & {
    setMessages?: (messages: ChatMessagesData[], mode?: 'replace' | 'prepend' | 'append') => void;
    registerMergeStrategy?: (type: string, handler: (chunk: SSEChunkData, existing?: AIMessageContent) => AIMessageContent) => void;
    chatMessageValue?: ChatMessagesData[];
  }>(null);

  const historyLoadedRef = useRef<string>('');
  const lastCompletionIdRef = useRef<string>('');
  // Effective conversation id: falls back to an internally generated one so
  // the first send of a brand-new conversation still carries a stable id even
  // if the parent page has not updated its route yet.
  const internalIdRef = useRef<string>('');
  const effectiveConversationId = () => {
    if (props.conversationId) {
      return props.conversationId;
    }
    if (!internalIdRef.current) {
      internalIdRef.current =
        'chatcmpl-' + Math.random().toString(36).slice(2) + Date.now().toString(36);
    }
    return internalIdRef.current;
  };

  const mapRecord = (r: ConversationRecordLike): ChatMessagesData => {
    const content: AIMessageContent[] = [];
    if (r.reasoning_content) {
      content.push({
        type: 'thinking',
        data: { text: r.reasoning_content, title: t('thoughts') || 'Thoughts' },
        status: 'complete',
      } as AIMessageContent);
    }
    content.push({ type: 'markdown', data: r.content || '', status: 'complete' });
    return {
      id: r.chat_completion_id || String(Math.random()),
      role: r.role === 'user' ? 'user' : 'assistant',
      status: 'complete',
      content,
    } as ChatMessagesData;
  };

  // Re-hydrate the chat when switching conversations.
  useEffect(() => {
    if (!chatRef.current?.setMessages || !props.initialRecords) {
      return;
    }
    const key = JSON.stringify(props.initialRecords.map((r) => r.chat_completion_id));
    if (historyLoadedRef.current === key) {
      return;
    }
    historyLoadedRef.current = key;
    chatRef.current.setMessages(props.initialRecords.map(mapRecord), 'replace');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.initialRecords]);

  // Merge streaming thinking chunks, whose default merge appends a new block
  // per delta instead of growing the text.
  useEffect(() => {
    chatRef.current?.registerMergeStrategy?.('thinking', (chunk, existing) => {
      const incoming = (chunk.data as { text?: string })?.text || '';
      const prev = (existing?.data as { text?: string }) || {};
      return {
        ...(existing || {}),
        type: 'thinking',
        data: { ...prev, text: (prev.text || '') + incoming },
      } as AIMessageContent;
    });
  }, []);

  const chatServiceConfig = useMemo(() => {
    const config: ChatServiceConfig = {
      endpoint: '/answer/api/v1/chat/completions',
      stream: true,
      onRequest: (params) => {
        const history = (chatRef.current?.chatMessageValue || []) as ChatMessagesData[];
        const historyMessages = history
          .filter((m) => m.role === 'user' || m.role === 'assistant')
          .map((m) => ({
            role: m.role,
            content: (m.content || [])
              .filter((c) => c.type === 'text' || c.type === 'markdown')
              .map((c) => (c.data as string) || '')
              .join(''),
          }))
          .filter((m) => m.content);
        const messages = props.guest
          ? [...historyMessages, { role: 'user', content: params.prompt }]
          : [{ role: 'user', content: params.prompt }];
        const cid = effectiveConversationId();
        if (!props.conversationId && props.onConversationCreated) {
          props.onConversationCreated(cid);
        }
        return {
          headers: buildGuestHeaders(),
          body: JSON.stringify({
            conversation_id: cid,
            messages,
          }),
        };
      },
      onMessage: (chunk: SSEChunkData): AIMessageContent | null => {
        const res = (chunk.data || {}) as StreamChunk;
        if (res.id) {
          lastCompletionIdRef.current = res.id;
        }
        const delta = res?.choices?.[0]?.delta;
        if (!delta) {
          return null;
        }
        if (delta.reasoning_content) {
          return {
            type: 'thinking',
            data: { text: delta.reasoning_content, title: t('thinking') || 'Thinking…' },
            status: 'streaming',
            id: `${res.id}-thinking`,
          } as AIMessageContent;
        }
        if (delta.content) {
          return {
            type: 'markdown',
            data: delta.content,
            status: 'streaming',
            strategy: 'merge',
            id: `${res.id}-answer`,
          } as AIMessageContent;
        }
        return null;
      },
      onComplete: () => {},
      onAbort: async () => {},
      onError: (err) => {
        console.error('AI chat error:', err);
      },
    };
    return config;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.conversationId, props.guest]);

  const messageProps = useMemo(() => {
    return {
      user: { variant: 'base', placement: 'right' },
      assistant: {
        placement: 'left',
        actions: props.guest ? ['copy', 'replay'] : ['copy', 'good', 'bad', 'replay'],
        handleActions: {
          good: () => {
            if (!props.guest && lastCompletionIdRef.current) {
              voteConversation({
                cancel: false,
                vote_type: 'helpful',
                chat_completion_id: lastCompletionIdRef.current,
              }).catch(() => {});
            }
          },
          bad: () => {
            if (!props.guest && lastCompletionIdRef.current) {
              voteConversation({
                cancel: false,
                vote_type: 'unhelpful',
                chat_completion_id: lastCompletionIdRef.current,
              }).catch(() => {});
            }
          },
        },
      },
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.guest]);

  return (
    <ChatBot
      ref={chatRef as never}
      style={{ height: props.height || '100%', ...props.style }}
      messageProps={messageProps as never}
      senderProps={{ placeholder: t('ask_placeholder') }}
      chatServiceConfig={chatServiceConfig as never}
      onSend={() => {
        // noop: keep engine flow; page-level hooks run through onRequest
      }}
    />
  );
};

export default AnswerChatBot;

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

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { ChatBot } from '@tdesign-react/chat';
import type {
  AIMessageContent,
  ChatMessagesData,
  ChatServiceConfig,
  SSEChunkData,
} from '@tdesign-react/chat';
import 'tdesign-react/dist/tdesign.css';
// Base + chat design tokens (--td-* and --td-chat-*). Without it every chat
// style that reads a var() falls back to its initial value (no borders,
// transparent backgrounds) and the whole chat renders unstyled.
import '@tdesign-react/chat/es/style/index.js';

import { voteConversation } from '@/services';
import { loggedUserInfoStore, aiControlStore } from '@/stores';
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
  height?: number | string;
  style?: React.CSSProperties;
  /** Children are forwarded to the underlying ChatBot for slot customization
   * (e.g. <div slot="sender-footer-prefix">…</div>). */
  children?: React.ReactNode;
}

type StreamChunk = {
  id?: string;
  choices?: Array<{
    delta?: { role?: string; content?: string; reasoning_content?: string };
    finish_reason?: string | null;
  }>;
};

/** Inline SVG avatar for the assistant (no extra asset bundling needed). */
const ASSISTANT_AVATAR = `data:image/svg+xml;utf8,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><rect width="48" height="48" rx="24" fill="#0052d9"/><g fill="#fff"><rect x="12" y="18" width="24" height="17" rx="5"/><rect x="22.6" y="9" width="2.8" height="9" rx="1.4"/><circle cx="24" cy="8" r="2.4"/></g><circle cx="18.6" cy="26" r="2.4" fill="#0052d9"/><circle cx="29.4" cy="26" r="2.4" fill="#0052d9"/><rect x="18" y="30.4" width="12" height="2.2" rx="1.1" fill="#0052d9"/></svg>',
)}`;

/** Welcome message id — excluded from request history and action bars. */
const WELCOME_ID = 'welcome';

const buildGuestHeaders = () => {
  const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
  return {
    Authorization: token,
    'X-Requested-With': 'XMLHttpRequest',
  };
};

/**
 * The completion id is carried on every streaming content block id as
 * `<chat_completion_id>-answer` / `<chat_completion_id>-thinking`, so each
 * message can be voted/replayed on its own instead of only the latest one.
 */
const extractCompletionId = (msg?: ChatMessagesData): string => {
  const content = ((msg?.content || []) as AIMessageContent[]) || [];
  const block = content.find(
    (c) =>
      typeof (c as { id?: string })?.id === 'string' &&
      /-(answer|thinking)$/.test((c as { id?: string }).id as string),
  );
  const blockId = (block as { id?: string } | undefined)?.id || '';
  return blockId.replace(/-(answer|thinking)$/, '');
};

/**
 * AnswerChatBot wraps TDesign React Chat with the Answer backend contract:
 * custom SSE framing, reasoning_content -> thinking blocks, guest mode,
 * per-message voting/replay and a suggestion-driven welcome screen.
 */
const AnswerChatBot: React.FC<IProps> = (props) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const userInfo = loggedUserInfoStore((s) => s.user);
  const chatRef = useRef<HTMLElement & {
    setMessages?: (messages: ChatMessagesData[], mode?: 'replace' | 'prepend' | 'append') => void;
    registerMergeStrategy?: (type: string, handler: (chunk: SSEChunkData, existing?: AIMessageContent) => AIMessageContent) => AIMessageContent;
    chatMessageValue?: ChatMessagesData[];
    sendSystemMessage?: (msg: string) => void;
    sendUserMessage?: (params: { prompt: string }) => Promise<void>;
    abortChat?: () => Promise<void>;
    regenerate?: (keepVersion?: boolean) => Promise<void>;
  }>(null);

  const voteMapRef = useRef<Record<string, 'helpful' | 'unhelpful' | ''>>({});
  const historyLoadedRef = useRef<string>('');
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
    const cid = r.chat_completion_id || '';
    if (r.reasoning_content) {
      content.push({
        type: 'thinking',
        id: cid ? `${cid}-thinking` : undefined,
        data: { text: r.reasoning_content, title: t('thoughts') || 'Thoughts' },
        status: 'complete',
      } as AIMessageContent);
    }
    content.push({
      type: 'markdown',
      id: cid ? `${cid}-answer` : undefined,
      data: r.content || '',
      status: 'complete',
    } as AIMessageContent);
    return {
      id: cid || String(Math.random()),
      role: r.role === 'user' ? 'user' : 'assistant',
      status: 'complete',
      content,
    } as ChatMessagesData;
  };

  // Welcome screen, rendered as ONE assistant message (like the official
  // TDesign demo): description text followed by suggestion chips. Chip data
  // comes from the admin "initial messages" lines; falls back to the
  // built-in i18n suggestions.
  const welcomeText = aiControlStore((s) => s.ai_welcome_text);
  const initialMessages = aiControlStore((s) => s.ai_initial_messages);
  const welcomeMessages = useMemo(() => {
    const chips = (initialMessages || '')
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => ({ title: line, prompt: line }));
    const suggestions = chips.length
      ? chips
      : [1, 2, 3, 4]
          .map((i) => {
            const title = t(`suggestion_${i}`);
            return title ? { title, prompt: title } : null;
          })
          .filter((s): s is { title: string; prompt: string } => !!s);
    const content: AIMessageContent[] = [
      { type: 'markdown', data: welcomeText || t('description') || '', status: 'complete' },
    ];
    if (suggestions.length) {
      content.push({
        type: 'suggestion',
        data: suggestions,
        status: 'complete',
      } as AIMessageContent);
    }
    return [
      {
        id: WELCOME_ID,
        role: 'assistant',
        status: 'complete',
        content,
      },
    ] as ChatMessagesData[];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [welcomeText, initialMessages]);

  // Remount counter for the ChatBot element (bumped on "new chat" resets).
  const [mountKey, setMountKey] = useState(0);

  // Re-hydrate the chat when switching conversations or after a remount. The
  // welcome message is only replaced when the conversation has records.
  useEffect(() => {
    if (!chatRef.current?.setMessages) {
      return;
    }
    if (!props.initialRecords?.length) {
      return;
    }
    const key = JSON.stringify(props.initialRecords.map((r) => r.chat_completion_id));
    if (historyLoadedRef.current === key) {
      return;
    }
    historyLoadedRef.current = key;
    chatRef.current.setMessages(props.initialRecords.map(mapRecord), 'replace');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.initialRecords, mountKey]);

  // Reset to the welcome screen when the parent starts a fresh conversation
  // ("new chat": id -> '' in the page, or a regenerated id in the widget).
  // A '' -> id transition is the first send assigning its id — keep the chat.
  //
  // The reset REMOUNTS the ChatBot (key bump) instead of calling
  // setMessages(): swapping the omi message store in place crashes the
  // action-bar re-render for conversations with completed messages.
  const prevCidRef = useRef(props.conversationId);
  useEffect(() => {
    const prev = prevCidRef.current;
    prevCidRef.current = props.conversationId;
    const isFreshStart =
      !props.conversationId || (prev && prev !== props.conversationId);
    if (!isFreshStart) {
      return;
    }
    internalIdRef.current = '';
    voteMapRef.current = {};
    historyLoadedRef.current = '';
    chatRef.current?.abortChat?.()?.catch?.(() => {});
    setMountKey((k) => k + 1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.conversationId]);

  // Never leave a stream running after the chat unmounts.
  useEffect(() => {
    return () => {
      chatRef.current?.abortChat?.()?.catch?.(() => {});
    };
  }, []);

  // Merge streaming thinking chunks, whose default merge appends a new block
  // per delta instead of growing the text. Re-registered on every remount
  // because the ChatBot element (and its engine) is brand new.
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
  }, [mountKey]);

  // Suggestion chips click -> send as a normal user message. The installed
  // web-components version has no directSend support: the click only reaches
  // this handler with { event, content }, so the send is driven here.
  const handleSuggestionClick = (data?: unknown) => {
    const content = (data as { content?: { prompt?: string; title?: string } })
      ?.content;
    const prompt = content?.prompt || content?.title || '';
    if (prompt) {
      chatRef.current?.sendUserMessage?.({ prompt })?.catch?.(() => {});
    }
  };

  const handleVote = (msg: ChatMessagesData, voteType: 'helpful' | 'unhelpful') => {
    if (props.guest) {
      return;
    }
    const cid = extractCompletionId(msg);
    if (!cid) {
      return;
    }
    const cancel = voteMapRef.current[cid] === voteType;
    voteConversation({
      cancel,
      vote_type: voteType,
      chat_completion_id: cid,
    })
      .then(() => {
        voteMapRef.current = {
          ...voteMapRef.current,
          [cid]: cancel ? '' : voteType,
        };
      })
      .catch(() => {});
  };

  const handleReplay = (msg: ChatMessagesData) => {
    const list = (chatRef.current?.chatMessageValue || []) as ChatMessagesData[];
    if (!list.length) {
      return;
    }
    const last = list[list.length - 1];
    const isLastAssistant =
      last?.id === msg?.id || extractCompletionId(last) === extractCompletionId(msg);
    if (isLastAssistant && last?.role === 'assistant') {
      chatRef.current?.regenerate?.(false)?.catch?.(() => {});
    }
  };

  const chatServiceConfig = useMemo(() => {
    const config: ChatServiceConfig = {
      endpoint: '/answer/api/v1/chat/completions',
      stream: true,
      onRequest: (params) => {
        const history = ((chatRef.current?.chatMessageValue || []) as ChatMessagesData[])
          .filter(
            (m) =>
              !(typeof m.id === 'string' && m.id.startsWith(WELCOME_ID)) &&
              (m.role === 'user' || m.role === 'assistant'),
          )
          .map((m) => ({
            role: m.role,
            content: (((m.content || []) as AIMessageContent[]) || [])
              .filter((c) => c.type === 'text' || c.type === 'markdown')
              .map((c) => (c.data as string) || '')
              .join(''),
          }))
          .filter((m) => m.content);
        const prompt = params.prompt || '';
        let messages = [...history];
        if (props.guest) {
          // Replays resubmit the last question which is already in history —
          // avoid appending it twice.
          const lastUser = [...history].reverse().find((m) => m.role === 'user');
          const isReplay = lastUser?.content === prompt;
          if (isReplay && messages.length && messages[messages.length - 1].role === 'assistant') {
            messages = messages.slice(0, -1);
          }
          if (!isReplay) {
            messages = [...messages, { role: 'user', content: prompt }];
          }
        } else {
          messages = [{ role: 'user', content: prompt }];
        }
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
      onError: () => {
        chatRef.current?.sendSystemMessage?.(
          t('request_failed') || 'Request failed, please try again later.',
        );
      },
    };
    return config;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.conversationId, props.guest]);

  // Per-message config: avatars/names per role, per-message votes and replay
  // (the completion id is recovered from the message's own content blocks).
  const messageProps = useMemo(() => {
    const userAvatar = props.guest ? '' : userInfo?.avatar || '';
    const userName = props.guest ? '' : userInfo?.display_name || userInfo?.username || '';

    return (msg: ChatMessagesData) => {
      if (typeof msg?.id === 'string' && msg.id.startsWith(WELCOME_ID)) {
        return {
          actions: false as const,
          avatar: ASSISTANT_AVATAR,
          name: t('ai_name') || 'AI',
          // The suggestion content block delivers its clicks through the
          // per-message handleActions — without this the chips do nothing.
          handleActions: { suggestion: handleSuggestionClick },
        };
      }
      if (msg?.role === 'user') {
        return {
          actions: false as const,
          variant: 'base' as const,
          placement: 'right' as const,
          avatar: userAvatar,
          name: userName,
        };
      }
      return {
        placement: 'left' as const,
        avatar: ASSISTANT_AVATAR,
        name: t('ai_name') || 'AI',
        actions: props.guest
          ? (['copy', 'replay'] as const)
          : (['copy', 'good', 'bad', 'replay'] as const),
        chatContentProps: {
          thinking: { maxHeight: 220, layout: 'border' as const },
          suggestion: { directSend: true },
          markdown: {
            options: { themeSettings: { codeBlockTheme: 'light' as const } },
          },
        },
        handleActions: {
          good: () => handleVote(msg, 'helpful'),
          bad: () => handleVote(msg, 'unhelpful'),
          replay: () => handleReplay(msg),
          suggestion: handleSuggestionClick,
        },
      };
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.guest, userInfo?.avatar, userInfo?.display_name, userInfo?.username]);

  return (
    <ChatBot
      key={mountKey}
      ref={chatRef as never}
      style={{ height: props.height || '100%', ...props.style }}
      defaultMessages={welcomeMessages as never}
      messageProps={messageProps as never}
      senderProps={{ placeholder: t('ask_placeholder') }}
      chatServiceConfig={chatServiceConfig as never}>
      {props.children}
    </ChatBot>
  );
};

export default AnswerChatBot;

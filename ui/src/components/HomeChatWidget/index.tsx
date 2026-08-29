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

import React, { Suspense, lazy, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { loggedUserInfoStore, aiControlStore } from '@/stores';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import { Storage } from '@/utils';

const AnswerChatBot = lazy(() => import('@/components/AnswerChatBot'));

/**
 * Floating AI chat entry on the homepage. Rendered only when the admin turns
 * on the homepage chat; logged-out visitors need the guest switch as well.
 * The panel mounts the same AnswerChatBot used by the assistant page, in a
 * lightweight single-conversation mode.
 */
const HomeChatWidget: React.FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const { t: tPage } = useTranslation('translation', { keyPrefix: 'page_title' });
  const { ai_enabled, ai_home_chat_enabled, ai_home_chat_guest_enabled } =
    aiControlStore((s) => s);
  const { user } = loggedUserInfoStore((s) => s);
  const [open, setOpen] = useState(false);
  const [conversationId, setConversationId] = useState<string>('');

  if (!ai_enabled || !ai_home_chat_enabled) {
    return null;
  }
  const isLogged = !!user?.id && !!Storage.get(LOGGED_TOKEN_STORAGE_KEY);
  if (!isLogged && !ai_home_chat_guest_enabled) {
    return null;
  }

  const newConversation = () => {
    setConversationId(
      'chatcmpl-' + Math.random().toString(36).slice(2) + Date.now().toString(36),
    );
  };

  return (
    <div
      style={{
        position: 'fixed',
        right: 24,
        bottom: 24,
        zIndex: 1050,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-end',
        gap: 12,
      }}>
      {open && (
        <div
          style={{
            width: 420,
            maxWidth: 'calc(100vw - 48px)',
            height: 560,
            maxHeight: 'calc(100vh - 140px)',
            background: '#fff',
            borderRadius: 12,
            boxShadow: '0 8px 30px rgba(0,0,0,0.18)',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}>
          <div
            style={{
              padding: '10px 14px',
              borderBottom: '1px solid #e7e7e7',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flex: '0 0 auto',
            }}>
            <strong>{tPage('ai_assistant')}</strong>
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                type="button"
                title={t('new')}
                onClick={newConversation}
                style={{
                  border: 'none',
                  background: 'none',
                  cursor: 'pointer',
                  fontSize: 15,
                }}>
                ⟳
              </button>
              <button
                type="button"
                aria-label="close"
                onClick={() => setOpen(false)}
                style={{
                  border: 'none',
                  background: 'none',
                  cursor: 'pointer',
                  fontSize: 16,
                }}>
                ✕
              </button>
            </div>
          </div>
          <div style={{ flex: '1 1 auto', minHeight: 0 }}>
            <Suspense
              fallback={
                <div style={{ padding: 20, color: '#888' }}>…</div>
              }>
              <AnswerChatBot
                conversationId={conversationId}
                onConversationCreated={(id) => {
                  if (!conversationId) {
                    setConversationId(id);
                  }
                }}
                guest={!isLogged}
                height="100%">
                <div
                  slot="sender-footer-prefix"
                  className="small text-secondary ps-2 text-truncate">
                  {t('ai_disclaimer')}
                </div>
              </AnswerChatBot>
            </Suspense>
          </div>
        </div>
      )}
      <button
        type="button"
        onClick={() => {
          if (!open && !conversationId) {
            newConversation();
          }
          setOpen(!open);
        }}
        title={tPage('ai_assistant')}
        style={{
          width: 52,
          height: 52,
          borderRadius: '50%',
          border: 'none',
          background: '#0052d9',
          color: '#fff',
          fontSize: 18,
          fontWeight: 600,
          cursor: 'pointer',
          boxShadow: '0 4px 14px rgba(0,82,217,0.4)',
        }}>
        AI
      </button>
    </div>
  );
};

export default HomeChatWidget;

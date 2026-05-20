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
import { Button, Modal } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { BubbleAi, BubbleUser } from '@/components';
import { useQueryAdminConversationDetail } from '@/services';

interface IProps {
  visible: boolean;
  id: string;
  onClose?: () => void;
}

const Index: FC<IProps> = ({ visible, id, onClose }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.conversations',
  });

  const { data: conversationDetail } = useQueryAdminConversationDetail(id);

  const handleClose = () => {
    onClose?.();
  };
  return (
    <Modal show={visible} size="lg" centered onHide={handleClose}>
      <Modal.Header closeButton>
        <div style={{ maxWidth: '85%' }} className="text-truncate">
          {conversationDetail?.topic}
        </div>
      </Modal.Header>
      <Modal.Body className="overflow-y-auto" style={{ maxHeight: '70vh' }}>
        {conversationDetail?.records.map((item, index) => {
          const isLastMessage =
            index === Number(conversationDetail?.records.length) - 1;
          return (
            <div
              key={`${item.chat_completion_id}-${item.role}`}
              className={`${isLastMessage ? '' : 'mb-4'}`}>
              {item.role === 'user' ? (
                <BubbleUser content={item.content} />
              ) : (
                <BubbleAi
                  canType={false}
                  chatId={item.chat_completion_id}
                  isLast={false}
                  isCompleted
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
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={handleClose}>
          {t('close', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default memo(Index);

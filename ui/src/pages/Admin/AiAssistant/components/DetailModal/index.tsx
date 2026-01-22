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

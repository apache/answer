import { FC, memo } from 'react';
import { Card, ListGroup } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

interface ConversationListItem {
  conversation_id: string;
  topic: string;
}

interface IProps {
  data: {
    count: number;
    list: ConversationListItem[];
  };
  loadMore: (e: React.MouseEvent<HTMLAnchorElement>) => void;
}

const Index: FC<IProps> = ({ data, loadMore }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });

  if (Number(data?.list.length) <= 0) return null;
  return (
    <Card>
      <Card.Header>
        <span>{t('recent_conversations')}</span>
      </Card.Header>
      <ListGroup variant="flush">
        {data?.list.map((item) => {
          return (
            <ListGroup.Item
              as={Link}
              action
              key={item.conversation_id}
              to={`/ai-assistant/${item.conversation_id}`}
              className="text-truncate">
              {item.topic}
            </ListGroup.Item>
          );
        })}
        {Number(data?.count) > data?.list.length && (
          <ListGroup.Item action onClick={loadMore} className="link-primary">
            {t('show_more')}
          </ListGroup.Item>
        )}
      </ListGroup>
    </Card>
  );
};

export default memo(Index);

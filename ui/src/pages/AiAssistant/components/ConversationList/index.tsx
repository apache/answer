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

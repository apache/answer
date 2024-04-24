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

import { ListGroup } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';
import isEmpty from 'lodash/isEmpty';

import { FormatTime, Empty } from '@/components';

const Inbox = ({ data, handleReadNotification }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'notifications' });
  if (!data) {
    return null;
  }
  if (isEmpty(data)) {
    return <Empty />;
  }
  return (
    <ListGroup className="rounded-0">
      {data.map((item) => {
        const { comment, question, answer } =
          item?.object_info?.object_map || {};
        let url = '';
        switch (item.object_info.object_type) {
          case 'question':
            url = `/questions/${item.object_info.object_id}`;
            break;
          case 'answer':
            url = `/questions/${question}/${item.object_info.object_id}`;
            break;
          case 'comment':
            url = `/questions/${question}/${answer}?commentId=${comment}`;
            break;
          default:
            url = '';
        }
        return (
          <ListGroup.Item
            key={item.id}
            className={classNames(
              'py-3 border-start-0 border-end-0',
              !item.is_read && 'warning',
            )}>
            <div>
              {item.user_info && item.user_info.status !== 'deleted' ? (
                <Link to={`/users/${item.user_info.username}`}>
                  {item.user_info.display_name}{' '}
                </Link>
              ) : (
                // someone for anonymous user display
                <span>{item.user_info?.display_name || t('someone')} </span>
              )}
              {item.notification_action}{' '}
              <Link to={url} onClick={() => handleReadNotification(item.id)}>
                {item.object_info.title}
              </Link>
            </div>
            <div className="text-secondary small">
              <FormatTime time={item.update_time} />
            </div>
          </ListGroup.Item>
        );
      })}
    </ListGroup>
  );
};

export default Inbox;

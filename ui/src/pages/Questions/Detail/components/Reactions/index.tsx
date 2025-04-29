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

import { FC, memo, useEffect, useState } from 'react';
import { Button, OverlayTrigger, Popover, Tooltip } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import { Icon } from '@/components';
import { queryReactions, updateReaction } from '@/services';
import { tryNormalLogged } from '@/utils/guard';
import { isDarkTheme } from '@/utils/common';
import { ReactionItem } from '@/common/interface';

interface Props {
  objectId: string;
  showAddCommentBtn?: boolean;
  handleClickComment: () => void;
}

const emojiMap = [
  {
    name: 'heart',
    icon: 'heart-fill',
    className: 'text-danger',
  },
  {
    name: 'smile',
    icon: 'emoji-laughing-fill',
    className: 'text-warning',
  },
  {
    name: 'frown',
    icon: 'emoji-frown-fill',
    className: 'text-warning',
  },
];

const Index: FC<Props> = ({
  objectId,
  showAddCommentBtn,
  handleClickComment,
}) => {
  const [reactions, setReactions] = useState<ReactionItem[]>();
  const [reactIsActive, setReactIsActive] = useState<boolean>(false);
  const { t } = useTranslation('translation');
  const darkMode = isDarkTheme();

  useEffect(() => {
    queryReactions(objectId).then((res) => {
      setReactions(res?.reaction_summary);
    });
  }, []);

  const handleSubmit = (params: { object_id: string; emoji: string }) => {
    if (!tryNormalLogged(true)) {
      setReactIsActive(false);
      return;
    }
    updateReaction({
      ...params,
      reaction: reactions?.find((v) => v.emoji === params.emoji)?.is_active
        ? 'deactivate'
        : 'activate',
    }).then((res) => {
      setReactions(res.reaction_summary);
      setReactIsActive(false);
    });
  };

  const renderPopover = (props) => (
    <Popover id="reaction-button-tooltip" {...props}>
      <Popover.Body className="d-block d-md-flex flex-wrap p-1">
        {emojiMap.map((d, index) => (
          <Button
            aria-label={
              reactions?.find((v) => v.emoji === d.name)?.is_active
                ? t('reaction.undo_emoji', { emoji: d.name })
                : t(`reaction.${d.name}`)
            }
            key={d.icon}
            variant={darkMode ? 'dark' : 'light'}
            active={reactions?.find((v) => v.emoji === d.name)?.is_active}
            className={`${index !== 0 ? 'ms-1' : ''}`}
            size="sm"
            onClick={() =>
              handleSubmit({ object_id: objectId, emoji: d.name })
            }>
            <Icon name={d.icon} className={d.className} />
          </Button>
        ))}
      </Popover.Body>
    </Popover>
  );

  return (
    <div
      className={classNames('d-flex flex-wrap', {
        'mb-3': !showAddCommentBtn,
      })}>
      {showAddCommentBtn && (
        <Button
          className="rounded-pill me-2 link-secondary btn-reaction"
          variant={darkMode ? 'dark' : 'light'}
          size="sm"
          onClick={handleClickComment}>
          <Icon name="chat-text-fill" />
          <span className="ms-1">{t('comment.btn_add_comment')}</span>
        </Button>
      )}

      <OverlayTrigger
        trigger="click"
        placement="top"
        overlay={renderPopover}
        show={reactIsActive}
        onToggle={(show) => setReactIsActive(show)}>
        <Button
          size="sm"
          aria-label={t('reaction.btn_label')}
          aria-haspopup="true"
          active={reactIsActive}
          className="smile-btn rounded-pill link-secondary btn-reaction"
          variant={darkMode ? 'dark' : 'light'}>
          <Icon name="emoji-smile-fill" />
          <span className="ms-1">+</span>
        </Button>
      </OverlayTrigger>

      {reactions?.map((data) => {
        if (!data.emoji || data?.count <= 0) {
          return null;
        }
        return (
          <OverlayTrigger
            key={data.emoji}
            placement="top"
            overlay={
              <Tooltip>
                <div className="text-start">
                  <b>{t(`reaction.${data.emoji}`)}</b> <br /> {data.tooltip}
                </div>
              </Tooltip>
            }>
            <Button
              className="rounded-pill ms-2 link-secondary d-flex align-items-center btn-reaction"
              aria-label={
                data?.is_active
                  ? t('reaction.unreact_emoji', { emoji: data.emoji })
                  : t('reaction.react_emoji', { emoji: data.emoji })
              }
              aria-pressed="true"
              variant={darkMode ? 'dark' : 'light'}
              active={data.is_active}
              size="sm"
              onClick={() =>
                handleSubmit({ object_id: objectId, emoji: data.emoji })
              }>
              <Icon
                name={String(emojiMap.find((v) => v.name === data.emoji)?.icon)}
                className={
                  emojiMap.find((v) => v.name === data.emoji)?.className
                }
              />
              <span className="ms-1 lh-1">{data.count}</span>
            </Button>
          </OverlayTrigger>
        );
      })}
    </div>
  );
};

export default memo(Index);

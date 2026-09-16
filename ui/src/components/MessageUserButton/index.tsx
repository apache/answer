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

import { FC, useEffect, useState } from 'react';
import { Button, Form, Modal } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import useSWR from 'swr';
import qs from 'qs';

import { Icon } from '@/components';
import { useToast } from '@/hooks';
import { sendAdminMessage } from '@/services';
import request from '@/utils/request';
import { loggedUserInfoStore } from '@/stores';

interface Props {
  /** fixed recipient; when omitted the modal asks for a user */
  username?: string;
  /** the question / answer / comment the message is about (shown as context, stored with the message) */
  objectId?: string;
  objectTitle?: string;
  variant?: 'icon' | 'button';
  className?: string;
  onSent?: () => void;
}

/**
 * "Write a message" for admins and moderators: the recipient gets it in Notifications → Messages.
 * Renders nothing for regular users.
 */
const MessageUserButton: FC<Props> = ({
  username,
  objectId,
  objectTitle,
  variant = 'button',
  className = '',
  onSent,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin_message' });
  const roleId = loggedUserInfoStore((state) => state.user?.role_id);
  const Toast = useToast();
  const [show, setShow] = useState(false);
  const [user, setUser] = useState(username || '');
  const [query, setQuery] = useState('');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);
  // user suggestions come from the admin user list, so only admins get them; moderators type the username
  const { data: users } = useSWR<{ list: any[] }>(
    show && !username && roleId === 2 && query.length >= 2
      ? `/answer/admin/api/users/page?${qs.stringify({ page: 1, page_size: 8, query })}`
      : null,
    request.instance.get,
  );

  useEffect(() => {
    setUser(username || '');
  }, [username, show]);

  if (roleId !== 2 && roleId !== 3) {
    return null;
  }

  const handleSubmit = () => {
    const to = (user || query).trim();
    if (!to || !title.trim() || !body.trim()) {
      return;
    }
    setSubmitting(true);
    sendAdminMessage({
      username: to,
      title: title.trim(),
      body: body.trim(),
      object_id: objectId || undefined,
    })
      .then(() => {
        Toast.onShow({
          msg: t('success', { username: to }),
          variant: 'success',
        });
        setShow(false);
        setTitle('');
        setBody('');
        onSent?.();
      })
      .catch((err) => {
        Toast.onShow({ msg: err?.msg || t('failed'), variant: 'danger' });
      })
      .finally(() => setSubmitting(false));
  };

  const open = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setShow(true);
  };

  return (
    <>
      {variant === 'icon' ? (
        <Button
          variant="link"
          size="sm"
          className={`p-0 align-baseline text-secondary ${className}`}
          title={t('btn')}
          aria-label={t('btn')}
          onClick={open}>
          <Icon name="envelope" />
        </Button>
      ) : (
        <Button
          variant="outline-primary"
          size="sm"
          className={className}
          onClick={open}>
          <Icon name="envelope" className="me-1" />
          {t('btn')}
        </Button>
      )}
      <Modal
        show={show}
        onHide={() => setShow(false)}
        onClick={(e) => e.stopPropagation()}>
        <Modal.Header closeButton>
          <Modal.Title as="h5">{t('title')}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3" controlId="admin_message_user">
              <Form.Label>{t('to')}</Form.Label>
              {username ? (
                <Form.Control value={username} disabled />
              ) : (
                <>
                  <Form.Control
                    value={user || query}
                    placeholder={t('to_placeholder')}
                    autoComplete="off"
                    onChange={(e) => {
                      setQuery(e.target.value);
                      setUser('');
                    }}
                  />
                  {!user && users?.list?.length ? (
                    <div className="list-group mt-1">
                      {users.list.map((u) => (
                        <button
                          type="button"
                          key={u.user_id}
                          className="list-group-item list-group-item-action py-1"
                          onClick={() => {
                            setUser(u.username);
                            setQuery('');
                          }}>
                          {u.display_name}{' '}
                          <span className="text-secondary small">
                            @{u.username}
                          </span>
                        </button>
                      ))}
                    </div>
                  ) : null}
                </>
              )}
            </Form.Group>
            {objectTitle ? (
              <div className="mb-3 small text-secondary">
                {t('context')}: <span className="text-body">{objectTitle}</span>
              </div>
            ) : null}
            <Form.Group className="mb-3" controlId="admin_message_title">
              <Form.Label>{t('subject')}</Form.Label>
              <Form.Control
                value={title}
                maxLength={200}
                placeholder={t('subject_placeholder')}
                onChange={(e) => setTitle(e.target.value)}
              />
            </Form.Group>
            <Form.Group controlId="admin_message_body">
              <Form.Label>{t('body')}</Form.Label>
              <Form.Control
                as="textarea"
                rows={6}
                maxLength={4000}
                value={body}
                placeholder={t('body_placeholder')}
                onChange={(e) => setBody(e.target.value)}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => setShow(false)}>
            {t('cancel', { keyPrefix: 'btns' })}
          </Button>
          <Button
            variant="primary"
            disabled={
              !(user || query).trim() ||
              !title.trim() ||
              !body.trim() ||
              submitting
            }
            onClick={handleSubmit}>
            {t('send')}
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
};

export default MessageUserButton;

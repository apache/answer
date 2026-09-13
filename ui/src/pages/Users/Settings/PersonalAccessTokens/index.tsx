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

import { useMemo, useState } from 'react';
import { Badge, Button, Form, Modal, Table } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';

import type * as Type from '@/common/interface';
import {
  createPersonalAccessToken,
  reauthenticate,
  revokePersonalAccessToken,
  usePersonalAccessTokens,
} from '@/services';
import { loggedUserInfoStore, siteSecurityStore } from '@/stores';
import { useToast } from '@/hooks';

const scopes: Type.PersonalAccessTokenScope[] = [
  'question.read',
  'question.create',
  'answer.read',
  'answer.create',
  'vote.write',
];

const PersonalAccessTokens = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.personal_access_tokens',
  });
  const toast = useToast();
  const security = siteSecurityStore();
  const user = loggedUserInfoStore((state) => state.user);
  const { data = [], mutate } = usePersonalAccessTokens();
  const [showCreate, setShowCreate] = useState(false);
  const [showSecret, setShowSecret] = useState(false);
  const [showReauthenticate, setShowReauthenticate] = useState(false);
  const [name, setName] = useState('');
  const [selectedScopes, setSelectedScopes] = useState<
    Type.PersonalAccessTokenScope[]
  >([]);
  const [expirationDays, setExpirationDays] = useState(30);
  const [createdToken, setCreatedToken] = useState('');
  const [password, setPassword] = useState('');
  const [pendingCreate, setPendingCreate] =
    useState<Type.CreatePersonalAccessTokenParams>();

  const canCreate = useMemo(
    () => name.trim().length > 0 && selectedScopes.length > 0,
    [name, selectedScopes],
  );

  const resetCreate = () => {
    setName('');
    setSelectedScopes([]);
    setExpirationDays(30);
    setShowCreate(false);
  };

  const finishCreate = (params: Type.CreatePersonalAccessTokenParams) =>
    createPersonalAccessToken(params)
      .then((result) => {
        setCreatedToken(result.token);
        setShowSecret(true);
        setPendingCreate(undefined);
        resetCreate();
        mutate();
      })
      .catch((error) => {
        if (
          error?.reason ===
          'error.personal_access_token.reauthentication_required'
        ) {
          setPendingCreate(params);
          setShowReauthenticate(true);
          return;
        }
        toast.onShow({
          msg: error?.msg || t('create_failed'),
          variant: 'danger',
        });
      });

  const createToken = () => {
    const params: Type.CreatePersonalAccessTokenParams = {
      name: name.trim(),
      scopes: selectedScopes,
      expires_at: dayjs().add(expirationDays, 'day').unix(),
    };
    finishCreate(params);
  };

  const confirmReauthentication = () => {
    if (!pendingCreate || !password) {
      return;
    }
    reauthenticate(password)
      .then(() => {
        setPassword('');
        setShowReauthenticate(false);
        finishCreate(pendingCreate);
      })
      .catch((error) => {
        toast.onShow({
          msg: error?.msg || t('reauthenticate_failed'),
          variant: 'danger',
        });
      });
  };

  const toggleScope = (scope: Type.PersonalAccessTokenScope) => {
    setSelectedScopes((current) =>
      current.includes(scope)
        ? current.filter((item) => item !== scope)
        : [...current, scope],
    );
  };

  const revoke = (id: number) => {
    if (!window.confirm(t('revoke_confirm'))) {
      return;
    }
    revokePersonalAccessToken(id).then(() => mutate());
  };

  return (
    <div>
      <h3 className="mb-3">{t('heading')}</h3>
      <p className="text-secondary">{t('description')}</p>
      {!security.personal_access_tokens_enabled && (
        <div className="alert alert-warning">{t('disabled')}</div>
      )}
      {security.personal_access_tokens_enabled && (
        <Button size="sm" className="mb-3" onClick={() => setShowCreate(true)}>
          {t('create')}
        </Button>
      )}
      <Table responsive>
        <thead>
          <tr>
            <th>{t('name')}</th>
            <th>{t('token')}</th>
            <th>{t('scopes')}</th>
            <th>{t('expires')}</th>
            <th>{t('status')}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {data.map((item) => (
            <tr key={item.id}>
              <td>{item.name}</td>
              <td>answer_pat_••••{item.token_suffix}</td>
              <td>
                {item.scopes.map((scope) => t(`scope.${scope}`)).join(', ')}
              </td>
              <td>{dayjs.unix(item.expires_at).format('YYYY-MM-DD')}</td>
              <td>
                <Badge bg={item.status === 'active' ? 'success' : 'secondary'}>
                  {t(`token_status.${item.status}`)}
                </Badge>
              </td>
              <td className="text-end">
                {item.status === 'active' && (
                  <Button
                    variant="link"
                    size="sm"
                    onClick={() => revoke(item.id)}>
                    {t('revoke')}
                  </Button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </Table>

      <Modal show={showCreate} onHide={resetCreate}>
        <Modal.Header closeButton>
          <Modal.Title>{t('create')}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form.Group className="mb-3">
            <Form.Label>{t('name')}</Form.Label>
            <Form.Control
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>{t('scopes')}</Form.Label>
            {scopes.map((scope) => (
              <Form.Check
                key={scope}
                type="checkbox"
                label={t(`scope.${scope}`)}
                checked={selectedScopes.includes(scope)}
                onChange={() => toggleScope(scope)}
              />
            ))}
          </Form.Group>
          <Form.Group>
            <Form.Label>{t('expiration')}</Form.Label>
            <Form.Select
              value={expirationDays}
              onChange={(event) =>
                setExpirationDays(Number(event.target.value))
              }>
              {[7, 30, 90, 365].map((days) => (
                <option key={days} value={days}>
                  {t('days', { count: days })}
                </option>
              ))}
            </Form.Select>
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={resetCreate}>
            {t('cancel', { keyPrefix: 'btns' })}
          </Button>
          <Button disabled={!canCreate} onClick={createToken}>
            {t('create')}
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal show={showSecret} onHide={() => setShowSecret(false)}>
        <Modal.Header closeButton>
          <Modal.Title>{t('created_title')}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>{t('created_warning')}</p>
          <Form.Control readOnly value={createdToken} />
        </Modal.Body>
        <Modal.Footer>
          <Button onClick={() => navigator.clipboard.writeText(createdToken)}>
            {t('copy')}
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal
        show={showReauthenticate}
        onHide={() => setShowReauthenticate(false)}>
        <Modal.Header closeButton>
          <Modal.Title>{t('reauthenticate')}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {user.have_password ? (
            <Form.Control
              type="password"
              value={password}
              placeholder={t('password')}
              onChange={(event) => setPassword(event.target.value)}
            />
          ) : (
            <p>{t('reauthenticate_external')}</p>
          )}
        </Modal.Body>
        {user.have_password && (
          <Modal.Footer>
            <Button disabled={!password} onClick={confirmReauthentication}>
              {t('continue')}
            </Button>
          </Modal.Footer>
        )}
      </Modal>
    </div>
  );
};

export default PersonalAccessTokens;

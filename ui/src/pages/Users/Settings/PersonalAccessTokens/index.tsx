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

import { useMemo, useRef, useState } from 'react';
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
import { useCaptchaPlugin } from '@/utils/pluginKit';

const scopeGroups: Array<{
  topic: 'question' | 'answer' | 'vote';
  scopes: Type.PersonalAccessTokenScope[];
}> = [
  {
    topic: 'question',
    scopes: ['question.read', 'question.create'],
  },
  {
    topic: 'answer',
    scopes: ['answer.read', 'answer.create'],
  },
  {
    topic: 'vote',
    scopes: ['vote.write'],
  },
];

const isRevocable = (status: string) =>
  status === 'active' || status === 'temporarily_unavailable';

const PersonalAccessTokens = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.personal_access_tokens',
  });
  const toast = useToast();
  const security = siteSecurityStore();
  const user = loggedUserInfoStore((state) => state.user);
  const reauthenticationCaptcha = useCaptchaPlugin('edit_userinfo');
  const {
    data,
    mutate,
    isLoading,
    error: loadError,
  } = usePersonalAccessTokens();
  const [showCreate, setShowCreate] = useState(false);
  const [showSecret, setShowSecret] = useState(false);
  const [showReauthenticate, setShowReauthenticate] = useState(false);
  const [name, setName] = useState('');
  const [selectedScopes, setSelectedScopes] = useState<
    Type.PersonalAccessTokenScope[]
  >([]);
  const [expirationDays, setExpirationDays] = useState('30');
  const [customExpiration, setCustomExpiration] = useState('');
  const [showInactive, setShowInactive] = useState(false);
  const [createdToken, setCreatedToken] = useState('');
  const [password, setPassword] = useState('');
  const pendingCreate = useRef<Type.CreatePersonalAccessTokenParams>();
  const submitting = useRef(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const refreshTokens = () => {
    void mutate().catch(() => {
      // SWR exposes refresh failures to the list's retryable error state.
    });
  };

  const canCreate = useMemo(
    () =>
      name.trim().length > 0 &&
      selectedScopes.length > 0 &&
      (expirationDays !== 'custom' || Boolean(customExpiration)),
    [customExpiration, expirationDays, name, selectedScopes],
  );

  const visibleTokens = useMemo(
    () =>
      showInactive
        ? (data ?? [])
        : (data ?? []).filter((item) => isRevocable(item.status)),
    [data, showInactive],
  );

  const resetCreate = () => {
    setName('');
    setSelectedScopes([]);
    setExpirationDays('30');
    setCustomExpiration('');
    setShowCreate(false);
  };

  const finishCreate = (params: Type.CreatePersonalAccessTokenParams) =>
    createPersonalAccessToken(params)
      .then((result) => {
        setCreatedToken(result.token);
        setShowSecret(true);
        pendingCreate.current = undefined;
        resetCreate();
        refreshTokens();
      })
      .catch((error) => {
        if (
          error?.reason ===
          'error.personal_access_token.reauthentication_required'
        ) {
          pendingCreate.current = params;
          setShowReauthenticate(true);
          return;
        }
        toast.onShow({
          msg: error?.msg || t('create_failed'),
          variant: 'danger',
        });
      });

  const dismissCreate = () => {
    if (submitting.current) return;
    pendingCreate.current = undefined;
    setPassword('');
    resetCreate();
  };

  const dismissReauthentication = () => {
    if (submitting.current) return;
    pendingCreate.current = undefined;
    setPassword('');
    setShowReauthenticate(false);
    void reauthenticationCaptcha?.close();
  };

  const createToken = async () => {
    if (submitting.current || !canCreate) return;
    submitting.current = true;
    setIsSubmitting(true);
    const params: Type.CreatePersonalAccessTokenParams = {
      name: name.trim(),
      scopes: selectedScopes,
      expires_at:
        expirationDays === 'custom'
          ? dayjs(customExpiration).startOf('day').unix()
          : dayjs().add(Number(expirationDays), 'day').unix(),
    };
    try {
      await finishCreate(params);
    } finally {
      submitting.current = false;
      setIsSubmitting(false);
    }
  };

  const confirmReauthentication = async () => {
    const createParams = pendingCreate.current;
    if (submitting.current || !createParams || !password) return;
    submitting.current = true;
    setIsSubmitting(true);
    const params: {
      password: string;
      captcha_id?: string;
      captcha_code?: string;
    } = { password };
    try {
      reauthenticationCaptcha?.resolveCaptchaReq(params);
      await reauthenticate(params);
      await reauthenticationCaptcha?.close();
      setPassword('');
      setShowReauthenticate(false);
      await finishCreate(createParams);
    } catch (failure: any) {
      if (failure?.isError) {
        reauthenticationCaptcha?.handleCaptchaError(failure.list);
      }
      toast.onShow({
        msg: failure?.msg || t('reauthenticate_failed'),
        variant: 'danger',
      });
    } finally {
      submitting.current = false;
      setIsSubmitting(false);
    }
  };

  const handleReauthentication = () => {
    const createParams = pendingCreate.current;
    if (submitting.current || !createParams || !password) return;
    const confirm = () => {
      if (pendingCreate.current === createParams) {
        void confirmReauthentication();
      }
    };
    if (!reauthenticationCaptcha) {
      confirm();
      return;
    }
    reauthenticationCaptcha.check(confirm);
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
    revokePersonalAccessToken(id).then(refreshTokens);
  };

  return (
    <div>
      <h3 className="mb-3">{t('heading')}</h3>
      <p className="text-secondary">{t('description_text')}</p>
      {!security.personal_access_tokens_enabled && (
        <div className="alert alert-warning">{t('disabled')}</div>
      )}
      {security.personal_access_tokens_enabled && (
        <Button
          size="sm"
          className="mb-3"
          disabled={isSubmitting || showSecret || showReauthenticate}
          onClick={() => setShowCreate(true)}>
          {t('create')}
        </Button>
      )}
      <Form.Check
        id="pat-show-inactive"
        className="mb-3"
        type="checkbox"
        label={t('show_inactive')}
        checked={showInactive}
        onChange={(event) => setShowInactive(event.target.checked)}
      />
      {loadError ? (
        <div role="alert" className="alert alert-danger">
          {t('load_failed')}
          <Button variant="link" onClick={refreshTokens}>
            {t('retry')}
          </Button>
        </div>
      ) : isLoading || data === undefined ? (
        <div role="status">{t('loading')}</div>
      ) : (
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
            {visibleTokens.length === 0 && (
              <tr>
                <td colSpan={6}>{t('empty')}</td>
              </tr>
            )}
            {visibleTokens.map((item) => (
              <tr key={item.id}>
                <td>{item.name}</td>
                <td>answer_pat_••••{item.token_suffix}</td>
                <td>
                  {item.scopes.map((scope) => t(`scope.${scope}`)).join(', ')}
                </td>
                <td>{dayjs.unix(item.expires_at).format('YYYY-MM-DD')}</td>
                <td>
                  <Badge
                    bg={item.status === 'active' ? 'success' : 'secondary'}>
                    {t(`token_status.${item.status}`)}
                  </Badge>
                </td>
                <td className="text-end">
                  {isRevocable(item.status) && (
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
      )}

      <Modal show={showCreate} onHide={dismissCreate}>
        <Modal.Header closeButton={!isSubmitting}>
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
            {scopeGroups.map((group) => (
              <fieldset key={group.topic} className="mb-3">
                <legend className="fs-6 mb-2">
                  {t(`scope_group.${group.topic}`)}
                </legend>
                {group.scopes.map((scope) => (
                  <Form.Check
                    key={scope}
                    id={`pat-scope-${scope}`}
                    type="checkbox"
                    label={t(`scope.${scope}`)}
                    checked={selectedScopes.includes(scope)}
                    onChange={() => toggleScope(scope)}
                  />
                ))}
              </fieldset>
            ))}
          </Form.Group>
          <Form.Group>
            <Form.Label>{t('expiration')}</Form.Label>
            <Form.Select
              value={expirationDays}
              onChange={(event) => setExpirationDays(event.target.value)}>
              {[7, 30, 90, 365].map((days) => (
                <option key={days} value={String(days)}>
                  {t('days', { count: days })}
                </option>
              ))}
              <option value="custom">{t('custom_expiration')}</option>
            </Form.Select>
            {expirationDays === 'custom' && (
              <Form.Control
                className="mt-2"
                type="date"
                min={dayjs().add(1, 'day').format('YYYY-MM-DD')}
                max={dayjs().add(365, 'day').format('YYYY-MM-DD')}
                value={customExpiration}
                onChange={(event) => setCustomExpiration(event.target.value)}
              />
            )}
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="link"
            disabled={isSubmitting}
            onClick={dismissCreate}>
            {t('cancel', { keyPrefix: 'btns' })}
          </Button>
          <Button
            disabled={!canCreate || isSubmitting || showReauthenticate}
            onClick={createToken}>
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

      <Modal show={showReauthenticate} onHide={dismissReauthentication}>
        <Modal.Header closeButton={!isSubmitting}>
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
            <Button
              disabled={!password || isSubmitting}
              onClick={handleReauthentication}>
              {t('continue')}
            </Button>
          </Modal.Footer>
        )}
      </Modal>
    </div>
  );
};

export default PersonalAccessTokens;

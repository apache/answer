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

// @vitest-environment jsdom

import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { afterEach, beforeEach, expect, test, vi } from 'vitest';

import PersonalAccessTokens from './index';

const {
  security,
  list,
  createToken,
  reauthenticate,
  revokeToken,
  captchaHook,
} = vi.hoisted(() => ({
  security: { personal_access_tokens_enabled: false },
  list: {
    data: [] as any[] | undefined,
    error: undefined as any,
    isLoading: false,
    mutate: vi.fn(),
  },
  createToken: vi.fn(),
  reauthenticate: vi.fn(),
  revokeToken: vi.fn(),
  captchaHook: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@/services', () => ({
  usePersonalAccessTokens: () => list,
  createPersonalAccessToken: createToken,
  revokePersonalAccessToken: revokeToken,
  reauthenticate,
}));

vi.mock('@/stores', () => ({
  siteSecurityStore: () => security,
  loggedUserInfoStore: (selector) =>
    selector({ user: { have_password: true } }),
}));

vi.mock('@/hooks', () => ({
  useToast: () => ({ onShow: vi.fn() }),
}));

vi.mock('@/utils/pluginKit', () => ({
  useCaptchaPlugin: captchaHook,
}));

beforeEach(() => {
  vi.resetAllMocks();
  security.personal_access_tokens_enabled = true;
  list.data = [];
  list.error = undefined;
  list.isLoading = false;
  revokeToken.mockResolvedValue(undefined);
  reauthenticate.mockResolvedValue(undefined);
  list.mutate.mockResolvedValue(undefined);
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function createButton() {
  const buttons = screen.getAllByRole('button', { name: 'create' });
  return buttons[buttons.length - 1] as HTMLButtonElement;
}
function fillCreate(name = 'test-token') {
  fireEvent.click(screen.getByRole('button', { name: 'create' }));
  fireEvent.change(screen.getByRole('textbox'), { target: { value: name } });
  fireEvent.click(screen.getByLabelText('scope.question.read'));
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

test('hides token creation while the instance feature is disabled', () => {
  security.personal_access_tokens_enabled = false;
  render(<PersonalAccessTokens />);

  expect(screen.getByText('disabled')).not.toBeNull();
  expect(screen.queryByRole('button', { name: 'create' })).toBeNull();
});

test('requires an explicit scope selection before token creation', () => {
  security.personal_access_tokens_enabled = true;
  render(<PersonalAccessTokens />);

  fireEvent.click(screen.getByRole('button', { name: 'create' }));
  expect(screen.getByText('scope_group.question')).not.toBeNull();
  expect(screen.getByText('scope_group.answer')).not.toBeNull();
  expect(screen.getByText('scope_group.vote')).not.toBeNull();
  const scopeCheckboxes = screen.getAllByRole('checkbox').slice(1);
  expect(scopeCheckboxes).toHaveLength(5);
  expect(
    scopeCheckboxes.every(
      (checkbox) => !(checkbox as HTMLInputElement).checked,
    ),
  ).toBe(true);
  fireEvent.change(screen.getByRole('textbox'), {
    target: { value: 'valid-token-name' },
  });
  expect(createButton().disabled).toBe(true);
  fireEvent.click(screen.getByLabelText('scope.question.read'));
  expect(createButton().disabled).toBe(false);
});

test('guards token creation until the one-time secret is available', async () => {
  const request = deferred<{ token: string }>();
  createToken.mockReturnValue(request.promise);
  render(<PersonalAccessTokens />);
  fillCreate();
  const submit = createButton();
  act(() => {
    fireEvent.click(submit);
    fireEvent.click(submit);
  });
  expect(createToken).toHaveBeenCalledTimes(1);
  expect(submit.disabled).toBe(true);
  await act(async () => {
    request.resolve({ token: 'one-time-secret' });
  });
  expect(screen.getByDisplayValue('one-time-secret')).not.toBeNull();
});

test('guards reauthentication through the subsequent token creation', async () => {
  const auth = deferred<void>();
  const creation = deferred<{ token: string }>();
  createToken
    .mockRejectedValueOnce({
      reason: 'error.personal_access_token.reauthentication_required',
    })
    .mockReturnValueOnce(creation.promise);
  reauthenticate.mockReturnValue(auth.promise);
  render(<PersonalAccessTokens />);
  fillCreate();
  fireEvent.click(createButton());
  const password = await screen.findByPlaceholderText('password');
  fireEvent.change(password, { target: { value: 'test-password' } });
  const submit = screen.getByRole('button', {
    name: 'continue',
  }) as HTMLButtonElement;
  act(() => {
    fireEvent.click(submit);
    fireEvent.click(submit);
  });
  expect(reauthenticate).toHaveBeenCalledTimes(1);
  expect(submit.disabled).toBe(true);
  await act(async () => {
    auth.resolve();
  });
  await waitFor(() => expect(createToken).toHaveBeenCalledTimes(2));
  const create = createButton();
  expect(create.disabled).toBe(true);
  fireEvent.click(create);
  expect(createToken).toHaveBeenCalledTimes(2);
  await act(async () => {
    creation.resolve({ token: 'reauthenticated-secret' });
  });
  expect(screen.getByDisplayValue('reauthenticated-secret')).not.toBeNull();
});

test('dismissal clears the password and cancels pending reauthentication data', async () => {
  let callback: (() => void) | undefined;
  captchaHook.mockReturnValue({
    check: vi.fn((next) => {
      callback = next;
    }),
    resolveCaptchaReq: vi.fn(),
    close: vi.fn().mockResolvedValue(undefined),
  });
  createToken.mockRejectedValue({
    reason: 'error.personal_access_token.reauthentication_required',
  });
  render(<PersonalAccessTokens />);
  fillCreate();
  fireEvent.click(createButton());
  fireEvent.change(await screen.findByPlaceholderText('password'), {
    target: { value: 'discard-this-password' },
  });
  fireEvent.click(screen.getByRole('button', { name: 'continue' }));
  const closes = screen.getAllByRole('button', { name: 'Close' });
  fireEvent.click(closes[closes.length - 1]);
  await act(async () => {
    callback?.();
  });
  expect(reauthenticate).not.toHaveBeenCalled();
  fireEvent.click(createButton());
  expect(
    ((await screen.findByPlaceholderText('password')) as HTMLInputElement)
      .value,
  ).toBe('');
});

test('temporarily unavailable tokens stay visible and revocable by default', async () => {
  security.personal_access_tokens_enabled = false;
  list.data = [
    {
      id: 11,
      name: 'disabled-token',
      status: 'temporarily_unavailable',
      scopes: ['question.read'],
      expires_at: 2000000000,
    },
    {
      id: 12,
      name: 'expired-token',
      status: 'expired',
      scopes: [],
      expires_at: 1,
    },
    {
      id: 13,
      name: 'revoked-token',
      status: 'revoked',
      scopes: [],
      expires_at: 2000000000,
    },
  ];
  vi.spyOn(window, 'confirm').mockReturnValue(true);
  render(<PersonalAccessTokens />);
  expect(screen.getByText('disabled-token')).not.toBeNull();
  expect(screen.queryByText('expired-token')).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'revoke' }));
  await waitFor(() => expect(revokeToken).toHaveBeenCalledWith(11));
  fireEvent.click(screen.getByLabelText('show_inactive'));
  expect(screen.getByText('expired-token')).not.toBeNull();
  expect(screen.getByText('revoked-token')).not.toBeNull();
  expect(screen.getAllByRole('button', { name: 'revoke' })).toHaveLength(1);
});

test('distinguishes loading and retryable fetch failure from a successful empty list', () => {
  list.data = undefined;
  list.isLoading = true;
  const { rerender } = render(<PersonalAccessTokens />);
  expect(screen.getByRole('status').textContent).toContain('loading');
  expect(screen.queryByText('empty')).toBeNull();
  list.isLoading = false;
  list.error = new Error('timeout');
  rerender(<PersonalAccessTokens />);
  expect(screen.getByRole('alert').textContent).toContain('load_failed');
  expect(screen.queryByText('empty')).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'retry' }));
  expect(list.mutate).toHaveBeenCalledTimes(1);
  list.data = [];
  list.error = undefined;
  rerender(<PersonalAccessTokens />);
  expect(screen.getByText('empty')).not.toBeNull();
  expect(screen.queryByRole('alert')).toBeNull();
});

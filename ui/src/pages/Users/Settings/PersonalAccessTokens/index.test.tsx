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

import { fireEvent, render, screen } from '@testing-library/react';

import PersonalAccessTokens from './index';

const security = {
  personal_access_tokens_enabled: false,
};

jest.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

jest.mock('@/services', () => ({
  usePersonalAccessTokens: () => ({ data: [], mutate: jest.fn() }),
  createPersonalAccessToken: jest.fn(),
  revokePersonalAccessToken: jest.fn(),
  reauthenticate: jest.fn(),
}));

jest.mock('@/stores', () => ({
  siteSecurityStore: () => security,
  loggedUserInfoStore: (selector) =>
    selector({ user: { have_password: true } }),
}));

jest.mock('@/hooks', () => ({
  useToast: () => ({ onShow: jest.fn() }),
}));

jest.mock('@/utils/pluginKit', () => ({
  useCaptchaPlugin: () => undefined,
}));

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
  const createButtons = screen.getAllByRole('button', { name: 'create' });
  expect(createButtons[createButtons.length - 1].hasAttribute('disabled')).toBe(
    true,
  );
});

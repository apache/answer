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
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { afterEach, beforeEach, expect, test, vi } from 'vitest';
import ModifyPass from './index';
const { modifyPassword } = vi.hoisted(() => ({ modifyPassword: vi.fn() }));
vi.mock('@/services', () => ({ modifyPassword }));
vi.mock('@/stores', () => ({
  loggedUserInfoStore: () => ({ user: { have_password: true } }),
}));
vi.mock('@/hooks', () => ({ useToast: () => ({ onShow: vi.fn() }) }));
vi.mock('@/utils/pluginKit', () => ({ useCaptchaPlugin: () => undefined }));
vi.mock('@/utils', () => ({
  handleFormError: vi.fn(),
  scrollToElementTop: vi.fn(),
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
beforeEach(() => {
  vi.resetAllMocks();
  modifyPassword.mockResolvedValue(undefined);
});
afterEach(cleanup);
function fillPassword() {
  fireEvent.change(screen.getByLabelText('current_pass.label'), {
    target: { value: 'old-password' },
  });
  fireEvent.change(screen.getByLabelText('new_pass.label'), {
    target: { value: 'new-password' },
  });
  fireEvent.change(screen.getByLabelText('pass_confirm.label'), {
    target: { value: 'new-password' },
  });
}
for (const action of ['cancel', 'save']) {
  test(`resets the revoke-all option after ${action}`, async () => {
    render(<ModifyPass />);
    fireEvent.click(screen.getByRole('button', { name: 'change_pass_btn' }));
    fillPassword();
    fireEvent.click(screen.getByLabelText('revoke_personal_access_tokens'));
    fireEvent.click(screen.getByRole('button', { name: action }));
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'change_pass_btn' }),
      ).not.toBeNull(),
    );
    fireEvent.click(screen.getByRole('button', { name: 'change_pass_btn' }));
    expect(
      (
        screen.getByLabelText(
          'revoke_personal_access_tokens',
        ) as HTMLInputElement
      ).checked,
    ).toBe(false);
    expect(
      (screen.getByLabelText('current_pass.label') as HTMLInputElement).value,
    ).toBe('');
    fillPassword();
    fireEvent.click(screen.getByRole('button', { name: 'save' }));
    await waitFor(() =>
      expect(modifyPassword).toHaveBeenLastCalledWith({
        old_pass: 'old-password',
        pass: 'new-password',
        revoke_personal_access_tokens: false,
      }),
    );
  });
}

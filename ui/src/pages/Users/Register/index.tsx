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

import React, { useState } from 'react';
import { Container, Col } from 'react-bootstrap';
import { Trans, useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';

import { usePageTags } from '@/hooks';
import type * as Type from '@/common/interface';
import { Unactivate, WelcomeTitle, PluginRender } from '@/components';
import { guard } from '@/utils';
import { loggedUserInfoStore, loginSettingStore } from '@/stores';
import { setupAppTheme } from '@/utils/localize';
import { PluginType } from '@/utils/pluginKit/interface';

import SignUpForm from './components/SignUpForm';

const Index: React.FC = () => {
  const [showForm, setShowForm] = useState(true);
  const { t } = useTranslation('translation', { keyPrefix: 'login' });
  const navigate = useNavigate();
  const loginSetting = loginSettingStore((state) => state.login);
  const updateUser = loggedUserInfoStore((state) => state.update);
  const onRegister = (user: Type.UserInfoRes) => {
    updateUser(user);
    setupAppTheme();
    if (user.mail_status === 2) {
      setShowForm(false);
      return;
    }
    guard.handleLoginRedirect(navigate);
  };
  usePageTags({
    title: t('sign_up', { keyPrefix: 'page_title' }),
  });

  if (!guard.singUpAgent().ok) {
    return null;
  }

  const showSignupForm =
    loginSetting?.allow_new_registrations &&
    loginSetting.allow_email_registrations;

  return (
    <Container style={{ paddingTop: '4rem', paddingBottom: '5rem' }}>
      <WelcomeTitle />

      {showForm ? (
        <Col className="mx-auto" md={6} lg={4} xl={3}>
          <PluginRender
            type={PluginType.Connector}
            slug_name="third_party_connector"
            className="mb-5"
          />
          {showSignupForm ? <SignUpForm callback={onRegister} /> : null}
          <div className="text-center mt-5">
            <Trans i18nKey="login.info_login" ns="translation">
              Already have an account? <Link to="/users/login">Log in</Link>
            </Trans>
          </div>
        </Col>
      ) : (
        <Unactivate visible={!showForm} />
      )}
    </Container>
  );
};

export default React.memo(Index);

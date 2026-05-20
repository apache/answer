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
import { useTranslation } from 'react-i18next';

import type * as Type from '@/common/interface';
import { getLoginSetting, putLoginSetting } from '@/services';
import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import { useToast } from '@/hooks';
import { handleFormError, scrollToElementTop } from '@/utils';
import { loginSettingStore } from '@/stores';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.login',
  });
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      allow_new_registrations: {
        type: 'boolean',
        title: t('membership.title'),
        description: t('membership.text'),
        default: true,
      },
      allow_email_registrations: {
        type: 'boolean',
        title: t('email_registration.title'),
        description: t('email_registration.text'),
        default: true,
      },
      allow_password_login: {
        type: 'boolean',
        title: t('password_login.title'),
        description: t('password_login.text'),
        default: true,
      },
      allow_email_domains: {
        type: 'string',
        title: t('allowed_email_domains.title'),
        description: t('allowed_email_domains.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    allow_new_registrations: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('membership.label'),
      },
    },
    allow_email_registrations: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('email_registration.label'),
      },
    },
    allow_password_login: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('password_login.label'),
      },
    },
    allow_email_domains: {
      'ui:widget': 'textarea',
    },
  };
  const [formData, setFormData] = useState(initFormData(schema));
  const { update: updateLoginSetting } = loginSettingStore((_) => _);

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const allowedEmailDomains: string[] = [];
    if (formData.allow_email_domains.value) {
      const domainList = formData.allow_email_domains.value.split('\n');
      domainList.forEach((li) => {
        li = li.trim();
        if (li) {
          allowedEmailDomains.push(li);
        }
      });
    }
    const reqParams: Type.AdminSettingsLogin = {
      allow_new_registrations: formData.allow_new_registrations.value,
      allow_email_registrations: formData.allow_email_registrations.value,
      allow_email_domains: allowedEmailDomains,
      allow_password_login: formData.allow_password_login.value,
    };

    putLoginSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        updateLoginSetting(reqParams);
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        }
      });
  };

  useEffect(() => {
    getLoginSetting().then((setting) => {
      if (setting) {
        const formMeta = { ...formData };
        formMeta.allow_new_registrations.value =
          setting.allow_new_registrations;
        formMeta.allow_email_registrations.value =
          setting.allow_email_registrations;
        formMeta.allow_email_domains.value = '';
        if (Array.isArray(setting.allow_email_domains)) {
          formMeta.allow_email_domains.value =
            setting.allow_email_domains.join('\n');
        }
        formMeta.allow_password_login.value = setting.allow_password_login;
        setFormData({ ...formMeta });
      }
    });
  }, []);

  const handleOnChange = (data) => {
    setFormData(data);
  };

  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <div className="max-w-748">
        <SchemaForm
          schema={schema}
          formData={formData}
          onSubmit={onSubmit}
          uiSchema={uiSchema}
          onChange={handleOnChange}
        />
      </div>
    </>
  );
};

export default Index;

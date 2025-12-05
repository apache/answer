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

import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import type * as Type from '@/common/interface';
import { useToast } from '@/hooks';
import { useSmtpSetting, updateSmtpSetting } from '@/services';
import pattern from '@/common/pattern';
import { SchemaForm, JSONSchema, UISchema, initFormData } from '@/components';
import { handleFormError, scrollToElementTop } from '@/utils';

const Smtp: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.smtp',
  });
  const Toast = useToast();
  const { data: setting } = useSmtpSetting();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      from_email: {
        type: 'string',
        title: t('from_email.label'),
        description: t('from_email.text'),
      },
      from_name: {
        type: 'string',
        title: t('from_name.label'),
        description: t('from_name.text'),
      },
      smtp_host: {
        type: 'string',
        title: t('smtp_host.label'),
        description: t('smtp_host.text'),
      },
      encryption: {
        type: 'string',
        title: t('encryption.label'),
        description: t('encryption.text'),
        enum: ['TLS', 'SSL', ''],
        enumNames: ['TLS', 'SSL', 'None'],
      },
      smtp_port: {
        type: 'number',
        title: t('smtp_port.label'),
        description: t('smtp_port.text'),
      },
      smtp_authentication: {
        type: 'boolean',
        title: t('smtp_authentication.title'),
        enum: [true, false],
        enumNames: [t('smtp_authentication.yes'), t('smtp_authentication.no')],
      },
      smtp_username: {
        type: 'string',
        title: t('smtp_username.label'),
      },
      smtp_password: {
        type: 'string',
        title: t('smtp_password.label'),
      },
      test_email_recipient: {
        type: 'string',
        title: t('test_email_recipient.label'),
        description: t('test_email_recipient.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    from_email: {
      'ui:options': {
        inputType: 'email',
      },
    },
    encryption: {
      'ui:widget': 'select',
    },
    smtp_username: {
      'ui:options': {
        validator: (value: string, formData) => {
          if (formData.smtp_authentication.value) {
            if (!value) {
              return t('smtp_username.msg');
            }
          }
          return true;
        },
      },
    },
    smtp_password: {
      'ui:options': {
        inputType: 'password',
        validator: (value: string, formData) => {
          if (formData.smtp_authentication.value) {
            if (!value) {
              return t('smtp_password.msg');
            }
          }
          return true;
        },
      },
    },
    smtp_authentication: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('smtp_authentication.label'),
      },
    },
    smtp_port: {
      'ui:options': {
        inputType: 'number',
        validator: (value) => {
          if (!/^[1-9][0-9]*$/.test(value) || Number(value) > 65535) {
            return t('smtp_port.msg');
          }
          return true;
        },
      },
    },
    test_email_recipient: {
      'ui:options': {
        inputType: 'email',
        validator: (value) => {
          if (value && !pattern.email.test(value)) {
            return t('test_email_recipient.msg');
          }
          return true;
        },
      },
    },
  };
  const [formData, setFormData] = useState<Type.FormDataType>(
    initFormData(schema),
  );

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: Type.AdminSettingsSmtp = {
      from_email: formData.from_email.value,
      from_name: formData.from_name.value,
      smtp_host: formData.smtp_host.value,
      encryption: formData.encryption.value,
      smtp_port: formData.smtp_port.value,
      smtp_authentication: formData.smtp_authentication.value,
      ...(formData.smtp_authentication.value
        ? { smtp_username: formData.smtp_username.value }
        : {}),
      ...(formData.smtp_authentication.value
        ? { smtp_password: formData.smtp_password.value }
        : {}),
      test_email_recipient: formData.test_email_recipient.value,
    };

    updateSmtpSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
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
    if (!setting) {
      return;
    }
    const formMeta = {};
    Object.keys(setting).forEach((k) => {
      formMeta[k] = { ...formData[k], value: setting[k] };
    });
    setFormData({ ...formData, ...formMeta });
  }, [setting]);

  useEffect(() => {
    if (!/true|false/.test(formData.smtp_authentication.value)) {
      return;
    }
    if (formData.smtp_authentication.value) {
      setFormData({
        ...formData,
        smtp_username: { ...formData.smtp_username, hidden: false },
        smtp_password: { ...formData.smtp_password, hidden: false },
      });
    } else {
      setFormData({
        ...formData,
        smtp_username: { ...formData.smtp_username, hidden: true },
        smtp_password: { ...formData.smtp_password, hidden: true },
      });
    }
  }, [formData.smtp_authentication.value]);

  const handleOnChange = (data) => {
    setFormData(data);
  };
  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <SchemaForm
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onChange={handleOnChange}
        onSubmit={onSubmit}
      />
    </>
  );
};

export default Smtp;

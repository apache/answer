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

import React, { useEffect, useState, FormEvent } from 'react';
import { useTranslation } from 'react-i18next';

import type { LangsType, FormDataType } from '@/common/interface';
import { useToast } from '@/hooks';
import { updateUserInterface } from '@/services';
import { localize } from '@/utils';
import { loggedUserInfoStore } from '@/stores';
import { SchemaForm, JSONSchema, UISchema } from '@/components';

const Index = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.interface',
  });
  const loggedUserInfo = loggedUserInfoStore.getState().user;
  const toast = useToast();
  const [langs, setLangs] = useState<LangsType[]>();
  const [formData, setFormData] = useState<FormDataType>({
    language: {
      value: loggedUserInfo.language,
      isInvalid: false,
      errorMsg: '',
    },
    color_scheme: {
      value: loggedUserInfo.color_scheme || 'default',
      isInvalid: false,
      errorMsg: '',
    },
  });
  const schema: JSONSchema = {
    title: t('heading'),
    properties: {
      language: {
        type: 'string',
        title: t('lang.label'),
        description: t('lang.text'),
        enum: langs?.map((_) => _.value),
        enumNames: langs?.map((_) => _.label),
        default: loggedUserInfo.language,
      },
      color_scheme: {
        type: 'string',
        title: t('color_scheme.label', { keyPrefix: 'admin.themes' }),
        enum: ['default', 'system', 'light', 'dark'],
        enumNames: [
          t('default', { keyPrefix: 'btns' }),
          t('system_setting', { keyPrefix: 'btns' }),
          t('light', { keyPrefix: 'btns' }),
          t('dark', { keyPrefix: 'btns' }),
        ],
        default: loggedUserInfo.color_scheme,
      },
    },
  };

  const uiSchema: UISchema = {
    language: {
      'ui:widget': 'select',
    },
    color_scheme: {
      'ui:widget': 'select',
    },
  };

  const getLangs = async () => {
    const res: LangsType[] = await localize.loadLanguageOptions();
    setFormData({
      ...formData,
      language: {
        ...formData.language,
        value: loggedUserInfo.language || res[0].value,
      },
    });
    setLangs(res);
  };

  const handleOnChange = (d) => {
    setFormData(d);
  };
  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    const params = {
      language: formData.language.value,
      color_scheme: formData.color_scheme.value,
    };
    updateUserInterface(params).then(() => {
      loggedUserInfoStore.getState().update({
        ...loggedUserInfo,
        ...params,
      });
      localize.setupAppLanguage();
      localize.setupAppTheme();
      toast.onShow({
        msg: t('update', { keyPrefix: 'toast' }),
        variant: 'success',
      });
    });
  };

  useEffect(() => {
    getLangs();
  }, []);
  return (
    <>
      <h3 className="mb-4">{t('heading')}</h3>
      <SchemaForm
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onChange={handleOnChange}
        onSubmit={handleSubmit}
      />
    </>
  );
};

export default React.memo(Index);

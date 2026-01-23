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

import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import { SchemaForm, JSONSchema, UISchema, TabNav } from '@/components';
import {
  ADMIN_USERS_NAV_MENUS,
  SYSTEM_AVATAR_OPTIONS,
} from '@/common/constants';
import { FormDataType } from '@/common/interface';
import { useAdminUsersSettings, updateAdminUsersSettings } from '@/services';
import { useToast } from '@/hooks';
import { siteInfoStore } from '@/stores';
import { handleFormError, scrollToElementTop } from '@/utils';

const UsersSettings = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.interface',
  });
  const { data: setting } = useAdminUsersSettings();
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      default_avatar: {
        type: 'string',
        title: t('avatar.label'),
        description: t('avatar.text'),
        enum: SYSTEM_AVATAR_OPTIONS?.map((v) => v.value),
        enumNames: SYSTEM_AVATAR_OPTIONS?.map((v) => v.label),
        default: setting?.default_avatar || 'system',
      },
      gravatar_base_url: {
        type: 'string',
        title: t('gravatar_base_url.label'),
        description: t('gravatar_base_url.text'),
        default: setting?.gravatar_base_url || '',
      },
    },
  };

  const [formData, setFormData] = useState<FormDataType>({
    default_avatar: {
      value: setting?.default_avatar || 'system',
      isInvalid: false,
      errorMsg: '',
    },
    gravatar_base_url: {
      value: setting?.gravatar_base_url || '',
      isInvalid: false,
      errorMsg: '',
    },
  });

  const uiSchema: UISchema = {
    default_avatar: {
      'ui:widget': 'select',
    },
    gravatar_base_url: {
      'ui:widget': 'input',
      'ui:options': {
        placeholder: 'https://www.gravatar.com/avatar/',
      },
    },
  };

  const handleValueChange = (data: FormDataType) => {
    setFormData(data);
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    const reqParams = {
      default_avatar: formData.default_avatar.value,
      gravatar_base_url: formData.gravatar_base_url.value,
    };
    updateAdminUsersSettings(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        siteInfoStore.getState().updateUsers({
          ...siteInfoStore.getState().users,
          ...reqParams,
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
    if (setting) {
      const formMeta = {};
      Object.keys(setting).forEach((k) => {
        let v = setting[k];
        if (k === 'default_avatar' && !v) {
          v = 'system';
        }
        if (k === 'gravatar_base_url' && !v) {
          v = '';
        }
        formMeta[k] = { ...formData[k], value: v };
      });
      setFormData({ ...formData, ...formMeta });
    }
  }, [setting]);

  return (
    <>
      <h3 className="mb-4">{t('tags', { keyPrefix: 'nav_menus' })}</h3>
      <TabNav menus={ADMIN_USERS_NAV_MENUS} />
      <div className="max-w-748">
        <SchemaForm
          schema={schema}
          uiSchema={uiSchema}
          formData={formData}
          onChange={handleValueChange}
          onSubmit={onSubmit}
        />
      </div>
    </>
  );
};

export default UsersSettings;

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

import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import type * as Type from '@/common/interface';
import { useToast } from '@/hooks';
import { siteInfoStore } from '@/stores';
import { useGeneralSetting, updateGeneralSetting } from '@/services';
import Pattern from '@/common/pattern';
import { REACT_BASE_PATH } from '@/router/alias';
import { handleFormError, scrollToElementTop } from '@/utils';

const General: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.general',
  });
  const Toast = useToast();
  const updateSiteInfo = siteInfoStore((state) => state.update);

  const { data: setting } = useGeneralSetting();
  const schema: JSONSchema = {
    title: t('page_title'),
    required: ['name', 'site_url', 'contact_email'],
    properties: {
      name: {
        type: 'string',
        title: t('name.label'),
        description: t('name.text'),
      },
      site_url: {
        type: 'string',
        title: t('site_url.label'),
        description: t('site_url.text'),
      },
      short_description: {
        type: 'string',
        title: `${t('short_desc.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('short_desc.text'),
      },
      description: {
        type: 'string',
        title: `${t('desc.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('desc.text'),
      },
      contact_email: {
        type: 'string',
        title: t('contact_email.label'),
        description: t('contact_email.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    site_url: {
      'ui:options': {
        inputType: 'url',
        validator: (value) => {
          let url: URL | undefined;
          try {
            url = new URL(value);
          } catch (ex) {
            return t('site_url.validate');
          }
          if (
            !url ||
            !/^https?:$/.test(url.protocol) ||
            (REACT_BASE_PATH && url.pathname !== REACT_BASE_PATH) ||
            (!REACT_BASE_PATH && url.pathname !== '/') ||
            url.search !== '' ||
            url.hash !== ''
          ) {
            return t('site_url.validate');
          }

          return true;
        },
      },
    },
    contact_email: {
      'ui:options': {
        inputType: 'email',
        validator: (value) => {
          if (!Pattern.email.test(value)) {
            return t('contact_email.validate');
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
    const reqParams: Type.AdminSettingsGeneral = {
      name: formData.name.value,
      description: formData.description.value,
      short_description: formData.short_description.value,
      site_url: formData.site_url.value,
      contact_email: formData.contact_email.value,
    };

    updateGeneralSetting(reqParams)
      .then((res) => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        if (res.name) {
          formData.name.value = res.name;
          formData.description.value = res.description;
          formData.short_description.value = res.short_description;
          formData.site_url.value = res.site_url;
          formData.contact_email.value = res.contact_email;
        }

        setFormData({ ...formData });
        updateSiteInfo(res);
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
    const formMeta: Type.FormDataType = {};
    Object.keys(formData).forEach((k) => {
      formMeta[k] = { ...formData[k], value: setting[k] };
    });
    setFormData({ ...formData, ...formMeta });
  }, [setting]);

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

export default General;

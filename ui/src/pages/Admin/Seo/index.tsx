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
import { getSeoSetting, putSeoSetting } from '@/services';
import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import { useToast } from '@/hooks';
import { handleFormError, scrollToElementTop } from '@/utils';
import { seoSettingStore } from '@/stores';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.seo',
  });
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      permalink: {
        type: 'number',
        title: t('permalink.label'),
        description: t('permalink.text'),
        enum: [4, 3, 2, 1],
        enumNames: [
          '/questions/D1D1',
          '/questions/D1D1/post-title',
          '/questions/10010000000000001',
          '/questions/10010000000000001/post-title',
        ],
        default: 4,
      },
      robots: {
        type: 'string',
        title: t('robots.label'),
        description: t('robots.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    permalink: {
      'ui:widget': 'select',
    },
    robots: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: 'font-monospace',
      },
    },
  };
  const [formData, setFormData] = useState(initFormData(schema));

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: Type.AdminSettingsSeo = {
      permalink: Number(formData.permalink.value),
      robots: formData.robots.value,
    };

    putSeoSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        seoSettingStore.getState().update(reqParams);
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
    getSeoSetting().then((setting) => {
      if (setting) {
        const formMeta = { ...formData };
        formMeta.robots.value = setting.robots;
        formMeta.permalink.value = setting.permalink;
        if (!/[1234]/.test(formMeta.permalink.value)) {
          formMeta.permalink.value = 4;
        }
        setFormData(formMeta);
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

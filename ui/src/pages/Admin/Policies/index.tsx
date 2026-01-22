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

import { marked } from 'marked';

import type * as Type from '@/common/interface';
import {
  SchemaForm,
  JSONSchema,
  initFormData,
  UISchema,
  TabNav,
} from '@/components';
import { useToast } from '@/hooks';
import { getPoliciesSetting, putPoliciesSetting } from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { ADMIN_RULES_NAV_MENUS } from '@/common/constants';

const Legal: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.legal',
  });
  const Toast = useToast();

  const schema: JSONSchema = {
    title: t('page_title'),
    required: ['terms_of_service', 'privacy_policy'],
    properties: {
      terms_of_service: {
        type: 'string',
        title: t('terms_of_service.label'),
        description: t('terms_of_service.text'),
      },
      privacy_policy: {
        type: 'string',
        title: t('privacy_policy.label'),
        description: t('privacy_policy.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    terms_of_service: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
      },
    },
    privacy_policy: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
      },
    },
  };
  const [formData, setFormData] = useState(initFormData(schema));

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: Type.AdminSettingsLegal = {
      terms_of_service_original_text: formData.terms_of_service.value,
      terms_of_service_parsed_text: marked.parse(
        formData.terms_of_service.value,
      ),
      privacy_policy_original_text: formData.privacy_policy.value,
      privacy_policy_parsed_text: marked.parse(formData.privacy_policy.value),
    };

    putPoliciesSetting(reqParams)
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
    getPoliciesSetting().then((setting) => {
      if (setting) {
        const formMeta = { ...formData };
        formMeta.terms_of_service.value =
          setting.terms_of_service_original_text;
        formMeta.privacy_policy.value = setting.privacy_policy_original_text;
        setFormData(formMeta);
      }
    });
  }, []);

  const handleOnChange = (data) => {
    setFormData(data);
  };

  return (
    <>
      <h3 className="mb-4">{t('rules', { keyPrefix: 'nav_menus' })}</h3>
      <TabNav menus={ADMIN_RULES_NAV_MENUS} />
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

export default Legal;

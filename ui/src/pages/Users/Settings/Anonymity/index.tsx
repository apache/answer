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

import React, { useState, FormEvent, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import type { FormDataType } from '@/common/interface';
import { useToast } from '@/hooks';
import { useGetAnonymityConfig, putAnonymityConfig } from '@/services';
import { SchemaForm, JSONSchema, UISchema, initFormData } from '@/components';

const Index = () => {
  const toast = useToast();
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.anonymity',
  });
  const { data: configData } = useGetAnonymityConfig();

  const schema: JSONSchema = {
    title: t('heading'),
    properties: {
      enabled: {
        type: 'boolean',
        title: t('enabled.label'),
        description: t('enabled.description'),
        default: configData?.enabled ?? false,
      },
    },
  };

  const uiSchema: UISchema = {
    enabled: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('turn_on'),
      },
    },
  };

  const [formData, setFormData] = useState<FormDataType>(initFormData(schema));

  useEffect(() => {
    setFormData(initFormData(schema));
  }, [configData]);

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();

    putAnonymityConfig({ enabled: formData.enabled.value }).then(() => {
      toast.onShow({
        msg: t('update', { keyPrefix: 'toast' }),
        variant: 'success',
      });
    });
  };

  const handleChange = (ud) => {
    setFormData(ud);
  };

  return (
    <>
      <h3 className="mb-4">{t('heading')}</h3>
      <p className="text-secondary mb-4">{t('intro')}</p>
      <SchemaForm
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onChange={handleChange}
        onSubmit={handleSubmit}
      />
    </>
  );
};

export default React.memo(Index);

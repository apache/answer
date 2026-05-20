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

import {
  SchemaForm,
  JSONSchema,
  UISchema,
  initFormData,
  TabNav,
} from '@/components';
import { ADMIN_QA_NAV_MENUS } from '@/common/constants';
import * as Type from '@/common/interface';
import { writeSettingStore } from '@/stores';
import {
  getQuestionSetting,
  updateQuestionSetting,
} from '@/services/admin/question';
import { handleFormError, scrollToElementTop } from '@/utils';
import { useToast } from '@/hooks';

const QaSettings = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.write',
  });
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      min_tags: {
        type: 'number',
        title: t('min_tags.label'),
        description: t('min_tags.text'),
        default: 0,
      },
      min_content: {
        type: 'number',
        title: t('min_content.label'),
        description: t('min_content.text'),
      },
      restrict_answer: {
        type: 'boolean',
        title: t('restrict_answer.label'),
        description: t('restrict_answer.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    min_tags: {
      'ui:widget': 'input',
      'ui:options': {
        inputType: 'number',
      },
    },
    min_content: {
      'ui:widget': 'input',
      'ui:options': {
        inputType: 'number',
      },
    },
    restrict_answer: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('restrict_answer.label'),
      },
    },
  };
  const [formData, setFormData] = useState<Type.FormDataType>(
    initFormData(schema),
  );

  const handleValueChange = (data: Type.FormDataType) => {
    setFormData(data);
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    // TODO: submit data
    const reqParams: Type.AdminQuestionSetting = {
      min_tags: formData.min_tags.value,
      min_content: formData.min_content.value,
      restrict_answer: formData.restrict_answer.value,
    };
    updateQuestionSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        writeSettingStore.getState().update({ ...reqParams });
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
    getQuestionSetting().then((res) => {
      if (res) {
        const formMeta = { ...formData };
        formMeta.min_tags.value = res.min_tags;
        formMeta.min_content.value = res.min_content;
        formMeta.restrict_answer.value = res.restrict_answer;
        console.log('res', res, formMeta);
        setFormData(formMeta);
      }
    });
  }, []);

  return (
    <>
      <h3 className="mb-4">
        {t('page_title', { keyPrefix: 'admin.questions' })}
      </h3>
      <TabNav menus={ADMIN_QA_NAV_MENUS} />
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

export default QaSettings;

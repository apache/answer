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

import { FC, FormEvent, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { useToast } from '@/hooks';
import { FormDataType } from '@/common/interface';
import { JSONSchema, SchemaForm, UISchema, initFormData } from '@/components';
import {
  getPrivilegeSetting,
  putPrivilegeSetting,
  AdminSettingsPrivilege,
  AdminSettingsPrivilegeReq,
} from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { ADMIN_PRIVILEGE_CUSTOM_LEVEL } from '@/common/constants';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.privilege',
  });
  const Toast = useToast();
  const [privilege, setPrivilege] = useState<AdminSettingsPrivilege>();
  const [schema, setSchema] = useState<JSONSchema>({
    title: t('title'),
    properties: {},
  });
  const [uiSchema, setUiSchema] = useState<UISchema>({
    level: {
      'ui:widget': 'select',
    },
  });
  const [formData, setFormData] = useState<FormDataType>(initFormData(schema));

  const setFormConfig = (state: FormDataType) => {
    const selectedLevel = Number(state.level.value);
    const levelOptions = privilege?.options;
    const curLevel = levelOptions?.find((li) => {
      return li.level === selectedLevel;
    });
    if (!levelOptions || !curLevel) {
      return;
    }
    const uiState = {
      level: uiSchema.level,
    };
    const props: JSONSchema['properties'] = {
      level: {
        type: 'number',
        title: t('level.label'),
        description: t('level.text'),
        enum: levelOptions.map((_) => _.level),
        enumNames: levelOptions.map((_) => _.level_desc),
        default: selectedLevel,
      },
    };
    curLevel.privileges.forEach((li) => {
      props[li.key] = {
        type: 'number',
        title: li.label,
        default: li.value,
      };
      uiState[li.key] = {
        'ui:options': {
          readOnly: curLevel.level !== ADMIN_PRIVILEGE_CUSTOM_LEVEL,
          validator: (value: string) => {
            const val = Number(value);
            if (Number.isNaN(val)) {
              return t('msg.should_be_number');
            }
            if (val < 1) {
              return t('msg.number_larger_1');
            }
            return true;
          },
          inputType: 'number',
          inputMode: 'numeric',
        },
      };
    });
    const schemaState = {
      ...schema,
      properties: props,
    };
    const formState = initFormData(schemaState);
    curLevel.privileges.forEach((li) => {
      formState[li.key] = {
        value: li.value,
        isInvalid: false,
        errorMsg: '',
      };
    });
    setSchema(schemaState);
    setUiSchema(uiState);
    setFormData(formState);
  };

  const onSubmit = (evt: FormEvent) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: AdminSettingsPrivilegeReq = {
      level: Number(formData.level.value),
      custom_privileges: [],
    };

    if (reqParams.level === ADMIN_PRIVILEGE_CUSTOM_LEVEL) {
      // construct custom level request data
      Object.entries(formData).forEach(([key, value]) => {
        if (key === 'level') {
          return;
        }
        reqParams.custom_privileges?.push({
          key,
          value: Number(value.value),
        });
      });
    }

    putPrivilegeSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        if (reqParams.level === ADMIN_PRIVILEGE_CUSTOM_LEVEL) {
          getPrivilegeSetting().then((resp) => {
            setPrivilege(resp);
          });
        }
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
    if (!privilege) {
      return;
    }
    setFormConfig({
      level: {
        value: privilege.selected_level,
        isInvalid: false,
        errorMsg: '',
      },
    });
  }, [privilege]);
  useEffect(() => {
    getPrivilegeSetting().then((resp) => {
      setPrivilege(resp);
    });
  }, []);
  const handleOnChange = (state) => {
    // if updated values in Custom form
    if (
      state.level.value === ADMIN_PRIVILEGE_CUSTOM_LEVEL &&
      formData?.level?.value === state.level.value
    ) {
      setFormData(state);
    } else {
      setFormConfig(state);
    }
  };

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <SchemaForm
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onSubmit={onSubmit}
        onChange={handleOnChange}
      />
    </>
  );
};

export default Index;

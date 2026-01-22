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

import { FC, memo, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { JSONSchema, SchemaForm, UISchema, ImgViewer } from '@/components';
import { FormDataType } from '@/common/interface';
import { brandSetting, getBrandSetting } from '@/services';
import { brandingStore } from '@/stores';
import { useToast } from '@/hooks';
import { handleFormError, scrollToElementTop } from '@/utils';

const uploadType = 'branding';
const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.branding',
  });
  const { branding: brandingInfo, update } = brandingStore();
  const Toast = useToast();

  const [formData, setFormData] = useState<FormDataType>({
    logo: {
      value: brandingInfo.logo,
      isInvalid: false,
      errorMsg: '',
    },
    mobile_logo: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    square_icon: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    favicon: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });

  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      logo: {
        type: 'string',
        title: `${t('logo.label')} ${t('optional', { keyPrefix: 'form' })}`,
        description: t('logo.text'),
      },
      mobile_logo: {
        type: 'string',
        title: `${t('mobile_logo.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('mobile_logo.text'),
      },
      square_icon: {
        type: 'string',
        title: `${t('square_icon.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('square_icon.text'),
      },
      favicon: {
        type: 'string',
        title: `${t('favicon.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('favicon.text'),
      },
    },
  };

  const uiSchema: UISchema = {
    logo: {
      'ui:widget': 'upload',
      'ui:options': {
        imageType: uploadType,
        className: 'object-fit-contain',
      },
    },
    mobile_logo: {
      'ui:widget': 'upload',
      'ui:options': {
        imageType: uploadType,
        className: 'object-fit-contain',
      },
    },
    square_icon: {
      'ui:widget': 'upload',
      'ui:options': {
        imageType: uploadType,
        className: 'object-fit-contain',
      },
    },
    favicon: {
      'ui:widget': 'upload',
      'ui:options': {
        acceptType: ',image/x-icon,image/vnd.microsoft.icon',
        imageType: uploadType,
        className: 'object-fit-contain',
      },
    },
  };

  const handleOnChange = (data) => {
    setFormData(data);
  };

  const onSubmit = () => {
    const params = {
      logo: formData.logo.value,
      mobile_logo: formData.mobile_logo.value,
      square_icon: formData.square_icon.value,
      favicon: formData.favicon.value,
    };
    brandSetting(params)
      .then(() => {
        update(params);
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

  const getBrandData = async () => {
    const res = await getBrandSetting();
    if (res) {
      formData.logo.value = res.logo;
      formData.mobile_logo.value = res.mobile_logo;
      formData.square_icon.value = res.square_icon;
      formData.favicon.value = res.favicon;
      setFormData({ ...formData });
    }
  };

  useEffect(() => {
    getBrandData();
  }, []);

  return (
    <ImgViewer>
      <h3 className="mb-4">{t('page_title')}</h3>
      <div className="max-w-748">
        <SchemaForm
          schema={schema}
          uiSchema={uiSchema}
          formData={formData}
          onSubmit={onSubmit}
          onChange={handleOnChange}
        />
      </div>
    </ImgViewer>
  );
};

export default memo(Index);

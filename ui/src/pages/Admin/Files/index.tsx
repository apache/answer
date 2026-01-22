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
import { Form, Button } from 'react-bootstrap';

import type * as Type from '@/common/interface';
import { useToast } from '@/hooks';
import { getAdminFilesSetting, updateAdminFilesSetting } from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { writeSettingStore } from '@/stores';

const initFormData = {
  max_image_size: {
    value: 0,
    errorMsg: '',
    isInvalid: false,
  },
  max_attachment_size: {
    value: 0,
    errorMsg: '',
    isInvalid: false,
  },
  max_image_megapixel: {
    value: 0,
    errorMsg: '',
    isInvalid: false,
  },
  authorized_image_extensions: {
    value: '',
    errorMsg: '',
    isInvalid: false,
  },
  authorized_attachment_extensions: {
    value: '',
    errorMsg: '',
    isInvalid: false,
  },
};

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.write',
  });
  const Toast = useToast();

  const [formData, setFormData] = useState(initFormData);

  const handleValueChange = (value) => {
    setFormData({
      ...formData,
      ...value,
    });
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: Type.AdminSettingsWrite = {
      max_image_size: Number(formData.max_image_size.value),
      max_attachment_size: Number(formData.max_attachment_size.value),
      max_image_megapixel: Number(formData.max_image_megapixel.value),
      authorized_image_extensions:
        formData.authorized_image_extensions.value?.length > 0
          ? formData.authorized_image_extensions.value
              .split(',')
              ?.map((item) => item.trim().toLowerCase())
          : [],
      authorized_attachment_extensions:
        formData.authorized_attachment_extensions.value?.length > 0
          ? formData.authorized_attachment_extensions.value
              .split(',')
              ?.map((item) => item.trim().toLowerCase())
          : [],
    };
    updateAdminFilesSetting(reqParams)
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

  const initData = () => {
    getAdminFilesSetting().then((res) => {
      formData.max_image_size.value = res.max_image_size;
      formData.max_attachment_size.value = res.max_attachment_size;
      formData.max_image_megapixel.value = res.max_image_megapixel;
      formData.authorized_image_extensions.value =
        res.authorized_image_extensions?.join(', ').toLowerCase();
      formData.authorized_attachment_extensions.value =
        res.authorized_attachment_extensions?.join(', ').toLowerCase();
      setFormData({ ...formData });
    });
  };

  useEffect(() => {
    initData();
  }, []);

  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <div className="max-w-748">
        <Form noValidate onSubmit={onSubmit}>
          <Form.Group className="mb-3" controlId="max_image_size">
            <Form.Label>{t('image_size.label')}</Form.Label>
            <Form.Control
              type="number"
              inputMode="numeric"
              min={0}
              value={formData.max_image_size.value}
              isInvalid={formData.max_image_size.isInvalid}
              onChange={(evt) => {
                handleValueChange({
                  max_image_size: {
                    value: evt.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Text>{t('image_size.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.max_image_size.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="max_attachment_size">
            <Form.Label>{t('attachment_size.label')}</Form.Label>
            <Form.Control
              type="number"
              inputMode="numeric"
              min={0}
              value={formData.max_attachment_size.value}
              isInvalid={formData.max_attachment_size.isInvalid}
              onChange={(evt) => {
                handleValueChange({
                  max_attachment_size: {
                    value: evt.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Text>{t('attachment_size.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.max_attachment_size.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="max_image_megapixel">
            <Form.Label>{t('image_megapixels.label')}</Form.Label>
            <Form.Control
              type="number"
              inputMode="numeric"
              min={0}
              isInvalid={formData.max_image_megapixel.isInvalid}
              value={formData.max_image_megapixel.value}
              onChange={(evt) => {
                handleValueChange({
                  max_image_megapixel: {
                    value: evt.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Text>{t('image_megapixels.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.max_image_megapixel.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="authorized_image_extensions">
            <Form.Label>{t('image_extensions.label')}</Form.Label>
            <Form.Control
              type="text"
              value={formData.authorized_image_extensions.value}
              isInvalid={formData.authorized_image_extensions.isInvalid}
              onChange={(evt) => {
                handleValueChange({
                  authorized_image_extensions: {
                    value: evt.target.value.toLowerCase(),
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Text>{t('image_extensions.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.authorized_image_extensions.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group
            className="mb-3"
            controlId="authorized_attachment_extensions">
            <Form.Label>{t('attachment_extensions.label')}</Form.Label>
            <Form.Control
              type="text"
              value={formData.authorized_attachment_extensions.value}
              isInvalid={formData.authorized_attachment_extensions.isInvalid}
              onChange={(evt) => {
                handleValueChange({
                  authorized_attachment_extensions: {
                    value: evt.target.value.toLowerCase(),
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Text>{t('attachment_extensions.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.authorized_attachment_extensions.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3">
            <Button type="submit">{t('save', { keyPrefix: 'btns' })}</Button>
          </Form.Group>
        </Form>
      </div>
    </>
  );
};

export default Index;

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

import React, { FC, FormEvent, useState } from 'react';
import { Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classname from 'classnames';

import { useToast } from '@/hooks';
import { useCaptchaPlugin } from '@/utils/pluginKit';
import type { FormDataType } from '@/common/interface';
import { modifyPassword } from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { loggedUserInfoStore } from '@/stores';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.account',
  });
  const { user } = loggedUserInfoStore();
  const [showForm, setFormState] = useState(false);
  const toast = useToast();
  const [formData, setFormData] = useState<FormDataType>({
    old_pass: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    pass: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    pass2: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    revoke_personal_access_tokens: {
      value: false,
      isInvalid: false,
      errorMsg: '',
    },
  });

  const infoCaptcha = useCaptchaPlugin('edit_userinfo');

  const handleFormState = () => {
    setFormState((pre) => !pre);
  };

  const handleChange = (params: FormDataType) => {
    setFormData({ ...formData, ...params });
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const { old_pass, pass, pass2 } = formData;
    if (!old_pass.value && user.have_password) {
      bol = false;
      formData.old_pass = {
        value: '',
        isInvalid: true,
        errorMsg: t('current_pass.msg.empty'),
      };
    }

    if (!pass.value) {
      bol = false;
      formData.pass = {
        value: '',
        isInvalid: true,
        errorMsg: t('current_pass.msg.empty'),
      };
    }

    if (bol && pass.value && pass.value.length < 8) {
      bol = false;
      formData.pass = {
        value: pass.value,
        isInvalid: true,
        errorMsg: t('current_pass.msg.length'),
      };
    }

    if (!pass2.value) {
      bol = false;
      formData.pass2 = {
        value: '',
        isInvalid: true,
        errorMsg: t('current_pass.msg.empty'),
      };
    }

    if (bol && pass2.value && pass2.value.length < 8) {
      bol = false;
      formData.pass2 = {
        value: pass2.value,
        isInvalid: true,
        errorMsg: t('current_pass.msg.length'),
      };
    }
    if (bol && pass.value && pass2.value && pass.value !== pass2.value) {
      bol = false;
      formData.pass2 = {
        value: pass2.value,
        isInvalid: true,
        errorMsg: t('current_pass.msg.different'),
      };
    }
    setFormData({
      ...formData,
    });
    if (!bol) {
      const errObj = Object.keys(formData).filter(
        (key) => formData[key].isInvalid,
      );
      const ele = document.getElementById(errObj[0]);
      scrollToElementTop(ele);
    }

    return bol;
  };

  const postModifyPass = (event?: any) => {
    if (event) {
      event.preventDefault();
    }
    const params: any = {
      old_pass: formData.old_pass.value,
      pass: formData.pass.value,
      revoke_personal_access_tokens:
        formData.revoke_personal_access_tokens.value,
    };

    const imgCode = infoCaptcha?.getCaptcha();
    if (imgCode?.verify) {
      params.captcha_code = imgCode.captcha_code;
      params.captcha_id = imgCode.captcha_id;
    }
    modifyPassword(params)
      .then(async () => {
        await infoCaptcha?.close();
        toast.onShow({
          msg: t('update_password', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        handleFormState();
      })
      .catch((err) => {
        if (err.isError) {
          infoCaptcha?.handleCaptchaError(err.list);
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        }
      });
  };

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    if (!infoCaptcha) {
      postModifyPass();
      return;
    }

    infoCaptcha.check(() => {
      postModifyPass();
    });
  };

  return (
    <div className="mt-5">
      {showForm ? (
        <Form noValidate onSubmit={handleSubmit}>
          <Form.Group
            controlId="old_pass"
            className={classname('mb-3', user.have_password ? '' : 'd-none')}>
            <Form.Label>{t('current_pass.label')}</Form.Label>
            <Form.Control
              autoComplete="off"
              required
              type="password"
              placeholder=""
              isInvalid={formData.old_pass.isInvalid}
              onChange={(e) =>
                handleChange({
                  old_pass: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                })
              }
            />
            <Form.Control.Feedback type="invalid">
              {formData.old_pass.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group controlId="new_pass" className="mb-3">
            <Form.Label>{t('new_pass.label')}</Form.Label>
            <Form.Control
              autoComplete="off"
              required
              type="password"
              isInvalid={formData.pass.isInvalid}
              onChange={(e) =>
                handleChange({
                  pass: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                })
              }
            />
            <Form.Control.Feedback type="invalid">
              {formData.pass.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group controlId="pass2" className="mb-3">
            <Form.Label>{t('pass_confirm.label')}</Form.Label>
            <Form.Control
              autoComplete="off"
              required
              type="password"
              isInvalid={formData.pass2.isInvalid}
              onChange={(e) =>
                handleChange({
                  pass2: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                })
              }
            />
            <Form.Control.Feedback type="invalid">
              {formData.pass2.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>
          <Form.Group
            controlId="revoke_personal_access_tokens"
            className="mb-3">
            <Form.Check
              type="checkbox"
              label={t('revoke_personal_access_tokens')}
              checked={formData.revoke_personal_access_tokens.value}
              onChange={(e) =>
                handleChange({
                  revoke_personal_access_tokens: {
                    value: e.target.checked,
                    isInvalid: false,
                    errorMsg: '',
                  },
                })
              }
            />
          </Form.Group>
          <div>
            <Button type="submit" variant="primary" className="me-2">
              {t('save', { keyPrefix: 'btns' })}
            </Button>

            <Button variant="link" onClick={() => handleFormState()}>
              {t('cancel', { keyPrefix: 'btns' })}
            </Button>
          </div>
        </Form>
      ) : (
        <>
          <Form.Label>{t('password_title')}</Form.Label>
          <br />
          <Button
            variant="outline-secondary"
            type="submit"
            onClick={() => {
              handleFormState();
            }}>
            {t('change_pass_btn')}
          </Button>
        </>
      )}
    </div>
  );
};

export default React.memo(Index);

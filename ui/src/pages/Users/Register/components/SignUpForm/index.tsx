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

import React, { FormEvent, useState } from 'react';
import { Form, Button } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { Trans, useTranslation } from 'react-i18next';

import { useCaptchaPlugin } from '@/utils/pluginKit';
import type {
  FormDataType,
  RegisterReqParams,
  UserInfoRes,
} from '@/common/interface';
import { register } from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { useLegalClick } from '@/behaviour/useLegalClick';

interface Props {
  callback: (user: UserInfoRes) => void;
}

const Index: React.FC<Props> = ({ callback }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'login' });
  const [formData, setFormData] = useState<FormDataType>({
    name: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    e_mail: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    pass: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });

  const emailCaptcha = useCaptchaPlugin('email');
  const nameRegex = /^[\w.-\s]{2,30}$/;

  const handleChange = (params: FormDataType) => {
    setFormData({ ...formData, ...params });
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const { name, e_mail, pass } = formData;

    if (!name.value) {
      bol = false;
      formData.name = {
        value: '',
        isInvalid: true,
        errorMsg: t('name.msg.empty'),
      };
    } else if (name.value.length < 2 || name.value.length > 30) {
      bol = false;
      formData.name = {
        value: name.value,
        isInvalid: true,
        errorMsg: t('name.msg.range'),
      };
    } else if (!nameRegex.test(name.value)) {
      bol = false;
      formData.name = {
        value: name.value,
        isInvalid: true,
        errorMsg: t('name.msg.character'),
      };
    }

    if (!e_mail.value) {
      bol = false;
      formData.e_mail = {
        value: '',
        isInvalid: true,
        errorMsg: t('email.msg.empty'),
      };
    }

    if (!pass.value) {
      bol = false;
      formData.pass = {
        value: '',
        isInvalid: true,
        errorMsg: t('password.msg.empty'),
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

  const legalClick = useLegalClick();

  const handleRegister = (event?: any) => {
    if (event) {
      event.preventDefault();
    }
    const reqParams: RegisterReqParams = {
      name: formData.name.value,
      e_mail: formData.e_mail.value,
      pass: formData.pass.value,
    };

    const captcha = emailCaptcha?.getCaptcha();
    if (captcha?.verify) {
      reqParams.captcha_code = captcha.captcha_code;
      reqParams.captcha_id = captcha.captcha_id;
    }

    register(reqParams)
      .then(async (res) => {
        await emailCaptcha?.close();
        callback(res);
      })
      .catch((err) => {
        if (err.isError) {
          emailCaptcha?.handleCaptchaError(err.list);
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        }
      });
  };

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    if (!emailCaptcha) {
      handleRegister();
      return;
    }
    emailCaptcha.check(() => {
      handleRegister();
    });
  };

  return (
    <>
      <Form noValidate onSubmit={handleSubmit} autoComplete="off">
        <Form.Group controlId="name" className="mb-3">
          <Form.Label>{t('name.label')}</Form.Label>
          <Form.Control
            autoComplete="off"
            required
            type="text"
            isInvalid={formData.name.isInvalid}
            value={formData.name.value}
            onChange={(e) =>
              handleChange({
                name: {
                  value: e.target.value,
                  isInvalid: false,
                  errorMsg: '',
                },
              })
            }
          />
          <Form.Control.Feedback type="invalid">
            {formData.name.errorMsg}
          </Form.Control.Feedback>
        </Form.Group>
        <Form.Group controlId="email" className="mb-3">
          <Form.Label>{t('email.label')}</Form.Label>
          <Form.Control
            autoComplete="off"
            required
            type="e_mail"
            isInvalid={formData.e_mail.isInvalid}
            value={formData.e_mail.value}
            onChange={(e) =>
              handleChange({
                e_mail: {
                  value: e.target.value,
                  isInvalid: false,
                  errorMsg: '',
                },
              })
            }
          />
          <Form.Control.Feedback type="invalid">
            {formData.e_mail.errorMsg}
          </Form.Control.Feedback>
        </Form.Group>

        <Form.Group controlId="password" className="mb-3">
          <Form.Label>{t('password.label')}</Form.Label>
          <Form.Control
            autoComplete="off"
            required
            type="password"
            isInvalid={formData.pass.isInvalid}
            value={formData.pass.value}
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

        <div className="d-grid">
          <Button variant="primary" type="submit">
            {t('signup', { keyPrefix: 'btns' })}
          </Button>
        </div>
      </Form>
      <div className="text-center small mt-3">
        <Trans i18nKey="login.agreements" ns="translation">
          By registering, you agree to the
          <Link
            to="/privacy"
            onClick={(evt) => {
              legalClick(evt, 'privacy');
            }}
            target="_blank">
            privacy policy
          </Link>
          and
          <Link
            to="/tos"
            onClick={(evt) => {
              legalClick(evt, 'tos');
            }}
            target="_blank">
            terms of service
          </Link>
          .
        </Trans>
      </div>
    </>
  );
};

export default React.memo(Index);

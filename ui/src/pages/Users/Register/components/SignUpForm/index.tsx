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

import React, { FormEvent, useEffect, useState } from 'react';
import { Form, Button, InputGroup } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { Trans, useTranslation } from 'react-i18next';

import { useCaptchaPlugin } from '@/utils/pluginKit';
import type {
  FormDataType,
  RegisterEmailCodeReq,
  RegisterReqParams,
  UserInfoRes,
} from '@/common/interface';
import { register, sendRegisterEmailCode } from '@/services';
import { handleFormError, scrollToElementTop } from '@/utils';
import { useLegalClick } from '@/behaviour/useLegalClick';
import { useToast } from '@/hooks';

interface Props {
  callback: (user: UserInfoRes) => void;
}

const EMAIL_CODE_COUNTDOWN_SECONDS = 60;
const ALLOWED_EMAIL_DOMAINS = new Set([
  'hainanu.edu.cn',
  'alumni.hainanu.edu.cn',
]);

const normalizeEmail = (email: string) => email.trim().toLowerCase();

const getEmailDomain = (email: string) => {
  const normalizedEmail = normalizeEmail(email);
  const atIndex = normalizedEmail.lastIndexOf('@');
  if (atIndex <= 0 || atIndex !== normalizedEmail.indexOf('@')) {
    return '';
  }
  return normalizedEmail.slice(atIndex + 1);
};

const isAllowedEmail = (email: string) =>
  ALLOWED_EMAIL_DOMAINS.has(getEmailDomain(email));

const Index: React.FC<Props> = ({ callback }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'login' });
  const Toast = useToast();
  const [emailCodeCountdown, setEmailCodeCountdown] = useState(0);
  const [emailCodeSending, setEmailCodeSending] = useState(false);
  const [emailCodeSentTo, setEmailCodeSentTo] = useState('');
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
    pass_confirm: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    email_code: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });

  const emailCaptcha = useCaptchaPlugin('email');
  const nameRegex = /^[\w.-\s]{2,30}$/;

  useEffect(() => {
    if (emailCodeCountdown <= 0) {
      return undefined;
    }
    const timer = window.setTimeout(() => {
      setEmailCodeCountdown((seconds) => Math.max(0, seconds - 1));
    }, 1000);
    return () => window.clearTimeout(timer);
  }, [emailCodeCountdown]);

  const handleChange = (params: FormDataType) => {
    setFormData({ ...formData, ...params });
  };

  const checkEmailValidated = (): boolean => {
    const email = normalizeEmail(formData.e_mail.value);
    if (!email) {
      handleChange({
        e_mail: {
          value: '',
          isInvalid: true,
          errorMsg: t('email.msg.empty'),
        },
      });
      return false;
    }
    if (!/^[^\s@]+@[^\s@]+$/.test(email)) {
      handleChange({
        e_mail: {
          value: email,
          isInvalid: true,
          errorMsg: t('email.msg.invalid'),
        },
      });
      return false;
    }
    if (!isAllowedEmail(email)) {
      handleChange({
        e_mail: {
          value: email,
          isInvalid: true,
          errorMsg: t('email.msg.domain'),
        },
      });
      return false;
    }
    return true;
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const {
      name,
      pass,
      pass_confirm: passConfirm,
      email_code: emailCode,
    } = formData;

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

    const email = normalizeEmail(formData.e_mail.value);
    if (!email) {
      bol = false;
      formData.e_mail = {
        value: '',
        isInvalid: true,
        errorMsg: t('email.msg.empty'),
      };
    } else if (!/^[^\s@]+@[^\s@]+$/.test(email)) {
      bol = false;
      formData.e_mail = {
        value: email,
        isInvalid: true,
        errorMsg: t('email.msg.invalid'),
      };
    } else if (!isAllowedEmail(email)) {
      bol = false;
      formData.e_mail = {
        value: email,
        isInvalid: true,
        errorMsg: t('email.msg.domain'),
      };
    }

    if (!pass.value) {
      bol = false;
      formData.pass = {
        value: '',
        isInvalid: true,
        errorMsg: t('password.msg.empty'),
      };
    } else if (pass.value.length < 8 || pass.value.length > 32) {
      bol = false;
      formData.pass = {
        value: pass.value,
        isInvalid: true,
        errorMsg: t('password.msg.range'),
      };
    }

    if (!passConfirm.value) {
      bol = false;
      formData.pass_confirm = {
        value: '',
        isInvalid: true,
        errorMsg: t('password_confirm.msg.empty'),
      };
    } else if (pass.value !== passConfirm.value) {
      bol = false;
      formData.pass_confirm = {
        value: passConfirm.value,
        isInvalid: true,
        errorMsg: t('password_confirm.msg.different'),
      };
    }

    if (!emailCode.value) {
      bol = false;
      formData.email_code = {
        value: '',
        isInvalid: true,
        errorMsg: t('verification_code.msg.empty'),
      };
    } else if (!/^\d{6}$/.test(emailCode.value)) {
      bol = false;
      formData.email_code = {
        value: emailCode.value,
        isInvalid: true,
        errorMsg: t('verification_code.msg.invalid'),
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

  const sendEmailCode = () => {
    const reqParams: RegisterEmailCodeReq = {
      e_mail: normalizeEmail(formData.e_mail.value),
    };
    const captcha = emailCaptcha?.getCaptcha();
    if (captcha?.verify) {
      reqParams.captcha_code = captcha.captcha_code;
      reqParams.captcha_id = captcha.captcha_id;
    }

    setEmailCodeSending(true);
    sendRegisterEmailCode(reqParams)
      .then(async () => {
        await emailCaptcha?.close();
        setEmailCodeSentTo(reqParams.e_mail);
        setEmailCodeCountdown(EMAIL_CODE_COUNTDOWN_SECONDS);
        Toast.onShow({
          msg: t('verification_code.sent'),
          variant: 'success',
        });
      })
      .catch((err) => {
        if (err?.code === 429 && err?.data?.retry_after) {
          setEmailCodeCountdown(Number(err.data.retry_after));
        }
        if (err?.isError) {
          const captchaError = emailCaptcha?.handleCaptchaError(err.list);
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const firstFormError = err.list.find(
            (item: { error_field: string }) =>
              item.error_field !== 'captcha_code',
          );
          if (!captchaError || firstFormError) {
            const ele = document.getElementById(
              firstFormError?.error_field || err.list[0].error_field,
            );
            scrollToElementTop(ele);
          }
          return;
        }
        if (err?.msg) {
          Toast.onShow({ msg: err.msg, variant: 'danger' });
        }
      })
      .finally(() => {
        setEmailCodeSending(false);
      });
  };

  const handleSendEmailCode = () => {
    if (emailCodeSending || emailCodeCountdown > 0) {
      return;
    }
    if (!checkEmailValidated()) {
      return;
    }
    if (!emailCaptcha) {
      sendEmailCode();
      return;
    }
    emailCaptcha.check(sendEmailCode);
  };

  const legalClick = useLegalClick();

  const handleRegister = () => {
    const reqParams: RegisterReqParams = {
      name: formData.name.value,
      e_mail: normalizeEmail(formData.e_mail.value),
      pass: formData.pass.value,
      pass_confirm: formData.pass_confirm.value,
      email_code: formData.email_code.value,
    };

    register(reqParams)
      .then((res) => {
        callback(res);
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

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    handleRegister();
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
        <Form.Group controlId="e_mail" className="mb-3">
          <Form.Label>{t('email.label')}</Form.Label>
          <Form.Control
            autoComplete="off"
            required
            type="email"
            isInvalid={formData.e_mail.isInvalid}
            value={formData.e_mail.value}
            onChange={(e) => {
              const email = e.target.value;
              const changedAfterCodeSent = Boolean(
                emailCodeSentTo && normalizeEmail(email) !== emailCodeSentTo,
              );
              if (changedAfterCodeSent) {
                setEmailCodeSentTo('');
                setEmailCodeCountdown(0);
              }
              handleChange({
                e_mail: {
                  value: email,
                  isInvalid: false,
                  errorMsg: '',
                },
                email_code: {
                  value: changedAfterCodeSent ? '' : formData.email_code.value,
                  isInvalid: false,
                  errorMsg: '',
                },
              });
            }}
          />
          <Form.Control.Feedback type="invalid">
            {formData.e_mail.errorMsg}
          </Form.Control.Feedback>
        </Form.Group>

        <Form.Group controlId="email_code" className="mb-3">
          <Form.Label>{t('verification_code.label')}</Form.Label>
          <InputGroup hasValidation>
            <Form.Control
              autoComplete="one-time-code"
              required
              inputMode="numeric"
              maxLength={6}
              placeholder={t('verification_code.placeholder')}
              isInvalid={formData.email_code.isInvalid}
              value={formData.email_code.value}
              onChange={(e) =>
                handleChange({
                  email_code: {
                    value: e.target.value.replace(/\D/g, '').slice(0, 6),
                    isInvalid: false,
                    errorMsg: '',
                  },
                })
              }
            />
            <Button
              type="button"
              variant="outline-secondary"
              disabled={emailCodeSending || emailCodeCountdown > 0}
              className="text-nowrap"
              onClick={handleSendEmailCode}>
              {emailCodeSending
                ? t('verification_code.sending')
                : emailCodeCountdown > 0
                  ? t('verification_code.resend', {
                      seconds: emailCodeCountdown,
                    })
                  : t('verification_code.send')}
            </Button>
            <Form.Control.Feedback type="invalid">
              {formData.email_code.errorMsg}
            </Form.Control.Feedback>
          </InputGroup>
        </Form.Group>

        <Form.Group controlId="pass" className="mb-3">
          <Form.Label>{t('password.label')}</Form.Label>
          <Form.Control
            autoComplete="new-password"
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

        <Form.Group controlId="pass_confirm" className="mb-3">
          <Form.Label>{t('password_confirm.label')}</Form.Label>
          <Form.Control
            autoComplete="new-password"
            required
            type="password"
            isInvalid={formData.pass_confirm.isInvalid}
            value={formData.pass_confirm.value}
            onChange={(e) =>
              handleChange({
                pass_confirm: {
                  value: e.target.value,
                  isInvalid: false,
                  errorMsg: '',
                },
              })
            }
          />
          <Form.Control.Feedback type="invalid">
            {formData.pass_confirm.errorMsg}
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

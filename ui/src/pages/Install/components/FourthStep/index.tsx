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

import { FC, FormEvent } from 'react';
import { Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import type { FormDataType } from '@/common/interface';
import Pattern from '@/common/pattern';
import Progress from '../Progress';

interface Props {
  data: FormDataType;
  changeCallback: (value: FormDataType) => void;
  nextCallback: () => void;
  visible: boolean;
}
const Index: FC<Props> = ({ visible, data, changeCallback, nextCallback }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'install' });

  const checkValidated = (): boolean => {
    let bol = true;
    const { site_name, site_url, contact_email, name, password, email } = data;

    if (!site_name.value) {
      bol = false;
      data.site_name = {
        value: '',
        isInvalid: true,
        errorMsg: t('site_name.msg'),
      };
    }

    if (site_name.value && site_name.value.length > 30) {
      bol = false;
      data.site_url = {
        value: site_name.value,
        isInvalid: true,
        errorMsg: t('site_name.msg_max_length'),
      };
    }

    if (!site_url.value) {
      bol = false;
      data.site_url = {
        value: '',
        isInvalid: true,
        errorMsg: t('site_name.msg.empty'),
      };
    }

    const reg = /^(http|https):\/\//g;
    if (site_url.value && !site_url.value.match(reg)) {
      bol = false;
      data.site_url = {
        value: site_url.value,
        isInvalid: true,
        errorMsg: t('site_url.msg.incorrect'),
      };
    } else if (site_url.value.length > 512) {
      bol = false;
      data.site_url = {
        value: site_url.value,
        isInvalid: true,
        errorMsg: t('site_url.msg.max_length'),
      };
    }

    if (!contact_email.value) {
      bol = false;
      data.contact_email = {
        value: '',
        isInvalid: true,
        errorMsg: t('contact_email.msg.empty'),
      };
    }

    if (contact_email.value && !Pattern.email.test(contact_email.value)) {
      bol = false;
      data.contact_email = {
        value: contact_email.value,
        isInvalid: true,
        errorMsg: t('contact_email.msg.incorrect'),
      };
    }

    if (!name.value) {
      bol = false;
      data.name = {
        value: '',
        isInvalid: true,
        errorMsg: t('admin_name.msg'),
      };
    } else if (/[^a-z0-9\-._]/.test(name.value)) {
      bol = false;
      data.name = {
        value: name.value,
        isInvalid: true,
        errorMsg: t('admin_name.character'),
      };
    } else if (data.name.value.length > 30) {
      bol = false;
      data.name = {
        value: data.name.value,
        isInvalid: true,
        errorMsg: t('admin_name.msg_max_length'),
      };
    }

    if (!password.value) {
      bol = false;
      data.password = {
        value: '',
        isInvalid: true,
        errorMsg: t('admin_password.msg'),
      };
    }

    if (password.value && password.value.length < 4) {
      bol = false;
      data.password = {
        value: data.password.value,
        isInvalid: true,
        errorMsg: t('admin_password.msg_min_length'),
      };
    }

    if (password.value && password.value.length > 32) {
      bol = false;
      data.password = {
        value: data.password.value,
        isInvalid: true,
        errorMsg: t('admin_password.msg_max_length'),
      };
    }

    if (!email.value) {
      bol = false;
      data.email = {
        value: '',
        isInvalid: true,
        errorMsg: t('admin_email.msg.empty'),
      };
    }

    if (email.value && !Pattern.email.test(email.value)) {
      bol = false;
      data.email = {
        value: email.value,
        isInvalid: true,
        errorMsg: t('admin_email.msg.incorrect'),
      };
    }

    changeCallback({
      ...data,
    });
    return bol;
  };

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    nextCallback();
  };

  if (!visible) return null;
  return (
    <Form noValidate onSubmit={handleSubmit}>
      <h5>{t('site_information')}</h5>
      <Form.Group controlId="site_name" className="mb-3">
        <Form.Label>{t('site_name.label')}</Form.Label>
        <Form.Control
          required
          value={data.site_name.value}
          isInvalid={data.site_name.isInvalid}
          onChange={(e) => {
            changeCallback({
              site_name: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Control.Feedback type="invalid">
          {data.site_name.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>
      <Form.Group controlId="site_url" className="mb-3">
        <Form.Label>{t('site_url.label')}</Form.Label>
        <Form.Control
          required
          value={data.site_url.value}
          isInvalid={data.site_url.isInvalid}
          onChange={(e) => {
            changeCallback({
              site_url: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Text>{t('site_url.text')}</Form.Text>
        <Form.Control.Feedback type="invalid">
          {data.site_url.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>
      <Form.Group controlId="contact_email" className="mb-3">
        <Form.Label>{t('contact_email.label')}</Form.Label>
        <Form.Control
          required
          type="email"
          value={data.contact_email.value}
          isInvalid={data.contact_email.isInvalid}
          onChange={(e) => {
            changeCallback({
              contact_email: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Text>{t('contact_email.text')}</Form.Text>
        <Form.Control.Feedback type="invalid">
          {data.contact_email.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>

      <Form.Group controlId="login_required" className="mb-3">
        <Form.Label>{t('login_required.label')}</Form.Label>
        <Form.Check
          type="switch"
          id="login_required"
          label={t('login_required.switch')}
          checked={data.login_required.value}
          onChange={(e) => {
            changeCallback({
              login_required: {
                value: e.target.checked,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Text>{t('login_required.text')}</Form.Text>
      </Form.Group>

      <h5>{t('admin_account')}</h5>
      <Form.Group controlId="name" className="mb-3">
        <Form.Label>{t('admin_name.label')}</Form.Label>
        <Form.Control
          required
          value={data.name.value}
          isInvalid={data.name.isInvalid}
          onChange={(e) => {
            changeCallback({
              name: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Control.Feedback type="invalid">
          {data.name.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>

      <Form.Group controlId="password" className="mb-3">
        <Form.Label>{t('admin_password.label')}</Form.Label>
        <Form.Control
          required
          type="password"
          value={data.password.value}
          isInvalid={data.password.isInvalid}
          onChange={(e) => {
            changeCallback({
              password: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Text>{t('admin_password.text')}</Form.Text>
        <Form.Control.Feedback type="invalid">
          {data.password.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>

      <Form.Group controlId="email" className="mb-3">
        <Form.Label>{t('admin_email.label')}</Form.Label>
        <Form.Control
          required
          value={data.email.value}
          isInvalid={data.email.isInvalid}
          onChange={(e) => {
            changeCallback({
              email: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}
        />
        <Form.Text>{t('admin_email.text')}</Form.Text>
        <Form.Control.Feedback type="invalid">
          {data.email.errorMsg}
        </Form.Control.Feedback>
      </Form.Group>

      <div className="d-flex align-items-center justify-content-between">
        <Progress step={4} />
        <Button type="submit">{t('next')}</Button>
      </div>
    </Form>
  );
};

export default Index;

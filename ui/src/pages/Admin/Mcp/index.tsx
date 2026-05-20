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

import { FormEvent, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Form, Button } from 'react-bootstrap';

import { useToast } from '@/hooks';
import { getMcpConfig, saveMcpConfig } from '@/services';

const Mcp = () => {
  const toast = useToast();
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.mcp',
  });
  const [formData, setFormData] = useState({
    enabled: true,
    type: '',
    url: '',
    http_header: '',
  });
  const [isLoading, setIsLoading] = useState(false);

  const handleOnChange = (form) => {
    setFormData({ ...formData, ...form });
  };
  const onSubmit = (evt: FormEvent) => {
    evt.preventDefault();
    evt.stopPropagation();
    saveMcpConfig({ enabled: formData.enabled }).then(() => {
      toast.onShow({
        msg: t('update', { keyPrefix: 'toast' }),
        variant: 'success',
      });
    });
  };

  useEffect(() => {
    getMcpConfig()
      .then((resp) => {
        setIsLoading(false);
        setFormData(resp);
      })
      .catch(() => {
        setIsLoading(false);
      });
  }, []);
  if (isLoading) {
    return null;
  }
  return (
    <>
      <h3 className="mb-4">{t('mcp', { keyPrefix: 'nav_menus' })}</h3>
      <div className="max-w-748">
        <Form onSubmit={onSubmit}>
          <Form.Group className="mb-3" controlId="mcp_server">
            <Form.Label>{t('mcp_server.label')}</Form.Label>
            <Form.Check
              type="switch"
              label={t('mcp_server.switch')}
              checked={formData.enabled}
              onChange={(e) => handleOnChange({ enabled: e.target.checked })}
            />
          </Form.Group>
          {formData.enabled && (
            <>
              <Form.Group className="mb-3" controlId="type">
                <Form.Label>{t('type.label')}</Form.Label>
                <Form.Control type="text" disabled value={formData.type} />
              </Form.Group>
              <Form.Group className="mb-3" controlId="url">
                <Form.Label>{t('url.label')}</Form.Label>
                <Form.Control type="text" disabled value={formData.url} />
              </Form.Group>
              <Form.Group className="mb-3">
                <Form.Label>{t('http_header.label')}</Form.Label>
                <Form.Control
                  type="text"
                  disabled
                  value={formData.http_header}
                />
                <Form.Text className="text-muted">
                  {t('http_header.text')}
                </Form.Text>
              </Form.Group>
            </>
          )}
          <Button variant="primary" type="submit">
            {t('save', { keyPrefix: 'btns' })}
          </Button>
        </Form>
      </div>
    </>
  );
};

export default Mcp;

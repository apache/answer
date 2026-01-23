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

import { useState } from 'react';
import { Modal, Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { handleFormError } from '@/utils';
import { addApiKey, updateApiKey } from '@/services';

const initFormData = {
  description: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
  scope: {
    value: 'read-only',
    isInvalid: false,
    errorMsg: '',
  },
};

const Index = ({ data, visible = false, onClose, callback }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.apikeys.add_or_edit_modal',
  });
  const [formData, setFormData] = useState<any>(initFormData);

  const handleValueChange = (value) => {
    setFormData({
      ...formData,
      ...value,
    });
  };

  const handleAdd = () => {
    const { description, scope } = formData;
    if (!description.value) {
      setFormData({
        ...formData,
        description: {
          ...description,
          isInvalid: true,
          errorMsg: t('description_required'),
        },
      });
      return;
    }
    addApiKey({
      description: description.value,
      scope: scope.value,
    })
      .then((res) => {
        callback('add', res.access_key);
        setFormData(initFormData);
      })
      .catch((error) => {
        const obj = handleFormError(error, formData);
        setFormData({ ...obj });
      });
  };

  const handleEdit = () => {
    const { description } = formData;
    if (!description.value) {
      setFormData({
        ...formData,
        description: {
          ...description,
          isInvalid: true,
          errorMsg: t('description_required'),
        },
      });
      return;
    }
    updateApiKey({
      description: description.value,
      id: data?.id,
    })
      .then(() => {
        callback('edit', null);
        setFormData(initFormData);
      })
      .catch((error) => {
        const obj = handleFormError(error, formData);
        setFormData({ ...obj });
      });
  };

  const handleSubmit = () => {
    if (data?.id) {
      handleEdit();
      return;
    }
    handleAdd();
  };

  const closeModal = () => {
    setFormData(initFormData);
    onClose(false, null);
  };
  return (
    <Modal show={visible} onHide={closeModal}>
      <Modal.Header closeButton>
        {data?.id ? t('edit_title') : t('add_title')}
      </Modal.Header>
      <Modal.Body>
        <Form>
          <Form.Group controlId="description" className="mb-3">
            <Form.Label>{t('description')}</Form.Label>
            <Form.Control
              type="text"
              isInvalid={formData.description.isInvalid}
              value={formData.description.value}
              onChange={(e) => {
                handleValueChange({
                  description: {
                    value: e.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                });
              }}
            />
            <Form.Control.Feedback type="invalid">
              {formData.description.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          {!data?.id && visible && (
            <Form.Group controlId="scope" className="mb-3">
              <Form.Label>{t('scope')}</Form.Label>
              <Form.Select
                isInvalid={formData.scope.isInvalid}
                value={formData.scope.value}
                onChange={(e) => {
                  handleValueChange({
                    scope: {
                      value: e.target.value,
                      errorMsg: '',
                      isInvalid: false,
                    },
                  });
                }}>
                <option value="read-only">{t('read-only')}</option>
                <option value="global">{t('global')}</option>
              </Form.Select>
              <Form.Control.Feedback type="invalid">
                {formData.scope.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
          )}
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={closeModal}>
          {t('cancel', { keyPrefix: 'btns' })}
        </Button>
        <Button type="button" variant="primary" onClick={handleSubmit}>
          {t('submit', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default Index;

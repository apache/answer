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

import { useLayoutEffect, useState, useRef } from 'react';
import { Modal, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import ReactDOM from 'react-dom/client';

import type * as Type from '@/common/interface';
import { SchemaForm, JSONSchema, UISchema, initFormData } from '@/components';
import { handleFormError } from '@/utils';

const div = document.createElement('div');
const root = ReactDOM.createRoot(div);

interface IProps {
  title?: string;
  onConfirm?: (formData: any) => Promise<any>;
}
const useAddUserModal = (props: IProps = {}) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.user_modal',
  });

  const { title = t('title'), onConfirm } = props;
  const [visible, setVisibleState] = useState(false);
  const schema: JSONSchema = {
    title: t('title'),
    required: ['users'],
    properties: {
      users: {
        type: 'string',
        title: t('form.fields.users.label'),
        description: t('form.fields.users.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    users: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 7,
        placeholder: t('form.fields.users.placeholder'),
        className: 'small font-monospace',
      },
    },
  };
  const [formData, setFormData] = useState<Type.FormDataType>(
    initFormData(schema),
  );

  const formRef = useRef<{
    validator: () => Promise<boolean>;
  }>(null);

  const onClose = () => {
    setVisibleState(false);
  };

  const onShow = () => {
    setVisibleState(true);
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    event.stopPropagation();
    const isValid = await formRef.current?.validator();

    if (!isValid) {
      return;
    }

    if (onConfirm instanceof Function) {
      onConfirm({
        users: formData.users.value,
      })
        .then(() => {
          setFormData({
            users: {
              value: '',
              isInvalid: false,
              errorMsg: '',
            },
          });
          onClose();
        })
        .catch((err) => {
          if (err.isError) {
            const data = handleFormError(err, formData);
            setFormData({ ...data });
          }
        });
    }
  };

  const handleOnChange = (data) => {
    setFormData(data);
  };

  useLayoutEffect(() => {
    root.render(
      <Modal show={visible} title={title} onHide={onClose}>
        <Modal.Header closeButton>
          <Modal.Title as="h5">{title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <SchemaForm
            ref={formRef}
            schema={schema}
            uiSchema={uiSchema}
            formData={formData}
            onChange={handleOnChange}
            hiddenSubmit
          />
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => onClose()}>
            {t('btn_cancel')}
          </Button>
          <Button variant="primary" onClick={handleSubmit}>
            {t('btn_submit')}
          </Button>
        </Modal.Footer>
      </Modal>,
    );
  });
  return {
    onClose,
    onShow,
  };
};

export default useAddUserModal;

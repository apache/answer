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

import { Modal, Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

const Index = ({ visible, api_key = '', onClose }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.apikeys.created_modal',
  });

  return (
    <Modal show={visible} onHide={onClose}>
      <Modal.Header closeButton>{t('title')}</Modal.Header>
      <Modal.Body>
        <Form>
          <Form.Group controlId="api_key" className="mb-3">
            <Form.Label>{t('api_key')}</Form.Label>
            <Form.Control
              type="text"
              defaultValue={api_key}
              readOnly
              disabled
            />
          </Form.Group>

          <div className="mb-3">{t('description')}</div>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={onClose}>
          {t('close', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default Index;

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
import { Modal, Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

const DeleteUserModal = ({
  show,
  onClose,
  onDelete,
  count = 1,
  isBulk = false,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.users' });
  const [checkVal, setCheckVal] = useState(false);

  const handleClose = () => {
    onClose();
    setCheckVal(false);
  };

  return (
    <Modal show={show} onHide={handleClose}>
      <Modal.Header closeButton>
        <Modal.Title>
          {t(isBulk ? 'bulk_delete.title' : 'delete_user.title', { count })}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <p>
          {t(isBulk ? 'bulk_delete.content' : 'delete_user.content', {
            count,
          })}
        </p>
        <div className="text-danger mb-2">
          {t('delete_user.remove')} {t('optional', { keyPrefix: 'form' })}
        </div>
        <Form>
          <Form.Group controlId="delete_user">
            <Form.Check type="checkbox" id="delete_user">
              <Form.Check.Input
                type="checkbox"
                checked={checkVal}
                onChange={(e) => {
                  setCheckVal(e.target.checked);
                }}
              />
              <Form.Check.Label htmlFor="delete_user">
                <span>{t('delete_user.label')}</span>
                <br />
                <span className="small text-secondary">
                  {t('delete_user.text')}
                </span>
              </Form.Check.Label>
            </Form.Check>
          </Form.Group>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={handleClose}>
          {t('cancel', { keyPrefix: 'btns' })}
        </Button>
        <Button
          variant="danger"
          onClick={() => {
            onDelete(checkVal);
            setCheckVal(false);
          }}>
          {t('delete', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default DeleteUserModal;

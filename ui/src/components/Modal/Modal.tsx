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

import React, { FC } from 'react';
import { Button, Modal } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

export interface Props {
  id?: string;
  /** header title */
  title?: string;
  children?: React.ReactNode;
  /** visible */
  visible?: boolean;
  centered?: boolean;
  onCancel?: () => void;
  onConfirm?: (event: any) => void;
  cancelText?: string;
  showCancel?: boolean;
  cancelBtnVariant?: string;
  confirmText?: string;
  showConfirm?: boolean;
  confirmBtnDisabled?: boolean;
  confirmBtnVariant?: string;
  /** body style */
  bodyClass?: string;
  scrollable?: boolean;
  className?: string;
}
const Index: FC<Props> = ({
  id = '',
  title = 'title',
  visible = false,
  centered = true,
  onCancel,
  children,
  onConfirm,
  cancelText = '',
  showCancel = true,
  cancelBtnVariant = 'primary',
  confirmText = '',
  showConfirm = true,
  confirmBtnVariant = 'link',
  confirmBtnDisabled = false,
  bodyClass = '',
  scrollable = false,
  className = '',
}) => {
  const { t } = useTranslation();
  return (
    <Modal
      id={id}
      className={className}
      scrollable={scrollable}
      show={visible}
      onHide={onCancel}
      centered={centered}
      fullscreen="sm-down">
      <Modal.Header closeButton>
        <Modal.Title as="h5">
          {title || t('title', { keyPrefix: 'modal_confirm' })}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body className={classNames('text-break', bodyClass)}>
        {children}
      </Modal.Body>
      {(showCancel || showConfirm) && (
        <Modal.Footer>
          {showCancel && (
            <Button variant={cancelBtnVariant} onClick={onCancel}>
              {cancelText === 'close'
                ? t('btns.close')
                : cancelText || t('btns.cancel')}
            </Button>
          )}
          {showConfirm && (
            <Button
              variant={confirmBtnVariant}
              onClick={(event) => {
                onConfirm?.(event);
              }}
              id="ok_button"
              disabled={confirmBtnDisabled}>
              {confirmText === 'OK'
                ? t('btns.ok')
                : confirmText || t('btns.confirm')}
            </Button>
          )}
        </Modal.Footer>
      )}
    </Modal>
  );
};

export default React.memo(Index);

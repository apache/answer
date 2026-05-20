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

import { useLayoutEffect, useState } from 'react';
import { Toast, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import ReactDOM from 'react-dom/client';

import { EXTERNAL_CONTENT_DISPLAY_MODE } from '@/common/constants';
import { Storage } from '@/utils';

const toastPortal = document.createElement('div');
toastPortal.style.position = 'fixed';
toastPortal.style.top = '78px';
toastPortal.style.left = '50%';
toastPortal.style.transform = 'translate(-50%, 0)';
toastPortal.style.maxWidth = '100%';
toastPortal.style.zIndex = '1001';

const setPortalPosition = () => {
  const header = document.querySelector('#header');
  if (header) {
    toastPortal.style.top = `${header.getBoundingClientRect().top + 78}px`;
  }
};
const startHandlePortalPosition = () => {
  setPortalPosition();
  window.addEventListener('scroll', setPortalPosition);
};

const stopHandlePortalPosition = () => {
  setPortalPosition();
  window.removeEventListener('scroll', setPortalPosition);
};

const root = ReactDOM.createRoot(toastPortal);

const useExternalToast = () => {
  const [show, setShow] = useState(false);
  const { t } = useTranslation('translation', { keyPrefix: 'messages' });

  const onClose = () => {
    const parent = document.querySelector('.page-wrap');
    if (parent?.contains(toastPortal)) {
      parent.removeChild(toastPortal);
    }
    stopHandlePortalPosition();
    setShow(false);
  };

  const onShow = () => {
    startHandlePortalPosition();
    setShow(true);
  };

  const showExternalResourceMode = (mode) => {
    if (mode === 'always') {
      Storage.set(EXTERNAL_CONTENT_DISPLAY_MODE, 'always');
    } else {
      Storage.remove(EXTERNAL_CONTENT_DISPLAY_MODE);
    }
    const img = document.querySelectorAll('img');
    img.forEach((i) => {
      if (!i.src && i.dataset.src) {
        i.src = i.dataset.src;
        i.removeAttribute('data-src');
        i.classList.remove('broken');
      }
    });
    onClose();
  };

  useLayoutEffect(() => {
    const parent = document.querySelector('.page-wrap');
    parent?.appendChild(toastPortal);

    root.render(
      <div className="d-flex justify-content-center">
        <Toast
          className="align-items-center border-0"
          bg="warning"
          show={show}
          onClose={onClose}>
          <div className="d-flex">
            <Toast.Body>
              {t('external_content_warning')}
              <div className="d-flex align-items-center">
                <Button
                  variant="link"
                  onClick={() => showExternalResourceMode('below')}
                  className="btn-no-border small link-dark p-0 fw-bold">
                  {t('display_below', { keyPrefix: 'btns' })}
                </Button>
                <span className="mx-1">{t('or', { keyPrefix: 'btns' })}</span>
                <Button
                  variant="link"
                  onClick={() => showExternalResourceMode('always')}
                  className="btn-no-border small link-dark p-0 fw-bold">
                  {t('always_display', { keyPrefix: 'btns' })}
                </Button>
              </div>
            </Toast.Body>
            <button
              className="btn-close me-2 m-auto"
              onClick={onClose}
              data-bs-dismiss="toast"
              aria-label="Close"
            />
          </div>
        </Toast>
      </div>,
    );
  }, [show]);

  return {
    onShow,
  };
};

export default useExternalToast;

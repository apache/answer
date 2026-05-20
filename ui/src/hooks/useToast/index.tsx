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
import { Toast } from 'react-bootstrap';

import ReactDOM from 'react-dom/client';

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

interface Params {
  /** main content */
  msg: string;
  /** theme color */
  variant?: 'warning' | 'success' | 'danger';
}

const useToast = () => {
  const [show, setShow] = useState(false);
  const [data, setData] = useState<Params>({
    msg: '',
    variant: 'warning',
  });

  const onClose = () => {
    const parent = document.querySelector('.page-wrap');
    if (parent?.contains(toastPortal)) {
      parent.removeChild(toastPortal);
    }
    stopHandlePortalPosition();
    setShow(false);
  };

  const onShow = (t: Params) => {
    setData(t);
    startHandlePortalPosition();
    setShow(true);
  };
  useLayoutEffect(() => {
    const parent = document.querySelector('.page-wrap');
    parent?.appendChild(toastPortal);

    root.render(
      <div className="d-flex justify-content-center">
        <Toast
          className="align-items-center border-0"
          delay={5000}
          bg={data.variant || 'warning'}
          show={show}
          autohide
          onClose={onClose}>
          <div className="d-flex">
            <Toast.Body
              dangerouslySetInnerHTML={{ __html: data.msg }}
              className={`${data.variant !== 'warning' ? 'text-white' : ''}`}
            />
            <button
              className={`btn-close me-2 m-auto ${
                data.variant !== 'warning' ? 'btn-close-white' : ''
              }`}
              onClick={onClose}
              data-bs-dismiss="toast"
              aria-label="Close"
            />
          </div>
        </Toast>
      </div>,
    );
  }, [show, data]);

  return {
    onShow,
  };
};

export default useToast;

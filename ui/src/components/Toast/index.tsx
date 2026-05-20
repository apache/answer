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

import { FC, memo } from 'react';
import { Toast } from 'react-bootstrap';

interface IProps {
  /** main content */
  msg: string;
  /** theme color */
  variant?: 'warning' | 'success' | 'danger';
  /** callback click close */
  onClose: () => void;
}

const Index: FC<IProps> = ({ msg, variant = 'warning', onClose }) => {
  return (
    <div
      className="d-flex justify-content-center"
      style={{
        position: 'fixed',
        top: '78px',
        left: 0,
        right: 0,
        margin: 'auto',
        zIndex: 5,
      }}>
      <Toast
        className="align-items-center border-0"
        delay={5000}
        bg={variant}
        show={Boolean(msg)}
        autohide
        onClose={onClose}>
        <div className="d-flex">
          <Toast.Body
            dangerouslySetInnerHTML={{ __html: msg }}
            className={`${variant !== 'warning' ? 'text-white' : ''}`}
          />
          <button
            className={`btn-close me-2 m-auto ${
              variant !== 'warning' ? 'btn-close-white' : ''
            }`}
            onClick={onClose}
            data-bs-dismiss="toast"
            aria-label="Close"
          />
        </div>
      </Toast>
    </div>
  );
};

export default memo(Index);

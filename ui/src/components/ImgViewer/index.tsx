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

import { FC, MouseEvent, ReactNode, useEffect, useRef, useState } from 'react';
import { Modal } from 'react-bootstrap';

import './index.css';
import classnames from 'classnames';

const Index: FC<{
  children: ReactNode;
  className?: classnames.Argument;
}> = ({ children, className }) => {
  const [visible, setVisible] = useState(false);
  const [imgSrc, setImgSrc] = useState('');
  const ref = useRef<HTMLDivElement>(null);

  const onClose = () => {
    setVisible(false);
    setImgSrc('');
  };

  const checkIfInLink = (target) => {
    let ret = false;
    let el = target.parentElement;
    while (el) {
      if (el.nodeName.toLowerCase() === 'a') {
        ret = true;
        break;
      }
      el = el.parentElement;
    }
    return ret;
  };

  const checkClickForImgView = (evt: MouseEvent<HTMLElement>) => {
    const { target } = evt;
    // @ts-ignore
    if (target.nodeName.toLowerCase() !== 'img') {
      return;
    }
    const img = target as HTMLImageElement;
    if (!img.naturalWidth || !img.naturalHeight) {
      img.classList.add('broken');
      return;
    }
    const src = img.currentSrc || img.src;
    if (src && checkIfInLink(img) === false) {
      setImgSrc(src);
      setVisible(true);
    }
  };

  const handleCopy = (text: string, button: Element) => {
    navigator.clipboard.writeText(text).then(() => {
      button.classList.add('copied');
      setTimeout(() => {
        button.classList.remove('copied');
      }, 1000);
    });
  };

  useEffect(() => {
    if (ref.current) {
      const preElements = ref.current.querySelectorAll('pre');
      preElements.forEach((pre) => {
        let button = pre.querySelector('.copy-button');
        if (!button) {
          button = document.createElement('button');
          button.className = 'copy-button';
          button.addEventListener('click', () => {
            if (button) {
              handleCopy(pre.innerText, button);
            }
          });
          pre.appendChild(button);
        }
      });
    }
  }, [children]);

  useEffect(() => {
    return () => {
      onClose();
    };
  }, []);

  return (
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events
    <div
      ref={ref}
      className={classnames('img-viewer', className)}
      onClick={checkClickForImgView}>
      {children}
      <Modal
        show={visible}
        fullscreen
        centered
        scrollable
        contentClassName="bg-transparent"
        onHide={onClose}>
        <Modal.Body onClick={onClose} className="img-viewer p-0 d-flex">
          {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-noninteractive-element-interactions */}
          <img
            className="cursor-zoom-out img-fluid m-auto"
            onClick={(evt) => {
              evt.stopPropagation();
              onClose();
            }}
            src={imgSrc}
            alt={imgSrc}
          />
        </Modal.Body>
      </Modal>
    </div>
  );
};

export default Index;

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

import { useEffect, useRef, useState, memo } from 'react';
import { Button, Form, Modal } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import ToolItem from '../toolItem';
import { Editor } from '../types';

const Link = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const item = {
    label: 'link-45deg',
    keyMap: ['Ctrl-l'],
    tip: `${t('link.text')} (Ctrl+l)`,
  };
  const [visible, setVisible] = useState(false);
  const [currentEditor, setCurrentEditor] = useState<Editor | null>(null);
  const [link, setLink] = useState({
    value: 'https://',
    isInvalid: false,
    errorMsg: '',
  });
  const [name, setName] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (visible && inputRef.current) {
      inputRef.current.setSelectionRange(0, inputRef.current.value.length);
      inputRef.current.focus();
    }
  }, [visible]);

  const addLink = (editor: Editor) => {
    setCurrentEditor(editor);
    const text = editor.getSelection();
    setName({ ...name, value: text });
    setVisible(true);
  };
  const handleClick = () => {
    if (!currentEditor) {
      return;
    }

    if (!link.value) {
      setLink({ ...link, isInvalid: true });
      return;
    }

    currentEditor.insertLink(link.value, name.value || undefined);

    setVisible(false);
    currentEditor.focus();
    setLink({ ...link, value: '' });
    setName({ ...name, value: '' });
  };
  const onHide = () => setVisible(false);
  const onExited = () => {
    currentEditor?.focus();
  };

  return (
    <>
      <ToolItem {...item} onClick={addLink} />
      <Modal
        show={visible}
        onHide={onHide}
        onExited={onExited}
        fullscreen="sm-down">
        <Modal.Header closeButton>
          <h5 className="mb-0">{t('link.add_link')}</h5>
        </Modal.Header>
        <Modal.Body>
          <Form onSubmit={handleClick}>
            <Form.Group controlId="editor.internetSite" className="mb-3">
              <Form.Label>{t('link.form.fields.url.label')}</Form.Label>
              <Form.Control
                ref={inputRef}
                type="text"
                value={link.value}
                onChange={(e) => setLink({ ...link, value: e.target.value })}
                isInvalid={link.isInvalid}
              />
            </Form.Group>

            <Form.Group controlId="editor.internetSiteName" className="mb-3">
              <Form.Label>{`${t('link.form.fields.name.label')} ${t(
                'optional',
                {
                  keyPrefix: 'form',
                },
              )}`}</Form.Label>
              <Form.Control
                type="text"
                value={name.value}
                onChange={(e) => setName({ ...name, value: e.target.value })}
                isInvalid={name.isInvalid}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => setVisible(false)}>
            {t('link.btn_cancel')}
          </Button>
          <Button variant="primary" onClick={handleClick}>
            {t('link.btn_confirm')}
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
};

export default memo(Link);

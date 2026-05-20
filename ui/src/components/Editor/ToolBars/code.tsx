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

import Select from '../Select';
import ToolItem from '../toolItem';
import { Editor } from '../types';

const codeLanguageType = [
  'bash',
  'sh',
  'zsh',
  'c',
  'h',
  'cpp',
  'hpp',
  'c++',
  'h++',
  'cc',
  'hh',
  'cxx',
  'hxx',
  'c-like',
  'cs',
  'csharp',
  'c#',
  'clojure',
  'clj',
  'coffee',
  'coffeescript',
  'cson',
  'iced',
  'css',
  'dart',
  'erl',
  'erlang',
  'go',
  'golang',
  'hs',
  'haskell',
  'html',
  'xml',
  'xsl',
  'xhtml',
  'rss',
  'atom',
  'xjb',
  'xsd',
  'plist',
  'wsf',
  'svg',
  'http',
  'https',
  'ini',
  'toml',
  'java',
  'jsp',
  'js',
  'javascript',
  'jsx',
  'mjs',
  'cjs',
  'json',
  'kotlin',
  'kt',
  'latex',
  'tex',
  'less',
  'lisp',
  'lua',
  'makefile',
  'mk',
  'mak',
  'markdown',
  'md',
  'mkdown',
  'mkd',
  'matlab',
  'objectivec',
  'mm',
  'objc',
  'obj-c',
  'ocaml',
  'ml',
  'pascal',
  'delphi',
  'dpr',
  'dfm',
  'pas',
  'freepascal',
  'lazarus',
  'lpr',
  'lfm',
  'pl',
  'perl',
  'pm',
  'php',
  'php3',
  'php4',
  'php5',
  'php6',
  'php7',
  'php-template',
  'protobuf',
  'py',
  'python',
  'gyp',
  'ipython',
  'r',
  'rb',
  'ruby',
  'gemspec',
  'podspec',
  'thor',
  'irb',
  'rs',
  'rust',
  'scala',
  'scheme',
  'scss',
  'shell',
  'console',
  'sql',
  'swift',
  'typescript',
  'ts',
  'vhdl',
  'vbnet',
  'vb',
  'yaml',
  'yml',
];

const Code = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });

  const item = {
    label: 'code-slash',
    keyMap: ['Ctrl-k'],
    tip: `${t('code.text')} (Ctrl+k)`,
  };

  const [code, setCode] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });
  const [visible, setVisible] = useState(false);
  const [lang, setLang] = useState('');
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const SINGLELINEMAXLENGTH = 40;
  const [currentEditor, setCurrentEditor] = useState<Editor | null>(null);

  const addCode = (editor: Editor) => {
    setCurrentEditor(editor);
    const text = editor.getSelection();

    if (!text) {
      setVisible(true);
      return;
    }
    if (text.length > SINGLELINEMAXLENGTH) {
      editor.insertCodeBlock('', text);
    } else {
      editor.insertCode(text);
    }
    editor.focus();
  };

  useEffect(() => {
    if (visible && inputRef.current) {
      inputRef.current.focus();
    }
  }, [visible]);

  const handleClick = () => {
    if (!currentEditor) {
      return;
    }

    if (!code.value.trim()) {
      setCode({
        ...code,
        errorMsg: t('code.form.fields.code.msg.empty'),
        isInvalid: true,
      });
      return;
    }

    if (
      code.value.split('\n').length > 1 ||
      code.value.length >= SINGLELINEMAXLENGTH
    ) {
      currentEditor.insertCodeBlock(lang || undefined, code.value);
    } else {
      currentEditor.insertCode(code.value);
    }

    setCode({
      value: '',
      isInvalid: false,
      errorMsg: '',
    });
    setLang('');
    setVisible(false);
    currentEditor.focus();
  };
  const onHide = () => setVisible(false);
  const onExited = () => currentEditor?.focus();

  return (
    <ToolItem {...item} onClick={addCode}>
      <Modal
        show={visible}
        onHide={onHide}
        onExited={onExited}
        fullscreen="sm-down">
        <Modal.Header closeButton>
          <h5 className="mb-0">{t('code.add_code')}</h5>
        </Modal.Header>
        <Modal.Body>
          <Form.Group controlId="editor.code" className="mb-3">
            <Form.Label>{t('code.form.fields.code.label')}</Form.Label>
            <Form.Control
              ref={inputRef}
              as="textarea"
              rows={3}
              value={code.value}
              isInvalid={code.isInvalid}
              className="font-monospace"
              style={{ height: '200px' }}
              onChange={(e) => setCode({ ...code, value: e.target.value })}
            />
            {code.isInvalid && (
              <Form.Control.Feedback type="invalid">
                {code.errorMsg}
              </Form.Control.Feedback>
            )}
          </Form.Group>
          <Form.Group controlId="editor.codeLanguageType" className="mb-3">
            <Form.Label>{`${t('code.form.fields.language.label')} ${t(
              'optional',
              {
                keyPrefix: 'form',
              },
            )}`}</Form.Label>
            <Select
              options={codeLanguageType}
              value={lang}
              onChange={(e) => setLang(e.target.value)}
              onSelect={(val) => setLang(val)}
              placeholder={t('code.form.fields.language.placeholder')}
            />
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button
            variant="link"
            onClick={() => {
              setVisible(false);
              setCode({
                value: '',
                isInvalid: false,
                errorMsg: '',
              });
            }}>
            {t('code.btn_cancel')}
          </Button>
          <Button variant="primary" onClick={handleClick}>
            {t('code.btn_confirm')}
          </Button>
        </Modal.Footer>
      </Modal>
    </ToolItem>
  );
};

export default memo(Code);

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

import { useEffect, useState, memo, useContext } from 'react';
import { Button, Form, Modal, Tab, Tabs } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import ToolItem from '../toolItem';
import { EditorContext } from '../EditorContext';
import { Editor } from '../types';
import { useImageUpload } from '../hooks/useImageUpload';

const Image = () => {
  const editor = useContext(EditorContext);
  const [editorState, setEditorState] = useState<Editor | null>(editor);

  // Update editor state when editor context changes
  // This ensures event listeners are re-bound when switching editor modes
  useEffect(() => {
    if (editor) {
      setEditorState(editor);
    }
  }, [editor]);
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const { verifyImageSize, uploadFiles } = useImageUpload();

  const loadingText = `![${t('image.uploading')}...]()`;

  const item = {
    label: 'image-fill',
    keyMap: ['Ctrl-g'],
    tip: `${t('image.text')} (Ctrl+G)`,
  };
  const [currentTab, setCurrentTab] = useState('localImage');
  const [visible, setVisible] = useState(false);
  const [link, setLink] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
    type: '',
  });

  const [imageName, setImageName] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });

  function dragenter(e) {
    e.stopPropagation();
    e.preventDefault();
  }

  function dragover(e) {
    e.stopPropagation();
    e.preventDefault();
  }
  const drop = async (e) => {
    const fileList = e.dataTransfer.files;
    const bool = verifyImageSize(fileList);

    if (!bool) {
      return;
    }

    if (!editorState) {
      return;
    }
    const startPos = editorState.getCursor();

    const endPos = { ...startPos, ch: startPos.ch + loadingText.length };

    editorState.replaceSelection(loadingText);
    editorState.setReadOnly(true);
    const urls = await uploadFiles(fileList)
      .catch(() => {
        editorState.replaceRange('', startPos, endPos);
      })
      .finally(() => {
        editorState?.setReadOnly(false);
        editorState?.focus();
      });

    const text: string[] = [];
    if (Array.isArray(urls)) {
      urls.forEach(({ name, url, type }) => {
        if (name && url) {
          text.push(`${type === 'post' ? '!' : ''}[${name}](${url})`);
        }
      });
    }
    if (text.length) {
      editorState.replaceRange(text.join('\n'), startPos, endPos);
    } else {
      editorState?.replaceRange('', startPos, endPos);
    }
  };

  const paste = async (event) => {
    const clipboard = event.clipboardData;

    const bool = verifyImageSize(clipboard.files);

    if (bool) {
      event.preventDefault();
      if (!editorState) {
        return;
      }
      const startPos = editorState.getCursor();
      const endPos = { ...startPos, ch: startPos.ch + loadingText.length };

      editorState?.replaceSelection(loadingText);
      editorState?.setReadOnly(true);
      uploadFiles(clipboard.files)
        .then((urls) => {
          const text = urls.map(({ name, url, type }) => {
            return `${type === 'post' ? '!' : ''}[${name}](${url})`;
          });

          editorState.replaceRange(text.join('\n'), startPos, endPos);
        })
        .catch(() => {
          editorState.replaceRange('', startPos, endPos);
        })
        .finally(() => {
          editorState?.setReadOnly(false);
          editorState?.focus();
        });

      return;
    }

    const htmlStr = clipboard.getData('text/html');
    const imgRegex = /<img([\s\S]*?) src\s*=\s*(['"])([\s\S]*?)\2([^>]*)>/;

    if (!htmlStr.match(imgRegex)) {
      return;
    }
    event.preventDefault();
    const parser = new DOMParser();
    const doc = parser.parseFromString(htmlStr, 'text/html');
    const { body } = doc;

    let markdownText = '';

    function traverse(node) {
      if (node.nodeType === Node.TEXT_NODE) {
        // text node
        markdownText += node.textContent;
      } else if (node.nodeType === Node.ELEMENT_NODE) {
        // element node
        const tagName = node.tagName.toLowerCase();

        if (tagName === 'img') {
          // img node
          const src = node.getAttribute('src');
          const alt = node.getAttribute('alt') || t('image.text');
          markdownText += `![${alt}](${src})`;
        } else if (tagName === 'br') {
          // br node
          markdownText += '\n';
        } else {
          for (let i = 0; i < node.childNodes.length; i += 1) {
            traverse(node.childNodes[i]);
          }
        }

        const blockLevelElements = [
          'p',
          'div',
          'h1',
          'h2',
          'h3',
          'h4',
          'h5',
          'h6',
          'ul',
          'ol',
          'li',
          'blockquote',
          'pre',
          'table',
          'thead',
          'tbody',
          'tr',
          'th',
          'td',
        ];
        if (blockLevelElements.includes(tagName)) {
          markdownText += '\n\n';
        }
      }
    }

    traverse(body);

    markdownText = markdownText.replace(/[\n\s]+/g, (match) => {
      return match.length > 1 ? '\n\n' : match;
    });

    if (editorState) {
      editorState.replaceSelection(markdownText);
    }
  };
  const handleClick = () => {
    if (!link.value) {
      setLink({ ...link, isInvalid: true });
      return;
    }
    setLink({ ...link, type: '' });

    if (editorState) {
      editorState.insertImage(link.value, imageName.value || undefined);
    }

    setVisible(false);

    editorState?.focus();
    setLink({ ...link, value: '' });
    setImageName({ ...imageName, value: '' });
  };
  useEffect(() => {
    if (!editorState) {
      return undefined;
    }

    editorState.on('dragenter', dragenter);
    editorState.on('dragover', dragover);
    editorState.on('drop', drop);
    editorState.on('paste', paste);

    return () => {
      editorState.off('dragenter', dragenter);
      editorState.off('dragover', dragover);
      editorState.off('drop', drop);
      editorState.off('paste', paste);
    };
  }, [editorState]);

  useEffect(() => {
    if (link.value && link.type === 'drop') {
      handleClick();
    }
  }, [link.value]);

  const addLink = (editorInstance: Editor) => {
    setEditorState(editorInstance);
    const text = editorInstance?.getSelection();

    setImageName({ ...imageName, value: text });

    setVisible(true);
  };

  const { uploadSingleFile } = useImageUpload();

  const onUpload = async (e) => {
    if (!editor) {
      return;
    }
    const files = e.target?.files || [];
    const bool = verifyImageSize(files);

    if (!bool) {
      return;
    }

    uploadSingleFile(e.target.files[0]).then((url) => {
      setLink({ ...link, value: url });
      setImageName({ ...imageName, value: files[0].name });
    });
  };

  const onHide = () => setVisible(false);
  const onExited = () => editor?.focus();

  const handleSelect = (tab) => {
    setCurrentTab(tab);
  };
  return (
    <ToolItem {...item} onClick={addLink}>
      <Modal
        show={visible}
        onHide={onHide}
        onExited={onExited}
        fullscreen="sm-down">
        <Modal.Header closeButton>
          <h5 className="mb-0">{t('image.add_image')}</h5>
        </Modal.Header>
        <Modal.Body>
          <Tabs onSelect={handleSelect}>
            <Tab eventKey="localImage" title={t('image.tab_image')}>
              <Form className="mt-3" onSubmit={handleClick}>
                <Form.Group controlId="editor.imgLink" className="mb-3">
                  <Form.Label>
                    {t('image.form_image.fields.file.label')}
                  </Form.Label>
                  <Form.Control
                    type="file"
                    onChange={onUpload}
                    isInvalid={currentTab === 'localImage' && link.isInvalid}
                    accept="image/*"
                  />

                  <Form.Control.Feedback type="invalid">
                    {t('image.form_image.fields.file.msg.empty')}
                  </Form.Control.Feedback>
                </Form.Group>

                <Form.Group controlId="editor.imgDescription" className="mb-3">
                  <Form.Label>
                    {`${t('image.form_image.fields.desc.label')} ${t(
                      'optional',
                      {
                        keyPrefix: 'form',
                      },
                    )}`}
                  </Form.Label>
                  <Form.Control
                    type="text"
                    value={imageName.value}
                    onChange={(e) =>
                      setImageName({ ...imageName, value: e.target.value })
                    }
                    isInvalid={imageName.isInvalid}
                  />
                </Form.Group>
              </Form>
            </Tab>
            <Tab eventKey="remoteImage" title={t('image.tab_url')}>
              <Form className="mt-3" onSubmit={handleClick}>
                <Form.Group controlId="editor.imgUrl" className="mb-3">
                  <Form.Label>
                    {t('image.form_url.fields.url.label')}
                  </Form.Label>
                  <Form.Control
                    type="text"
                    value={link.value}
                    onChange={(e) =>
                      setLink({ ...link, value: e.target.value })
                    }
                    isInvalid={currentTab === 'remoteImage' && link.isInvalid}
                  />
                  <Form.Control.Feedback type="invalid">
                    {t('image.form_url.fields.url.msg.empty')}
                  </Form.Control.Feedback>
                </Form.Group>

                <Form.Group controlId="editor.imgName" className="mb-3">
                  <Form.Label>
                    {`${t('image.form_url.fields.name.label')} ${t('optional', {
                      keyPrefix: 'form',
                    })}`}
                  </Form.Label>
                  <Form.Control
                    type="text"
                    value={imageName.value}
                    onChange={(e) =>
                      setImageName({ ...imageName, value: e.target.value })
                    }
                    isInvalid={imageName.isInvalid}
                  />
                </Form.Group>
              </Form>
            </Tab>
          </Tabs>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => setVisible(false)}>
            {t('image.btn_cancel')}
          </Button>
          <Button variant="primary" onClick={handleClick}>
            {t('image.btn_confirm')}
          </Button>
        </Modal.Footer>
      </Modal>
    </ToolItem>
  );
};

export default memo(Image);

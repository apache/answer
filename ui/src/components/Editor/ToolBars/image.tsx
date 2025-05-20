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

import { useEffect, useState, memo } from 'react';
import { Button, Form, Modal, Tab, Tabs } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Modal as AnswerModal } from '@/components';
import ToolItem from '../toolItem';
import { IEditorContext, Editor } from '../types';
import { uploadImage } from '@/services';
import { writeSettingStore } from '@/stores';

let context: IEditorContext;
const Image = ({ editorInstance }) => {
  const [editor, setEditor] = useState<Editor>(editorInstance);
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const {
    max_image_size = 4,
    max_attachment_size = 8,
    authorized_image_extensions = [],
    authorized_attachment_extensions = [],
  } = writeSettingStore((state) => state.write);

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

  const verifyImageSize = (files: FileList) => {
    if (files.length === 0) {
      return false;
    }

    /**
     * When allowing attachments to be uploaded, verification logic for attachment information has been added. In order to avoid abnormal judgment caused by the order of drag and drop upload, the drag and drop upload verification of attachments and the drag and drop upload of images are put together.
     *
     */
    const canUploadAttachment = authorized_attachment_extensions.length > 0;
    const allowedAllType = [
      ...authorized_image_extensions,
      ...authorized_attachment_extensions,
    ];
    const unSupportFiles = Array.from(files).filter((file) => {
      const fileName = file.name.toLowerCase();
      return canUploadAttachment
        ? !allowedAllType.find((v) => fileName.endsWith(v))
        : file.type.indexOf('image') === -1;
    });

    if (unSupportFiles.length > 0) {
      AnswerModal.confirm({
        content: canUploadAttachment
          ? t('file.not_supported', { file_type: allowedAllType.join(', ') })
          : t('image.form_image.fields.file.msg.only_image'),
        showCancel: false,
      });
      return false;
    }

    const otherFiles = Array.from(files).filter((file) => {
      return file.type.indexOf('image') === -1;
    });

    if (canUploadAttachment && otherFiles.length > 0) {
      const attachmentOverSizeFiles = otherFiles.filter(
        (file) => file.size / 1024 / 1024 > max_attachment_size,
      );
      if (attachmentOverSizeFiles.length > 0) {
        AnswerModal.confirm({
          content: t('file.max_size', { size: max_attachment_size }),
          showCancel: false,
        });
        return false;
      }
    }

    const imageFiles = Array.from(files).filter(
      (file) => file.type.indexOf('image') > -1,
    );
    const oversizedImages = imageFiles.filter(
      (file) => file.size / 1024 / 1024 > max_image_size,
    );
    if (oversizedImages.length > 0) {
      AnswerModal.confirm({
        content: t('image.form_image.fields.file.msg.max_size', {
          size: max_image_size,
        }),
        showCancel: false,
      });
      return false;
    }

    return true;
  };

  const upload = (
    files: FileList,
  ): Promise<{ url: string; name: string; type: string }[]> => {
    const promises = Array.from(files).map(async (file) => {
      const type = file.type.indexOf('image') > -1 ? 'post' : 'post_attachment';
      const url = await uploadImage({ file, type });

      return {
        name: file.name,
        url,
        type,
      };
    });

    return Promise.all(promises);
  };
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

    const startPos = editor.getCursor();

    const endPos = { ...startPos, ch: startPos.ch + loadingText.length };

    editor.replaceSelection(loadingText);
    editor.setReadOnly(true);
    const urls = await upload(fileList)
      .catch(() => {
        editor.replaceRange('', startPos, endPos);
      })
      .finally(() => {
        editor.setReadOnly(false);
        editor.focus();
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
      editor.replaceRange(text.join('\n'), startPos, endPos);
    } else {
      editor.replaceRange('', startPos, endPos);
    }
  };

  const paste = async (event) => {
    const clipboard = event.clipboardData;

    const bool = verifyImageSize(clipboard.files);

    if (bool) {
      event.preventDefault();
      const startPos = editor.getCursor();
      const endPos = { ...startPos, ch: startPos.ch + loadingText.length };

      editor.replaceSelection(loadingText);
      editor.setReadOnly(true);
      upload(clipboard.files)
        .then((urls) => {
          const text = urls.map(({ name, url, type }) => {
            return `${type === 'post' ? '!' : ''}[${name}](${url})`;
          });

          editor.replaceRange(text.join('\n'), startPos, endPos);
        })
        .catch(() => {
          editor.replaceRange('', startPos, endPos);
        })
        .finally(() => {
          editor.setReadOnly(false);
          editor.focus();
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

    editor.replaceSelection(markdownText);
  };
  const handleClick = () => {
    if (!link.value) {
      setLink({ ...link, isInvalid: true });
      return;
    }
    setLink({ ...link, type: '' });

    const text = `![${imageName.value}](${link.value})`;

    editor.replaceSelection(text);

    setVisible(false);

    editor.focus();
    setLink({ ...link, value: '' });
    setImageName({ ...imageName, value: '' });
  };
  useEffect(() => {
    editor?.on('dragenter', dragenter);
    editor?.on('dragover', dragover);
    editor?.on('drop', drop);
    editor?.on('paste', paste);
    return () => {
      editor?.off('dragenter', dragenter);
      editor?.off('dragover', dragover);
      editor?.off('drop', drop);
      editor?.off('paste', paste);
    };
  }, [editor]);

  useEffect(() => {
    if (link.value && link.type === 'drop') {
      handleClick();
    }
  }, [link.value]);

  const addLink = (ctx) => {
    context = ctx;
    setEditor(context.editor);
    const text = context.editor?.getSelection();

    setImageName({ ...imageName, value: text });

    setVisible(true);
  };

  const onUpload = async (e) => {
    if (!editor) {
      return;
    }
    const files = e.target?.files || [];
    const bool = verifyImageSize(files);

    if (!bool) {
      return;
    }

    uploadImage({ file: e.target.files[0], type: 'post' }).then((url) => {
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

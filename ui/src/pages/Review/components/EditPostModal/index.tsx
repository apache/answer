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

import { FC, useState, useEffect } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import { putFlagReviewAction } from '@/services';
import { useCaptchaModal, usePageUsers } from '@/hooks';
import { Editor, TagSelector, Mentions, TextArea } from '@/components';
import {
  // matchedUsers,
  parseUserInfo,
  handleFormError,
  parseEditMentionUser,
} from '@/utils';
import type * as Type from '@/common/interface';

import './index.scss';

interface Props {
  originalData: {
    id: string;
    flag_id: string;
    question_id?: string;
    answer_id?: string;
    title: string;
    content: string;
    tags: Type.Tag[];
  };
  objectType: Type.FlagReviewItem['object_type'] | '';
  visible: boolean;
  handleClose: () => void;
  callback?: () => void;
}

interface FormDataItem {
  title: Type.FormValue<string>;
  tags: Type.FormValue<Type.Tag[]>;
  content: Type.FormValue<string>;
}

const initFormData = {
  title: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
  tags: {
    value: [],
    isInvalid: false,
    errorMsg: '',
  },
  content: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
};

const Index: FC<Props> = ({
  originalData,
  visible = false,
  objectType,
  handleClose,
  callback,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ask' });
  const [formData, setFormData] = useState<FormDataItem>(initFormData);
  const [focusEditor, setFocusEditor] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const pageUsers = usePageUsers();

  const editCaptcha = useCaptchaModal('edit');

  const onClose = (bol) => {
    if (bol) {
      callback?.();
    }
    setFormData(initFormData);
    handleClose();
    setLoaded(false);
  };

  const handleInput = (data: Partial<FormDataItem>) => {
    if (!loaded) {
      return;
    }
    setFormData({
      ...formData,
      ...data,
    });
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const { title, tags, content } = formData;
    if (objectType === 'question') {
      if (!title.value) {
        bol = false;
        formData.title = {
          value: title.value,
          isInvalid: true,
          errorMsg: t('form.fields.title.msg.empty', {
            keyPrefix: 'ask',
          }),
        };
      }

      if (!tags.value.length) {
        bol = false;
        formData.tags = {
          value: tags.value,
          isInvalid: true,
          errorMsg: t('form.fields.tags.msg.empty', {
            keyPrefix: 'ask',
          }),
        };
      }
    }

    if (!content.value || Array.from(content.value.trim()).length < 6) {
      bol = false;
      formData.content = {
        value: content.value,
        isInvalid: true,
        errorMsg: t('form.fields.answer.feedback.characters', {
          keyPrefix: 'edit_answer',
        }),
      };
    } else {
      formData.content = {
        value: content.value,
        isInvalid: false,
        errorMsg: '',
      };
    }

    setFormData({
      ...formData,
    });
    return bol;
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    event.stopPropagation();

    if (!checkValidated()) {
      return;
    }

    const params: Type.PutFlagReviewParams = {
      title: formData.title.value,
      content: formData.content.value,
      tags: formData.tags.value,
      operation_type: 'edit_post',
      flag_id: originalData.flag_id,
    };
    if (objectType === 'answer') {
      delete params.title;
      delete params.tags;
    }
    if (objectType === 'comment') {
      const { value } = formData.content;
      // const users = matchedUsers(value);
      // const userNames = unionBy(users.map((user) => user.userName));
      const commentMarkDown = parseUserInfo(value);

      // params.mention_username_list = userNames;
      params.content = commentMarkDown;

      delete params.title;
      delete params.tags;
    }

    editCaptcha.check(() => {
      if (objectType === 'question') {
        const imgCode = editCaptcha.getCaptcha();
        if (imgCode.verify) {
          params.captcha_code = imgCode.captcha_code;
          params.captcha_id = imgCode.captcha_id;
        }
      }
      putFlagReviewAction(params)
        .then(async () => {
          await editCaptcha.close();
          onClose(true);
        })
        .catch((err) => {
          if (err.isError) {
            editCaptcha.handleCaptchaError(err.list);
            const data = handleFormError(err, formData);
            setFormData({ ...data });
          }
        });
    });
  };

  const handleSelected = (val) => {
    if (!loaded) {
      return;
    }
    setFormData({
      ...formData,
      content: {
        value: val,
        errorMsg: '',
        isInvalid: false,
      },
    });
  };

  useEffect(() => {
    if (!visible) {
      return;
    }

    formData.title.value = originalData.title;
    formData.content.value = originalData.content;
    formData.tags.value = originalData.tags.map((item) => {
      return {
        ...item,
        parsed_text: '',
        original_text: '',
      };
    });
    setFormData({ ...formData });
    setLoaded(true);
  }, [visible]);

  return (
    <Modal
      show={visible}
      onHide={() => onClose(false)}
      className="w-100"
      dialogClassName="edit-post-modal">
      <Modal.Header closeButton>
        <Modal.Title>
          {t('edit_post', { keyPrefix: 'page_review' })}
        </Modal.Title>
      </Modal.Header>
      <Form noValidate onSubmit={handleSubmit}>
        <Modal.Body>
          {objectType === 'question' && (
            <Form.Group controlId="title" className="mb-3">
              <Form.Label>{t('form.fields.title.label')}</Form.Label>
              <Form.Control
                type="text"
                value={formData.title.value}
                isInvalid={formData.title.isInvalid}
                onChange={(e) => {
                  handleInput({
                    title: {
                      value: e.target.value,
                      isInvalid: false,
                      errorMsg: '',
                    },
                  });
                }}
                placeholder={t('form.fields.title.placeholder')}
                autoFocus
                contentEditable
              />

              <Form.Control.Feedback type="invalid">
                {formData.title.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
          )}

          {objectType !== 'comment' && (
            <Form.Group controlId="body">
              <Form.Label>
                {objectType === 'question'
                  ? t('form.fields.body.label')
                  : t('form.fields.answer.label')}
              </Form.Label>
              <Form.Control
                defaultValue={formData.content.value}
                isInvalid={formData.content.isInvalid}
                hidden
              />
              <Editor
                value={formData.content.value}
                onChange={(value) => {
                  handleInput({
                    content: { value, errorMsg: '', isInvalid: false },
                  });
                }}
                className={classNames(
                  'form-control p-0',
                  focusEditor ? 'focus' : '',
                )}
                onFocus={() => {
                  setFocusEditor(true);
                }}
                onBlur={() => {
                  setFocusEditor(false);
                }}
              />
              <Form.Control.Feedback type="invalid">
                {formData.content.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
          )}

          {objectType === 'question' && (
            <Form.Group controlId="tags" className="my-3">
              <Form.Label>{t('form.fields.tags.label')}</Form.Label>
              <Form.Control
                defaultValue={JSON.stringify(formData.tags.value)}
                isInvalid={formData.tags.isInvalid}
                hidden
              />
              <TagSelector
                value={formData.tags.value}
                onChange={(value) => {
                  handleInput({
                    tags: { value, errorMsg: '', isInvalid: false },
                  });
                }}
                showRequiredTag
                maxTagLength={5}
              />
              <Form.Control.Feedback type="invalid">
                {formData.tags.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
          )}

          {objectType === 'comment' && (
            <div className="w-100">
              <div
                className={classNames('custom-form-control', {
                  'is-invalid': formData.content.isInvalid,
                })}>
                <Form.Label>Comment</Form.Label>
                <Mentions
                  pageUsers={pageUsers.getUsers()}
                  onSelected={handleSelected}>
                  <TextArea
                    size="sm"
                    rows={4}
                    value={parseEditMentionUser(formData.content.value)}
                    onChange={(e) => {
                      handleInput({
                        content: {
                          value: e.target.value,
                          errorMsg: '',
                          isInvalid: false,
                        },
                      });
                    }}
                  />
                </Mentions>
              </div>
              <Form.Control.Feedback type="invalid">
                {formData.content.errorMsg}
              </Form.Control.Feedback>
            </div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => onClose(false)}>
            {t('close', { keyPrefix: 'btns' })}
          </Button>
          <Button variant="primary" type="submit">
            {t('submit', { keyPrefix: 'btns' })}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default Index;

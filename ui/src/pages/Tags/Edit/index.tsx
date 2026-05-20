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

import React, { useState, useRef, useEffect } from 'react';
import { Row, Col, Form, Button, Card } from 'react-bootstrap';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';
import classNames from 'classnames';

import { usePageTags, usePromptWithUnload } from '@/hooks';
import { Editor, EditorRef } from '@/components';
import { loggedUserInfoStore } from '@/stores';
import type * as Type from '@/common/interface';
import { TAG_SLUG_NAME_MAX_LENGTH } from '@/common/constants';
import { useTagInfo, modifyTag, useQueryRevisions } from '@/services';

interface FormDataItem {
  displayName: Type.FormValue<string>;
  slugName: Type.FormValue<string>;
  description: Type.FormValue<string>;
  editSummary: Type.FormValue<string>;
}
const initFormData = {
  displayName: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
  slugName: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
  description: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
  editSummary: {
    value: '',
    isInvalid: false,
    errorMsg: '',
  },
};

const Index = () => {
  const { role_id = 1 } = loggedUserInfoStore((state) => state.user);

  const { tagId } = useParams();
  const navigate = useNavigate();
  const { t } = useTranslation('translation', { keyPrefix: 'edit_tag' });
  const [focusType, setForceType] = useState('');

  const { data } = useTagInfo({ id: tagId });
  const { data: revisions = [] } = useQueryRevisions(data?.tag_id);
  const [formData, setFormData] = useState<FormDataItem>(initFormData);
  const [immData, setImmData] = useState(initFormData);
  const [contentChanged, setContentChanged] = useState(false);

  const editorRef = useRef<EditorRef>({
    getHtml: () => '',
  });

  usePromptWithUnload({
    when: contentChanged,
  });

  useEffect(() => {
    initFormData.displayName.value = data?.display_name || '';
    initFormData.slugName.value = data?.slug_name || '';
    initFormData.description.value = data?.original_text || '';
    setFormData(initFormData);
    setImmData(initFormData);
  }, [data]);

  useEffect(() => {
    const { displayName, slugName, description, editSummary } = formData;
    const {
      displayName: display_name,
      slugName: slug_name,
      description: original_text,
    } = immData;

    if (
      display_name.value !== displayName.value ||
      slug_name.value !== slugName.value ||
      original_text.value !== description.value ||
      editSummary.value
    ) {
      setContentChanged(true);
    } else {
      setContentChanged(false);
    }
  }, [
    formData.displayName.value,
    formData.slugName.value,
    formData.description.value,
    formData.editSummary.value,
  ]);

  const handleDescriptionChange = (value: string) =>
    setFormData((prev) => ({
      ...prev,
      description: { value, isInvalid: false, errorMsg: '' },
    }));

  const checkValidated = (): boolean => {
    let bol = true;
    const { displayName, slugName } = formData;

    if (!displayName.value) {
      bol = false;
      formData.displayName = {
        value: '',
        isInvalid: true,
        errorMsg: t('form.fields.display_name.msg.empty', {
          keyPrefix: 'tag_modal',
        }),
      };
    } else if (displayName.value.length > TAG_SLUG_NAME_MAX_LENGTH) {
      bol = false;
      formData.displayName = {
        value: displayName.value,
        isInvalid: true,
        errorMsg: t('form.fields.display_name.msg.range', {
          keyPrefix: 'tag_modal',
        }),
      };
    } else {
      formData.displayName = {
        value: displayName.value,
        isInvalid: false,
        errorMsg: '',
      };
    }

    if (!slugName.value) {
      bol = false;
      formData.slugName = {
        value: '',
        isInvalid: true,
        errorMsg: t('form.fields.slug_name.msg.empty', {
          keyPrefix: 'tag_modal',
        }),
      };
    } else if (slugName.value.length > TAG_SLUG_NAME_MAX_LENGTH) {
      bol = false;
      formData.slugName = {
        value: slugName.value,
        isInvalid: true,
        errorMsg: t('form.fields.slug_name.msg.range', {
          keyPrefix: 'tag_modal',
        }),
      };
    } else {
      formData.slugName = {
        value: slugName.value,
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
    setContentChanged(false);

    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }

    const params = {
      display_name: formData.displayName.value,
      slug_name: formData.slugName.value,
      original_text: formData.description.value,
      parsed_text: editorRef.current.getHtml(),
      tag_id: data?.tag_id,
      edit_summary: formData.editSummary.value,
    };
    modifyTag(params).then((res) => {
      navigate(`/tags/${encodeURIComponent(formData.slugName.value)}/info`, {
        replace: true,
        state: { isReview: res.wait_for_review },
      });
    });
  };

  const handleSelectedRevision = (e) => {
    const index = e.target.value;
    const revision = revisions[index];
    formData.description.value = revision.content.original_text;
    formData.displayName.value = revision.content.display_name;
    formData.slugName.value = revision.content.slug_name;
    setImmData({ ...formData });
    setFormData({ ...formData });
  };

  const handleDisplayNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      displayName: { ...formData.displayName, value: e.currentTarget.value },
    });
  };

  const handleEditSummaryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      editSummary: { ...formData.editSummary, value: e.currentTarget.value },
    });
  };

  const handleSlugNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      slugName: { ...formData.slugName, value: e.currentTarget.value },
    });
  };

  const backPage = () => {
    navigate(-1);
  };
  usePageTags({
    title: t('edit_tag', { keyPrefix: 'page_title' }),
  });
  return (
    <div className="pt-4 mb-5">
      <h3 className="mb-4">{t('title')}</h3>
      <Row>
        <Col className="page-main flex-auto">
          <Form noValidate onSubmit={handleSubmit}>
            <Form.Group controlId="revision" className="mb-3">
              <Form.Label>
                {t('form.fields.revision.label', { keyPrefix: 'tag_modal' })}
              </Form.Label>
              <Form.Select onChange={handleSelectedRevision}>
                {revisions.map(({ create_at, reason, user_info }, index) => {
                  const date = dayjs(create_at * 1000)
                    .tz()
                    .format(t('long_date_with_time', { keyPrefix: 'dates' }));
                  return (
                    <option key={`${create_at}`} value={index}>
                      {`${date} - ${user_info.display_name} - ${
                        reason ||
                        (index === revisions.length - 1
                          ? t('default_first_reason')
                          : t('default_reason'))
                      }`}
                    </option>
                  );
                })}
              </Form.Select>
            </Form.Group>
            <Form.Group controlId="display_name" className="mb-3">
              <Form.Label>
                {t('form.fields.display_name.label', {
                  keyPrefix: 'tag_modal',
                })}
              </Form.Label>
              <Form.Control
                value={formData.displayName.value}
                isInvalid={formData.displayName.isInvalid}
                disabled={role_id !== 2 && role_id !== 3}
                onChange={handleDisplayNameChange}
              />

              <Form.Control.Feedback type="invalid">
                {formData.displayName.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
            <Form.Group controlId="slug_name" className="mb-3">
              <Form.Label>
                {t('form.fields.slug_name.label', { keyPrefix: 'tag_modal' })}
              </Form.Label>
              <Form.Control
                value={formData.slugName.value}
                isInvalid={formData.slugName.isInvalid}
                disabled={role_id !== 2 && role_id !== 3}
                onChange={handleSlugNameChange}
              />
              <Form.Text as="div">
                {t('form.fields.slug_name.desc', { keyPrefix: 'tag_modal' })}
              </Form.Text>
              <Form.Control.Feedback type="invalid">
                {formData.slugName.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>

            <Form.Group controlId="description" className="mt-4">
              <Form.Label>
                {t('form.fields.desc.label', { keyPrefix: 'tag_modal' })}
              </Form.Label>
              <Editor
                value={formData.description.value}
                onChange={handleDescriptionChange}
                className={classNames(
                  'form-control p-0',
                  focusType === 'description' && 'focus',
                )}
                onFocus={() => {
                  setForceType('description');
                }}
                onBlur={() => {
                  setForceType('');
                }}
                ref={editorRef}
              />
              <Form.Control
                value={formData.description.value}
                type="text"
                isInvalid={formData.description.isInvalid}
                readOnly
                hidden
              />
              <Form.Control.Feedback type="invalid">
                {formData.description.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
            <Form.Group controlId="edit_summary" className="my-3">
              <Form.Label>
                {t('form.fields.edit_summary.label', {
                  keyPrefix: 'tag_modal',
                })}
              </Form.Label>
              <Form.Control
                type="text"
                defaultValue={formData.editSummary.value}
                isInvalid={formData.editSummary.isInvalid}
                onChange={handleEditSummaryChange}
                placeholder={t('form.fields.edit_summary.placeholder', {
                  keyPrefix: 'tag_modal',
                })}
              />
              <Form.Control.Feedback type="invalid">
                {formData.editSummary.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>

            <div className="mt-3">
              <Button type="submit">{t('btn_save_edits')}</Button>
              <Button variant="link" className="ms-2" onClick={backPage}>
                {t('btn_cancel')}
              </Button>
            </div>
          </Form>
        </Col>
        <Col className="page-right-side mt-4 mt-xl-0">
          <Card>
            <Card.Header>
              {t('title', { keyPrefix: 'how_to_format' })}
            </Card.Header>
            <Card.Body
              className="fmt small"
              dangerouslySetInnerHTML={{
                __html: t('desc', { keyPrefix: 'how_to_format' }),
              }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Index;

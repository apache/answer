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
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';
import classNames from 'classnames';

import { handleFormError } from '@/utils';
import { usePageTags, usePromptWithUnload } from '@/hooks';
import { useCaptchaPlugin, useRenderHtmlPlugin } from '@/utils/pluginKit';
import { pathFactory } from '@/router/pathFactory';
import { Editor, EditorRef, Icon, htmlRender } from '@/components';
import type * as Type from '@/common/interface';
import {
  useQueryAnswerInfo,
  modifyAnswer,
  useQueryRevisions,
} from '@/services';

import './index.scss';

interface FormDataItem {
  content: Type.FormValue<string>;
  description: Type.FormValue<string>;
}

const Index = () => {
  const { aid = '', qid = '' } = useParams();
  const [focusType, setForceType] = useState('');

  const { t } = useTranslation('translation', { keyPrefix: 'edit_answer' });
  const navigate = useNavigate();

  const initFormData = {
    content: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    description: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  };

  const { data } = useQueryAnswerInfo(aid);
  const [formData, setFormData] = useState<FormDataItem>(initFormData);
  const [immData, setImmData] = useState(initFormData);
  const [contentChanged, setContentChanged] = useState(false);
  const editCaptcha = useCaptchaPlugin('edit');

  useEffect(() => {
    if (data?.info?.content) {
      setFormData({
        ...formData,
        content: {
          value: data.info.content,
          isInvalid: false,
          errorMsg: '',
        },
      });
    }
  }, [data?.info?.content]);

  const { data: revisions = [] } = useQueryRevisions(aid);

  const editorRef = useRef<EditorRef>({
    getHtml: () => '',
  });

  const questionContentRef = useRef<HTMLDivElement>(null);
  useRenderHtmlPlugin(questionContentRef.current);

  useEffect(() => {
    if (!questionContentRef?.current) {
      return;
    }
    htmlRender(questionContentRef.current, {
      copySuccessText: t('copied', { keyPrefix: 'messages' }),
      copyText: t('copy', { keyPrefix: 'messages' }),
    });
  }, [questionContentRef]);

  usePromptWithUnload({
    when: contentChanged,
  });

  useEffect(() => {
    const { content, description } = formData;
    if (immData.content.value !== content.value || description.value) {
      setContentChanged(true);
    } else {
      setContentChanged(false);
    }
  }, [formData.content.value, formData.description.value]);

  const handleAnswerChange = (value: string) =>
    setFormData((prev) => ({
      ...prev,
      content: { ...prev.content, value },
    }));
  const handleSummaryChange = (evt) => {
    const v = evt.currentTarget.value;
    setFormData({
      ...formData,
      description: { ...formData.description, value: v },
    });
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const { content } = formData;

    if (!content.value || Array.from(content.value.trim()).length < 6) {
      bol = false;
      formData.content = {
        value: content.value,
        isInvalid: true,
        errorMsg: t('form.fields.answer.feedback.characters'),
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

  const submitEditAnswer = () => {
    const params: Type.AnswerParams = {
      content: formData.content.value,
      html: editorRef.current.getHtml(),
      question_id: qid,
      id: aid,
      edit_summary: formData.description.value,
    };
    editCaptcha?.resolveCaptchaReq(params);

    modifyAnswer(params)
      .then(async (res) => {
        await editCaptcha?.close();
        navigate(
          pathFactory.answerLanding({
            questionId: qid,
            slugTitle: data?.question?.url_title,
            answerId: aid,
          }),
          {
            state: { isReview: res?.wait_for_review },
          },
        );
      })
      .catch((ex) => {
        if (ex.isError) {
          editCaptcha?.handleCaptchaError(ex.list);
          const stateData = handleFormError(ex, formData);
          setFormData({ ...stateData });
        }
      });
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    setContentChanged(false);

    event.preventDefault();
    event.stopPropagation();

    if (!checkValidated()) {
      return;
    }

    if (!editCaptcha) {
      submitEditAnswer();
      return;
    }
    editCaptcha.check(() => submitEditAnswer());
  };
  const handleSelectedRevision = (e) => {
    const index = e.target.value;
    const revision = revisions[index];
    if (revision?.content) {
      formData.content.value = revision.content.content;
      setImmData({ ...formData });
      setFormData({ ...formData });
    }
  };

  const backPage = () => {
    navigate(-1);
  };
  usePageTags({
    title: t('edit_answer', { keyPrefix: 'page_title' }),
  });
  return (
    <div className="pt-4 mb-5 edit-answer-wrap">
      <h3 className="mb-4">{t('title')}</h3>
      <Row>
        <Col className="page-main flex-auto">
          <Link
            to={pathFactory.questionLanding(qid, data?.question.url_title)}
            target="_blank"
            rel="noreferrer">
            <h5 className="mb-3">{data?.question.title}</h5>
          </Link>

          <div className="question-content-wrap">
            <div
              ref={questionContentRef}
              className="content position-absolute top-0 w-100"
              dangerouslySetInnerHTML={{ __html: data?.question.html }}
            />
            <div
              className="resize-bottom"
              style={{ maxHeight: questionContentRef?.current?.scrollHeight }}
            />
            <div className="line bg-light  d-flex justify-content-center align-items-center">
              <Icon type="bi" name="grip-horizontal" className="mt-1" />
            </div>
          </div>

          <Form noValidate onSubmit={handleSubmit}>
            <Form.Group controlId="revision" className="mb-3">
              <Form.Label>{t('form.fields.revision.label')}</Form.Label>
              <Form.Select onChange={handleSelectedRevision} defaultValue={0}>
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

            <Form.Group controlId="answer" className="mt-3">
              <Form.Label>{t('form.fields.answer.label')}</Form.Label>
              <Editor
                value={formData.content.value}
                onChange={handleAnswerChange}
                className={classNames(
                  'form-control p-0',
                  focusType === 'answer' && 'focus',
                )}
                onFocus={() => {
                  setForceType('answer');
                }}
                onBlur={() => {
                  setForceType('');
                }}
                ref={editorRef}
              />
              <Form.Control
                value={formData.content.value}
                type="text"
                isInvalid={formData.content.isInvalid}
                readOnly
                hidden
              />
              <Form.Control.Feedback type="invalid">
                {formData.content.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
            <Form.Group controlId="edit_summary" className="my-3">
              <Form.Label>{t('form.fields.edit_summary.label')}</Form.Label>
              <Form.Control
                type="text"
                onChange={handleSummaryChange}
                defaultValue={formData.description.value}
                isInvalid={formData.description.isInvalid}
                placeholder={t('form.fields.edit_summary.placeholder')}
                contentEditable
              />
              <Form.Control.Feedback type="invalid">
                {formData.description.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>

            <div className="mt-3">
              <Button type="submit" className="me-2">
                {t('btn_save_edits')}
              </Button>
              <Button variant="link" onClick={backPage}>
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

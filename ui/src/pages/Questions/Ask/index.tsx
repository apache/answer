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

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Row, Col, Form, Button, Card } from 'react-bootstrap';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import dayjs from 'dayjs';
import classNames from 'classnames';
import isEqual from 'lodash/isEqual';
import debounce from 'lodash/debounce';
import fm from 'front-matter';

import { writeSettingStore } from '@/stores';
import { usePageTags, usePromptWithUnload } from '@/hooks';
import { Editor, EditorRef, TagSelector } from '@/components';
import type * as Type from '@/common/interface';
import { DRAFT_QUESTION_STORAGE_KEY } from '@/common/constants';
import {
  saveQuestion,
  questionDetail,
  modifyQuestion,
  useQueryRevisions,
  queryQuestionByTitle,
  getTagsBySlugName,
  saveQuestionWithAnswer,
} from '@/services';
import {
  handleFormError,
  SaveDraft,
  storageExpires,
  scrollToElementTop,
} from '@/utils';
import { pathFactory } from '@/router/pathFactory';
import { useCaptchaPlugin } from '@/utils/pluginKit';

import SearchQuestion from './components/SearchQuestion';

interface FormDataItem {
  title: Type.FormValue<string>;
  tags: Type.FormValue<Type.Tag[]>;
  content: Type.FormValue<string>;
  answer_content: Type.FormValue<string>;
  edit_summary: Type.FormValue<string>;
}

const saveDraft = new SaveDraft({ type: 'question' });

const Ask = () => {
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
    answer_content: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    edit_summary: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  };
  const { t } = useTranslation('translation', { keyPrefix: 'ask' });
  const [formData, setFormData] = useState<FormDataItem>(initFormData);
  const [immData, setImmData] = useState<FormDataItem>(initFormData);
  const [checked, setCheckState] = useState(false);
  const [blockState, setBlockState] = useState(false);
  const [focusType, setForceType] = useState('');
  const [hasDraft, setHasDraft] = useState(false);
  const resetForm = () => {
    setFormData(initFormData);
    setCheckState(false);
    setForceType('');
  };
  const [similarQuestions, setSimilarQuestions] = useState([]);

  const editorRef = useRef<EditorRef>({
    getHtml: () => '',
  });
  const editorRef2 = useRef<EditorRef>({
    getHtml: () => '',
  });

  const { qid } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const updateTags = (tags: string) => {
    getTagsBySlugName(tags).then((resp) => {
      // eslint-disable-next-line
      handleTagsChange(resp);
    });
  };
  const writeInfo = writeSettingStore((state) => state.write);

  const isEdit = qid !== undefined;

  const saveCaptcha = useCaptchaPlugin('question');
  const editCaptcha = useCaptchaPlugin('edit');

  const removeDraft = () => {
    saveDraft.save.cancel();
    saveDraft.remove();
    setHasDraft(false);
  };

  useEffect(() => {
    if (!qid) {
      // order: 1. tags query. 2. prefill query. 3. draft
      const queryTags = searchParams.get('tags');
      if (queryTags) {
        updateTags(queryTags);
      }
      const draft = storageExpires.get(DRAFT_QUESTION_STORAGE_KEY);

      const prefill = searchParams.get('prefill');
      if (prefill || draft) {
        if (prefill) {
          const file = fm<any>(decodeURIComponent(prefill));
          formData.title.value = file.attributes?.title;
          formData.content.value = file.body;
          if (!queryTags && file.attributes?.tags) {
            // Remove spaces in file.attributes.tags
            const filterTags = file.attributes.tags
              .split(',')
              .map((tag) => tag.trim())
              .join(',');
            updateTags(filterTags);
          }
        } else if (draft) {
          formData.title.value = draft.title;
          formData.content.value = draft.content;
          formData.tags.value = draft.tags;
          formData.answer_content.value = draft.answer_content;
          setCheckState(Boolean(draft.answer_content));
          setHasDraft(true);
        }
        setFormData({ ...formData });
      } else {
        resetForm();
      }
    }

    return () => {
      resetForm();
    };
  }, [qid]);

  useEffect(() => {
    const { title, tags, content, answer_content } = formData;
    const { title: editTitle, tags: editTags, content: editContent } = immData;

    // edited
    if (qid) {
      if (
        editTitle.value !== title.value ||
        editContent.value !== content.value ||
        !isEqual(
          editTags.value.map((v) => v.slug_name),
          tags.value.map((v) => v.slug_name),
        )
      ) {
        setBlockState(true);
      } else {
        setBlockState(false);
      }
      return;
    }
    // write
    if (
      title.value ||
      tags.value.length > 0 ||
      content.value ||
      answer_content.value
    ) {
      // save draft
      saveDraft.save({
        params: {
          title: title.value,
          tags: tags.value,
          content: content.value,
          answer_content: answer_content.value,
        },
        callback: () => setHasDraft(true),
      });
      setBlockState(true);
    } else {
      removeDraft();
      setBlockState(false);
    }
  }, [formData]);

  usePromptWithUnload({
    when: blockState,
  });

  const { data: revisions = [] } = useQueryRevisions(qid);

  useEffect(() => {
    if (!isEdit) {
      return;
    }
    questionDetail(qid).then((res) => {
      formData.title.value = res.title;
      formData.content.value = res.content;
      formData.tags.value = res.tags.map((item) => {
        return {
          ...item,
          parsed_text: '',
          original_text: '',
        };
      });
      setImmData({ ...formData });
      setFormData({ ...formData });
    });
  }, [qid]);

  const querySimilarQuestions = useCallback(
    debounce((title) => {
      queryQuestionByTitle(title).then((res) => {
        setSimilarQuestions(res);
      });
    }, 400),
    [],
  );

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      title: { value: e.currentTarget.value, errorMsg: '', isInvalid: false },
    });
    if (e.currentTarget.value.length >= 10) {
      querySimilarQuestions(e.currentTarget.value);
    }
    if (e.currentTarget.value.length === 0) {
      setSimilarQuestions([]);
    }
  };
  const handleContentChange = (value: string) => {
    setFormData({
      ...formData,
      content: { value, errorMsg: '', isInvalid: false },
    });
  };
  const handleTagsChange = (value) =>
    setFormData({
      ...formData,
      tags: { value, errorMsg: '', isInvalid: false },
    });

  const handleAnswerChange = (value: string) =>
    setFormData({
      ...formData,
      answer_content: { value, errorMsg: '', isInvalid: false },
    });

  const handleSummaryChange = (evt: React.ChangeEvent<HTMLInputElement>) =>
    setFormData({
      ...formData,
      edit_summary: {
        ...formData.edit_summary,
        value: evt.currentTarget.value,
      },
    });

  const deleteDraft = () => {
    const res = window.confirm(t('discard_confirm', { keyPrefix: 'draft' }));
    if (res) {
      removeDraft();
      resetForm();
    }
  };

  const submitModifyQuestion = (params) => {
    setBlockState(false);
    const ep = {
      ...params,
      id: qid,
      edit_summary: formData.edit_summary.value,
    };
    const imgCode = editCaptcha?.getCaptcha();
    if (imgCode?.verify) {
      ep.captcha_code = imgCode.captcha_code;
      ep.captcha_id = imgCode.captcha_id;
    }
    modifyQuestion(ep)
      .then(async (res) => {
        await editCaptcha?.close();
        navigate(pathFactory.questionLanding(qid, res?.url_title), {
          state: { isReview: res?.wait_for_review },
        });
      })
      .catch((err) => {
        if (err.isError) {
          editCaptcha?.handleCaptchaError(err.list);
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        }
      });
  };

  const submitQuestion = async (params) => {
    setBlockState(false);
    const imgCode = saveCaptcha?.getCaptcha();
    if (imgCode?.verify) {
      params.captcha_code = imgCode.captcha_code;
      params.captcha_id = imgCode.captcha_id;
    }
    let res;
    if (checked) {
      res = await saveQuestionWithAnswer({
        ...params,
        answer_content: formData.answer_content.value,
      }).catch((err) => {
        if (err.isError) {
          const captchaErr = saveCaptcha?.handleCaptchaError(err.list);
          if (!(captchaErr && err.list.length === 1)) {
            const data = handleFormError(err, formData);
            setFormData({ ...data });
            const ele = document.getElementById(err.list[0].error_field);
            scrollToElementTop(ele);
          }
        }
      });
    } else {
      res = await saveQuestion(params).catch((err) => {
        if (err.isError) {
          const captchaErr = saveCaptcha?.handleCaptchaError(err.list);
          if (!(captchaErr && err.list.length === 1)) {
            const data = handleFormError(err, formData);
            setFormData({ ...data });
            const ele = document.getElementById(err.list[0].error_field);
            scrollToElementTop(ele);
          }
        }
      });
    }

    const id = res?.id || res?.question?.id;
    if (id) {
      await saveCaptcha?.close();
      if (checked) {
        navigate(pathFactory.questionLanding(id, res?.question?.url_title));
      } else {
        navigate(pathFactory.questionLanding(id, res?.url_title));
      }
    }
    removeDraft();
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    event.stopPropagation();

    const params: Type.QuestionParams = {
      title: formData.title.value,
      content: formData.content.value,
      tags: formData.tags.value,
    };

    if (isEdit) {
      if (!editCaptcha) {
        submitModifyQuestion(params);
        return;
      }
      editCaptcha.check(() => submitModifyQuestion(params));
    } else {
      if (!saveCaptcha) {
        submitQuestion(params);
        return;
      }
      saveCaptcha?.check(async () => {
        submitQuestion(params);
      });
    }
  };
  const backPage = () => {
    navigate(-1);
  };

  const handleSelectedRevision = (e) => {
    const index = e.target.value;
    const revision = revisions[index];
    formData.content.value = revision.content?.content || '';
    setImmData({ ...formData });
    setFormData({ ...formData });
  };
  const bool = similarQuestions.length > 0 && !isEdit;
  let pageTitle = t('ask_a_question', { keyPrefix: 'page_title' });
  if (isEdit) {
    pageTitle = t('edit_question', { keyPrefix: 'page_title' });
  }
  usePageTags({
    title: pageTitle,
  });

  const handleContentHint = () => {
    if (
      !writeInfo ||
      writeInfo.min_content === undefined ||
      !writeInfo.min_content
    ) {
      return t(`form.fields.body.hint.optional_body`);
    }

    let str: string = t(`form.fields.body.hint.minimum_characters`);
    str = str.replace(
      `{{ min_content_length }}`,
      writeInfo.min_content.toString(),
    );
    return str;
  };

  return (
    <div className="pt-4 mb-5">
      <h3 className="mb-4">{isEdit ? t('edit_title') : t('title')}</h3>
      <Row>
        <Col className="page-main flex-auto">
          <Form noValidate onSubmit={handleSubmit}>
            {isEdit && (
              <Form.Group controlId="revision" className="mb-3">
                <Form.Label>{t('form.fields.revision.label')}</Form.Label>
                <Form.Select onChange={handleSelectedRevision}>
                  {revisions.map(({ reason, create_at, user_info }, index) => {
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
            )}
            <Form.Group controlId="title" className="mb-3">
              <Form.Label>{t('form.fields.title.label')}</Form.Label>
              <Form.Control
                type="text"
                value={formData.title.value}
                isInvalid={formData.title.isInvalid}
                onChange={handleTitleChange}
                placeholder={t('form.fields.title.placeholder')}
                autoFocus
                contentEditable
              />
              <Form.Control.Feedback type="invalid">
                {formData.title.errorMsg}
              </Form.Control.Feedback>
              {bool && <SearchQuestion similarQuestions={similarQuestions} />}
            </Form.Group>
            <Form.Group controlId="content">
              <Form.Label>{t('form.fields.body.label')}</Form.Label>
              <Editor
                value={formData.content.value}
                onChange={handleContentChange}
                className={classNames(
                  'form-control p-0',
                  focusType === 'content' && 'focus',
                  formData.content.isInvalid && 'is-invalid',
                )}
                onFocus={() => {
                  setForceType('content');
                }}
                onBlur={() => {
                  setForceType('');
                }}
                ref={editorRef}
              />
              <Form.Text>{handleContentHint()}</Form.Text>
              <Form.Control.Feedback type="invalid">
                {formData.content.errorMsg}
              </Form.Control.Feedback>
            </Form.Group>
            <Form.Group controlId="tags" className="my-3">
              <Form.Label>{t('form.fields.tags.label')}</Form.Label>
              <TagSelector
                value={formData.tags.value}
                onChange={handleTagsChange}
                showRequiredTag
                maxTagLength={5}
                isInvalid={formData.tags.isInvalid}
                errMsg={formData.tags.errorMsg}
              />
            </Form.Group>
            {!isEdit && (
              <>
                <Form.Switch
                  checked={checked}
                  type="switch"
                  label={t('answer_question')}
                  onChange={(e) => setCheckState(e.target.checked)}
                  id="radio-answer"
                />
                {checked && (
                  <Form.Group controlId="answer" className="mt-3">
                    <Form.Label>{t('form.fields.answer.label')}</Form.Label>
                    <Editor
                      value={formData.answer_content.value}
                      onChange={handleAnswerChange}
                      ref={editorRef2}
                      className={classNames(
                        'form-control p-0',
                        focusType === 'answer' && 'focus',
                        formData.answer_content.isInvalid && 'is-invalid',
                      )}
                      onFocus={() => {
                        setForceType('answer');
                      }}
                      onBlur={() => {
                        setForceType('');
                      }}
                    />
                    <Form.Control
                      type="text"
                      isInvalid={formData.answer_content.isInvalid}
                      hidden
                    />
                    <Form.Control.Feedback type="invalid">
                      {formData.answer_content.errorMsg}
                    </Form.Control.Feedback>
                  </Form.Group>
                )}
              </>
            )}
            {isEdit && (
              <Form.Group controlId="edit_summary" className="my-3">
                <Form.Label>{t('form.fields.edit_summary.label')}</Form.Label>
                <Form.Control
                  type="text"
                  defaultValue={formData.edit_summary.value}
                  isInvalid={formData.edit_summary.isInvalid}
                  placeholder={t('form.fields.edit_summary.placeholder')}
                  onChange={handleSummaryChange}
                  contentEditable
                />
                <Form.Control.Feedback type="invalid">
                  {formData.edit_summary.errorMsg}
                </Form.Control.Feedback>
              </Form.Group>
            )}
            {!checked && (
              <div className="mt-3">
                <Button type="submit" className="me-2">
                  {isEdit ? t('btn_save_edits') : t('btn_post_question')}
                </Button>
                {isEdit && (
                  <Button variant="link" onClick={backPage}>
                    {t('cancel', { keyPrefix: 'btns' })}
                  </Button>
                )}

                {hasDraft && (
                  <Button variant="link" onClick={deleteDraft}>
                    {t('discard_draft', { keyPrefix: 'btns' })}
                  </Button>
                )}
              </div>
            )}
            {checked && (
              <div className="mt-3">
                <Button type="submit">{t('post_question&answer')}</Button>
                {hasDraft && (
                  <Button variant="link" className="ms-2" onClick={deleteDraft}>
                    {t('discard_draft', { keyPrefix: 'btns' })}
                  </Button>
                )}
              </div>
            )}
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

export default Ask;

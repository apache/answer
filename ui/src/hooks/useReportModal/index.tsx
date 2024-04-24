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

import { useState, useRef, useEffect, useLayoutEffect } from 'react';
import { Modal, Form, Button, FormCheck } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import ReactDOM from 'react-dom/client';

import { useToast, useCaptchaModal } from '@/hooks';
import type * as Type from '@/common/interface';
import {
  reportList,
  postReport,
  closeQuestion,
  putReport,
  putFlagReviewAction,
} from '@/services';

interface Params {
  isBackend?: boolean;
  type: Type.ReportType;
  id: string;
  title?: string;
  action: Type.ReportAction;
  source?: string;
  content?: string;
  reportType?: any;
}

const useReportModal = (callback?: () => void) => {
  const { t } = useTranslation('translation', { keyPrefix: 'report_modal' });
  const toast = useToast();
  const [params, setParams] = useState<Params | null>(null);
  const [isInvalid, setInvalidState] = useState(false);
  const [reportType, setReportType] = useState({
    type: -1,
    haveContent: false,
  });
  const rootRef = useRef<{ root: ReactDOM.Root | null }>({
    root: null,
  });

  const [content, setContent] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });
  const [show, setShow] = useState(false);
  const [list, setList] = useState<any[]>([]);

  const rCaptcha = useCaptchaModal('report');

  useEffect(() => {
    const div = document.createElement('div');
    rootRef.current.root = ReactDOM.createRoot(div);
  }, []);
  const getList = ({ type, action, isBackend, ...otherParams }: Params) => {
    // @ts-ignore
    reportList({
      type,
      action,
      isBackend,
    }).then((res) => {
      setList(res);
      if (otherParams.reportType) {
        const findType = res.find(
          (v) => v.reason_type === otherParams.reportType,
        );
        if (findType) {
          setReportType({
            type: findType.reason_type,
            haveContent: Boolean(findType.content_type),
          });

          setContent({
            value: otherParams.content || '',
            isInvalid: false,
            errorMsg: '',
          });
        }
      }
      setShow(true);
    });
  };
  const asyncCallback = () => {
    setTimeout(() => {
      callback?.();
    });
  };
  const handleRadio = (val) => {
    setInvalidState(false);
    setContent({
      value:
        val.reason_type === params?.reportType ? String(params?.content) : '',
      isInvalid: false,
      errorMsg: '',
    });
    setReportType({
      type: val.reason_type,
      haveContent: Boolean(val.content_type),
    });
  };

  const onClose = () => {
    setReportType({
      type: -1,
      haveContent: false,
    });
    setContent({
      value: '',
      isInvalid: false,
      errorMsg: '',
    });
    setShow(false);
  };

  const checkValidate = () => {
    if (reportType.haveContent && !content.value) {
      setContent({
        value: content.value,
        isInvalid: true,
        errorMsg: t('remark.empty'),
      });
      return false;
    }

    if (reportType.type === 60) {
      // a duplicate
      let url: URL | undefined;
      try {
        url = new URL(content.value);
      } catch {
        setContent({
          value: content.value,
          isInvalid: true,
          errorMsg: t('msg.not_a_url'),
        });
      }
      if (!url) return false;

      if (url.origin !== window.location.origin) {
        setContent({
          value: content.value,
          isInvalid: true,
          errorMsg: t('msg.url_not_match'),
        });
        return false;
      }
    }

    return true;
  };

  const handleSubmit = () => {
    if (!params) {
      return;
    }
    if (reportType.type === -1) {
      setInvalidState(true);
      return;
    }

    if (!checkValidate()) {
      return;
    }

    if (params.type === 'question' && params.action === 'close') {
      if (params?.source === 'review') {
        putFlagReviewAction({
          flag_id: params.id,
          operation_type: 'close_post',
          close_type: reportType.type,
          close_msg: content.value,
        }).then(() => {
          onClose();
          asyncCallback();
        });
        return;
      }
      closeQuestion({
        id: params.id,
        close_type: reportType.type,
        close_msg: content.value,
      }).then(() => {
        onClose();
        asyncCallback();
      });
      return;
    }
    if (!params.isBackend && params.action === 'flag') {
      rCaptcha.check(() => {
        const flagReq = {
          source: params.type,
          report_type: reportType.type,
          object_id: params.id,
          content: content.value,
          captcha_code: undefined,
          captcha_id: undefined,
        };
        rCaptcha.resolveCaptchaReq(flagReq);

        postReport(flagReq)
          .then(async () => {
            await rCaptcha.close();
            toast.onShow({
              msg: t('flag_success', { keyPrefix: 'toast' }),
              variant: 'warning',
            });
            onClose();
            asyncCallback();
          })
          .catch((ex) => {
            if (ex.isError) {
              rCaptcha.handleCaptchaError(ex.list);
            }
          });
      });
    }

    if (params.isBackend && params.action === 'review') {
      putReport({
        action: params.type,
        flagged_content: content.value,
        flagged_type: reportType.type,
        id: params.id,
      }).then(() => {
        onClose();
        asyncCallback();
      });
    }
  };

  const onShow = (obj: Params) => {
    setParams(obj);
    getList(obj);
  };
  let title = '';
  if (typeof params === 'object' && params) {
    title = params.title || t(`${params.action}_title`);
    if (params.action === 'review') {
      title = t(`${params.action}_${params.type}_title`);
    }
  }
  useLayoutEffect(() => {
    rootRef.current.root?.render(
      <Modal show={show} onHide={onClose}>
        <Modal.Header closeButton>
          <Modal.Title as="h5">{title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            {list.map((item) => {
              return (
                <div key={item?.reason_type}>
                  <Form.Group
                    controlId={`report_${item?.reason_type}`}
                    className={`${
                      item.have_content && reportType === item.type
                        ? 'mb-2'
                        : 'mb-3'
                    }`}>
                    <FormCheck>
                      <FormCheck.Input
                        id={item.reason_type}
                        type="radio"
                        checked={reportType.type === item.reason_type}
                        onChange={() => handleRadio(item)}
                        isInvalid={isInvalid}
                      />
                      <FormCheck.Label htmlFor={item.reason_type}>
                        <span className="fw-bold">{item?.name}</span>
                        <br />
                        <span className="text-secondary">
                          {item?.description}
                        </span>
                      </FormCheck.Label>
                      <Form.Control.Feedback type="invalid">
                        {t('msg.empty')}
                      </Form.Control.Feedback>
                    </FormCheck>
                  </Form.Group>
                  {reportType.haveContent &&
                    reportType.type === item.reason_type && (
                      <Form.Group controlId="content" className="ps-4 mb-3">
                        <Form.Control
                          type="text"
                          as={
                            item.content_type === 'text' ? 'input' : 'textarea'
                          }
                          value={content.value}
                          isInvalid={content.isInvalid}
                          placeholder={item.placeholder}
                          onChange={(e) =>
                            setContent({
                              value: e.target.value,
                              isInvalid: false,
                              errorMsg: '',
                            })
                          }
                        />
                        <Form.Control.Feedback type="invalid">
                          {content.errorMsg}
                        </Form.Control.Feedback>
                      </Form.Group>
                    )}
                </div>
              );
            })}
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => onClose()}>
            {t('btn_cancel')}
          </Button>
          <Button variant="primary" onClick={handleSubmit}>
            {t('btn_submit')}
          </Button>
        </Modal.Footer>
      </Modal>,
    );
  });
  return {
    onClose,
    onShow,
  };
};

export default useReportModal;

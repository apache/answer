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

import { FormEvent, useEffect, useRef, useState } from 'react';
import { Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { TabNav } from '@/components';
import { ADMIN_AI_ASSISTANT_NAV_MENUS } from '@/common/constants';
import type * as Type from '@/common/interface';
import { useToast } from '@/hooks';
import { getAiPromptConfig, saveAiPromptConfig } from '@/services';

const getPromptByLang = (
  promptConfig: Type.AIPromptConfig | undefined,
  lang: string,
) => {
  if (!promptConfig) {
    return '';
  }
  const isZh = lang?.toLowerCase().startsWith('zh');
  return isZh ? promptConfig.zh_cn || '' : promptConfig.en_us || '';
};

const Settings = () => {
  const { t, i18n } = useTranslation('translation', {
    keyPrefix: 'admin.conversations',
  });
  const toast = useToast();
  const historyPromptConfigRef = useRef<Type.AIPromptConfig>();
  const [isSaving, setIsSaving] = useState(false);
  const [promptForm, setPromptForm] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });
  const isZhLang = i18n.language?.toLowerCase().startsWith('zh');

  const getAiPromptConfigData = async () => {
    const promptConfig = await getAiPromptConfig();
    historyPromptConfigRef.current = promptConfig;
    setPromptForm({
      value: getPromptByLang(promptConfig, i18n.language),
      isInvalid: false,
      errorMsg: '',
    });
  };

  const handleSavePrompt = (evt: FormEvent) => {
    evt.preventDefault();
    if (isSaving) {
      return;
    }
    setIsSaving(true);
    setPromptForm((prev) => ({
      ...prev,
      isInvalid: false,
      errorMsg: '',
    }));

    const params: Type.AIPromptConfig = {
      zh_cn: isZhLang
        ? promptForm.value
        : historyPromptConfigRef.current?.zh_cn || '',
      en_us: isZhLang
        ? historyPromptConfigRef.current?.en_us || ''
        : promptForm.value,
    };

    saveAiPromptConfig(params)
      .then(() => {
        historyPromptConfigRef.current = params;
        toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
      })
      .catch((err) => {
        setPromptForm((prev) => ({
          ...prev,
          isInvalid: true,
          errorMsg: err?.message || '',
        }));
      })
      .finally(() => {
        setIsSaving(false);
      });
  };

  useEffect(() => {
    getAiPromptConfigData();
  }, []);

  useEffect(() => {
    if (!historyPromptConfigRef.current) {
      return;
    }
    setPromptForm((prev) => ({
      ...prev,
      value: getPromptByLang(historyPromptConfigRef.current, i18n.language),
      isInvalid: false,
      errorMsg: '',
    }));
  }, [i18n.language]);

  return (
    <div className="d-flex flex-column flex-grow-1 position-relative">
      <h3 className="mb-4">{t('ai_assistant', { keyPrefix: 'nav_menus' })}</h3>
      <TabNav
        menus={ADMIN_AI_ASSISTANT_NAV_MENUS}
        i18nKeyPrefix="admin.conversations.tabs"
      />

      <div className="max-w-748">
        <Form noValidate onSubmit={handleSavePrompt}>
          <div className="mb-3">
            <label className="form-label" htmlFor="admin-prompt-textarea">
              {t('prompt.label', { keyPrefix: 'admin.ai_settings' })}
            </label>
            <Form.Control
              id="admin-prompt-textarea"
              as="textarea"
              rows={10}
              isInvalid={promptForm.isInvalid}
              value={promptForm.value}
              onChange={(e) =>
                setPromptForm({
                  value: e.target.value,
                  isInvalid: false,
                  errorMsg: '',
                })
              }
            />
            <div className="form-text mt-1">
              {t('prompt.text', { keyPrefix: 'admin.ai_settings' })}
            </div>
            <Form.Control.Feedback type="invalid">
              {promptForm.errorMsg}
            </Form.Control.Feedback>
          </div>
          <Button type="submit" className="btn-primary" disabled={isSaving}>
            {t('save', { keyPrefix: 'btns' })}
          </Button>
        </Form>
      </div>
    </div>
  );
};

export default Settings;

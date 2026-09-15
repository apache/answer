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

import { useState } from 'react';
import { Button, Form, Spinner } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { aiControlStore, interfaceStore, toastStore } from '@/stores';
import {
  translateContent,
  TranslateContentResponse,
} from '@/services/client/ai';
import Modal from '../Modal';

interface Props {
  title?: string;
  content: string;
  className?: string;
  onApply: (translation: { title?: string; content: string }) => void;
}

const AITranslateButton = ({ title, content, className, onApply }: Props) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_translate' });
  const { ai_enabled: aiEnabled, ai_translation_enabled: translationEnabled } =
    aiControlStore((state) => state);
  const targetLanguage = interfaceStore((state) => state.interface.language);
  const [loading, setLoading] = useState(false);
  const [translation, setTranslation] =
    useState<TranslateContentResponse | null>(null);

  if (!aiEnabled || !translationEnabled) {
    return null;
  }

  const requestTranslation = async () => {
    setLoading(true);
    try {
      const result = await translateContent({ title, content });
      setTranslation(result);
    } catch (error: any) {
      toastStore.getState().show({
        msg: error?.msg || t('error'),
        variant: 'danger',
      });
    } finally {
      setLoading(false);
    }
  };

  const updateTranslation = (changes: Partial<TranslateContentResponse>) => {
    setTranslation((current) => (current ? { ...current, ...changes } : null));
  };

  return (
    <>
      <Button
        type="button"
        variant="link"
        size="sm"
        className={`p-0 text-decoration-none ${className || ''}`}
        disabled={loading || (!title?.trim() && !content.trim())}
        onClick={requestTranslation}>
        {loading && <Spinner size="sm" className="me-2" />}
        {loading ? t('translating') : t('button')}
      </Button>
      <Modal
        title={t('review_title')}
        visible={Boolean(translation)}
        scrollable
        cancelText={t('discard')}
        cancelBtnVariant="outline-secondary"
        confirmText={t('apply')}
        confirmBtnVariant="primary"
        confirmBtnDisabled={
          !translation?.content.trim() && !translation?.title.trim()
        }
        onCancel={() => setTranslation(null)}
        onConfirm={() => {
          if (!translation) {
            return;
          }
          onApply({
            title: title === undefined ? undefined : translation.title,
            content: translation.content,
          });
          setTranslation(null);
        }}>
        <p className="text-secondary small">
          {t('review_description', {
            language: translation?.target_language || targetLanguage,
          })}
        </p>
        {title !== undefined && (
          <Form.Group className="mb-3">
            <Form.Label>{t('title_label')}</Form.Label>
            <Form.Control
              value={translation?.title || ''}
              maxLength={150}
              onChange={(event) =>
                updateTranslation({ title: event.currentTarget.value })
              }
            />
          </Form.Group>
        )}
        <Form.Group>
          <Form.Label>{t('content_label')}</Form.Label>
          <Form.Control
            as="textarea"
            rows={14}
            value={translation?.content || ''}
            onChange={(event) =>
              updateTranslation({ content: event.currentTarget.value })
            }
          />
        </Form.Group>
      </Modal>
    </>
  );
};

export default AITranslateButton;

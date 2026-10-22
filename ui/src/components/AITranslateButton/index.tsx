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

import { useEffect, useMemo, useState } from 'react';
import { Button, Spinner } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import { aiControlStore, interfaceStore, toastStore } from '@/stores';
import { translateContent } from '@/services/client/ai';
import { doesTextNeedTranslation } from '@/utils/languageDetection';
import Icon from '../Icon';

import './index.scss';

interface Props {
  title?: string;
  content?: string;
  className?: string;
  onApply: (value: string) => void;
}

const getLanguageName = (locale: string, displayLocale?: string) => {
  const language = locale.split(/[-_]/)[0];
  try {
    return (
      new Intl.DisplayNames([displayLocale?.replace('_', '-') || 'en'], {
        type: 'language',
      }).of(language) || locale
    );
  } catch {
    return locale;
  }
};

const AITranslateButton = ({ title, content, className, onApply }: Props) => {
  const { t, i18n } = useTranslation('translation', {
    keyPrefix: 'ai_translate',
  });
  const { ai_enabled: aiEnabled, ai_translation_enabled: translationEnabled } =
    aiControlStore((state) => state);
  const targetLanguage = interfaceStore((state) => state.interface.language);
  const [loading, setLoading] = useState(false);
  const [languageMismatch, setLanguageMismatch] = useState(false);
  const value = title ?? content ?? '';
  const languageName = useMemo(
    () => getLanguageName(targetLanguage, i18n.resolvedLanguage),
    [i18n.resolvedLanguage, targetLanguage],
  );

  useEffect(() => {
    if (!aiEnabled || !translationEnabled) {
      setLanguageMismatch(false);
      return undefined;
    }

    let active = true;
    const timeout = window.setTimeout(async () => {
      const mismatch = await doesTextNeedTranslation(value, targetLanguage);
      if (active) {
        setLanguageMismatch(mismatch);
      }
    }, 300);

    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [aiEnabled, targetLanguage, translationEnabled, value]);

  if (!aiEnabled || !translationEnabled || !languageMismatch) {
    return null;
  }

  const requestTranslation = async () => {
    setLoading(true);
    try {
      const result = await translateContent(
        title !== undefined ? { title } : { content },
      );
      onApply(title !== undefined ? result.title : result.content);
    } catch (error: any) {
      toastStore.getState().show({
        msg: error?.msg || t('error'),
        variant: 'danger',
      });
    } finally {
      setLoading(false);
    }
  };

  const label = loading
    ? t('translating')
    : t('button', { language: languageName });

  return (
    <Button
      type="button"
      variant="light"
      className={classNames('ai-translate-button', className)}
      disabled={loading}
      title={label}
      aria-label={label}
      onClick={requestTranslation}>
      {loading ? (
        <Spinner animation="border" size="sm" />
      ) : (
        <Icon type="bi" name="translate" />
      )}
    </Button>
  );
};

export default AITranslateButton;

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

import { useEffect, useState, useRef, FC } from 'react';
import { Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import classnames from 'classnames';

import { Icon } from '@/components';

import './index.scss';

interface IProps {
  onSubmit?: (value: string) => void;
  onCancel?: () => void;
  isGenerate: boolean;
  hasConversation: boolean;
}

const Sender: FC<IProps> = ({
  onSubmit,
  onCancel,
  isGenerate,
  hasConversation,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const containerRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [initialized, setInitialized] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [isFocus, setIsFocus] = useState(false);

  const handleFocus = () => {
    setIsFocus(true);
    textareaRef?.current?.focus();
  };

  const handleBlur = () => {
    setIsFocus(false);
  };

  const autoResize = () => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    textarea.style.height = '32px';

    const minHeight = 32; // minimum height
    const maxHeight = 96; // maximum height

    // calculate the height needed
    const { scrollHeight } = textarea;
    const newHeight = Math.min(Math.max(scrollHeight, minHeight), maxHeight);

    // set the new height
    textarea.style.height = `${newHeight}px`;

    // control the scrollbar display
    if (scrollHeight > maxHeight) {
      textarea.style.overflowY = 'auto';
    } else {
      textarea.style.overflowY = 'hidden';
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputValue(e.target.value);
    setTimeout(autoResize, 0);
  };

  const handleSubmit = () => {
    if (isGenerate || !inputValue.trim()) {
      return;
    }
    onSubmit?.(inputValue);
    setInputValue('');
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault(); // Prevent default behavior of Enter key
      handleSubmit();
    } else if (e.key === 'Escape') {
      setInputValue((prev) => `${prev}\n`); // Add a new line on Escape key
    }
  };

  useEffect(() => {
    setInitialized(true);
  }, []);

  useEffect(() => {
    const handleOutsideClick = (event) => {
      if (
        initialized &&
        containerRef.current &&
        !containerRef.current?.contains(event.target)
      ) {
        handleBlur();
      }
    };
    document.addEventListener('click', handleOutsideClick);
    return () => {
      document.removeEventListener('click', handleOutsideClick);
    };
  }, [initialized]);
  return (
    <div
      className={classnames(
        'sender-wrap',
        hasConversation ? 'sticky-bottom pb-4' : 'mt-0',
      )}
      ref={containerRef}>
      <div
        onClick={handleFocus}
        className={classnames(
          'position-relative form-control p-3',
          isFocus ? 'form-control-focus' : '',
        )}>
        <Form.Control
          as="textarea"
          ref={textareaRef}
          style={{ height: '32px' }}
          className="input border-0 p-0"
          placeholder={t('ask_placeholder')}
          value={inputValue}
          onFocus={handleFocus}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
        />
        <div className="clearfix tools">
          {isGenerate ? (
            <Button
              variant="link"
              onClick={onCancel}
              className="p-0 lh-1 link-dark float-end">
              <Icon name="stop-circle-fill" size="24px" />
            </Button>
          ) : (
            <Button
              variant="link"
              className="p-0 lh-1 link-dark float-end"
              onClick={handleSubmit}>
              <Icon name="arrow-up-circle-fill" size="24px" />
            </Button>
          )}
        </div>
      </div>

      <Form.Text className="text-center d-block">{t('ai_generate')}</Form.Text>
    </div>
  );
};

export default Sender;

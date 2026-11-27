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
import aiControlStore from '@/stores/aiControl';
import { useToast } from '@/hooks';

import './index.scss';

interface IProps {
  onSubmit?: (value: string, images: string[]) => void;
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
  const toast = useToast();
  const containerRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [initialized, setInitialized] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [isFocus, setIsFocus] = useState(false);
  const visionEnabled = aiControlStore((s) => s.ai_vision_enabled);

  const addImages = (files: FileList | null) => {
    if (!files?.length) return;
    const room = 4 - images.length;
    if (room <= 0) {
      toast.onShow({
        msg: t('image_max_count'),
        variant: 'danger',
      });
      return;
    }
    Array.from(files)
      .slice(0, room)
      .forEach((file) => {
        if (!/^image\/(png|jpe?g|webp)$/.test(file.type)) return;
        if (file.size > 4 * 1024 * 1024) {
          toast.onShow({
            msg: t('image_too_large'),
            variant: 'danger',
          });
          return;
        }
        const reader = new FileReader();
        reader.onload = () => {
          const dataUrl = String(reader.result);
          setImages((prev) =>
            prev.length < 4 && !prev.includes(dataUrl) ? [...prev, dataUrl] : prev,
          );
        };
        reader.readAsDataURL(file);
      });
  };

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
    onSubmit?.(inputValue, images);
    setInputValue('');
    setImages([]);
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
        {images.length > 0 && (
          <div className="d-flex flex-wrap gap-2 mb-2">
            {images.map((img, i) => (
              <div key={`${i}-${img.slice(-12)}`} className="position-relative">
                <img
                  src={img}
                  alt=""
                  style={{ width: '48px', height: '48px', objectFit: 'cover' }}
                  className="rounded border"
                />
                <button
                  type="button"
                  aria-label="remove"
                  onClick={() =>
                    setImages((prev) => prev.filter((_, j) => j !== i))
                  }
                  className="btn btn-sm btn-dark position-absolute top-0 end-0 py-0 px-1 rounded-circle"
                  style={{ transform: 'translate(40%, -40%)' }}>
                  ×
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="clearfix tools">
          {visionEnabled && !isGenerate && (
            <>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp"
                multiple
                hidden
                onChange={(e) => {
                  addImages(e.target.files);
                  e.target.value = '';
                }}
              />
              <Button
                variant="link"
                className="p-0 lh-1 link-secondary float-start me-2"
                title={t('image_upload')}
                onClick={() => fileInputRef.current?.click()}>
                <Icon name="image" size="20px" />
              </Button>
            </>
          )}
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

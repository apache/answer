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

import { FC, useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Form, Dropdown } from 'react-bootstrap';

import { TagInfo } from '@/common/interface';
import request from '@/utils/request';

import Modal from './Modal';

interface Props {
  visible: boolean;
  sourceTag: TagInfo;
  onClose: () => void;
  onConfirm: (sourceTagID: string, targetTagID: string) => void;
}

interface SearchTagResp {
  tag_id: string;
  slug_name: string;
  display_name: string;
  recommend: boolean;
  reserved: boolean;
}

const MergeTagModal: FC<Props> = ({
  visible,
  sourceTag,
  onClose,
  onConfirm,
}) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'tag_info.merge',
  });
  const [targetTag, setTargetTag] = useState<SearchTagResp | null>(null);
  const [searchValue, setSearchValue] = useState('');
  const [tags, setTags] = useState<SearchTagResp[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [dropdownVisible, setDropdownVisible] = useState(false);
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const searchTags = async (search: string) => {
    try {
      const res = await request.get<SearchTagResp[]>(
        '/answer/api/v1/question/tags',
        {
          params: { tag: search },
        },
      );
      // Filter out the source tag from results
      const filteredTags = res.filter(
        (tag) => tag.slug_name !== sourceTag.slug_name,
      );
      setTags(filteredTags);
      if (filteredTags.length > 0 && isFocused) {
        setDropdownVisible(true);
      }
    } catch (error) {
      console.error('Failed to search tags:', error);
      setTags([]);
    }
  };

  // Debounced search function
  const debouncedSearch = useCallback(
    (() => {
      let timeout: number | undefined;
      return (search: string) => {
        if (timeout) {
          clearTimeout(timeout);
        }
        timeout = window.setTimeout(() => {
          searchTags(search);
        }, 1000);
      };
    })(),
    [isFocused],
  );

  useEffect(() => {
    if (visible) {
      searchTags('');
      setSearchValue('');
      setTargetTag(null);
      setCurrentIndex(0);
      setDropdownVisible(false);
      setIsFocused(false);
    }
  }, [visible]);

  const handleConfirm = () => {
    if (!targetTag) return;
    onConfirm(sourceTag.tag_id, targetTag.tag_id);
  };

  const handleSelect = (tag: SearchTagResp) => {
    setTargetTag(tag);
    setDropdownVisible(false);
    setSearchValue(tag.display_name);
    setIsFocused(false);
    inputRef.current?.blur();
  };

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    setSearchValue(value);
    if (value) {
      debouncedSearch(value);
    } else {
      searchTags('');
    }
  };

  const handleFocus = () => {
    setIsFocused(true);
    setDropdownVisible(true);
  };

  const handleBlur = () => {
    // Use setTimeout to allow click events on dropdown items to fire before closing
    setTimeout(() => {
      setIsFocused(false);
      if (!targetTag) {
        setDropdownVisible(false);
      }
    }, 200);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!tags.length) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setCurrentIndex((prev) => (prev < tags.length - 1 ? prev + 1 : prev));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setCurrentIndex((prev) => (prev > 0 ? prev - 1 : prev));
        break;
      case 'Enter':
        e.preventDefault();
        if (currentIndex >= 0 && currentIndex < tags.length) {
          handleSelect(tags[currentIndex]);
        }
        break;
      case 'Escape':
        e.preventDefault();
        inputRef.current?.blur();
        setDropdownVisible(false);
        break;
      default:
        break;
    }
  };

  return (
    <Modal
      title={t('title')}
      visible={visible}
      onCancel={onClose}
      onConfirm={handleConfirm}
      confirmText={t('btn_submit')}
      confirmBtnVariant="primary"
      cancelText={t('btn_close')}
      cancelBtnVariant="link"
      confirmBtnDisabled={!targetTag}>
      <Form>
        <Form.Group className="mb-3">
          <Form.Label>{t('source_tag_title')}</Form.Label>
          <Form.Control value={sourceTag.display_name} disabled />
          <Form.Text className="text-muted">
            {t('source_tag_description')}
          </Form.Text>
        </Form.Group>
        <Form.Group>
          <Form.Label>{t('target_tag_title')}</Form.Label>
          <Dropdown
            show={dropdownVisible && (tags.length > 0 || Boolean(searchValue))}
            onToggle={setDropdownVisible}>
            <Form.Control
              ref={inputRef}
              type="text"
              value={searchValue}
              onChange={handleSearch}
              onKeyDown={handleKeyDown}
              onFocus={handleFocus}
              onBlur={handleBlur}
              autoComplete="off"
            />
            <Dropdown.Menu className="w-100">
              {tags.map((tag, index) => (
                <Dropdown.Item
                  key={tag.slug_name}
                  active={index === currentIndex}
                  onClick={() => handleSelect(tag)}>
                  {tag.display_name}
                </Dropdown.Item>
              ))}
              {tags.length === 0 && searchValue && (
                <Dropdown.Item disabled>{t('no_results')}</Dropdown.Item>
              )}
            </Dropdown.Menu>
          </Dropdown>
          <Form.Text className="text-muted">
            {t('target_tag_description')}
          </Form.Text>
        </Form.Group>
      </Form>
    </Modal>
  );
};

export default MergeTagModal;

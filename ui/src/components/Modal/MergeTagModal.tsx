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

import { FC, useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Form, Dropdown } from 'react-bootstrap';

import debounce from 'lodash/debounce';

import { TagInfo } from '@/common/interface';
import { queryTags } from '@/services';

import Modal from './Modal';

const DEBOUNCE_DELAY = 300;

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
  const [hasSearched, setHasSearched] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const filteredTags = useMemo(() => {
    return tags.filter((tag) => tag.slug_name !== sourceTag.slug_name);
  }, [tags, sourceTag.slug_name]);

  const searchTags = useCallback(async (search: string) => {
    try {
      const res = await queryTags(search);
      setTags(res || []);
      setHasSearched(true);
    } catch (error) {
      console.error('Failed to search tags:', error);
      setTags([]);
      setHasSearched(true);
    }
  }, []);

  const debouncedSearch = useMemo(
    () => debounce(searchTags, DEBOUNCE_DELAY),
    [searchTags],
  );

  const handleConfirm = useCallback(() => {
    if (!targetTag) return;
    onConfirm(sourceTag.tag_id, targetTag.tag_id);
  }, [targetTag, sourceTag.tag_id, onConfirm]);

  const handleSelect = useCallback((tag: SearchTagResp) => {
    setTargetTag(tag);
    setDropdownVisible(false);
    setSearchValue(tag.display_name);
    setIsFocused(false);
    inputRef.current?.blur();
  }, []);

  const handleSearch = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const { value } = e.target;
      setSearchValue(value);
      setHasSearched(false);
      if (value) {
        debouncedSearch(value);
      } else {
        searchTags('');
      }
    },
    [debouncedSearch, searchTags],
  );

  const handleFocus = useCallback(() => {
    setIsFocused(true);
    setDropdownVisible(true);
  }, []);

  const handleBlur = useCallback(() => {
    setTimeout(() => {
      setIsFocused(false);
      if (!targetTag) {
        setDropdownVisible(false);
      }
    }, 200);
  }, [targetTag]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!filteredTags.length) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setCurrentIndex((prev) =>
            prev < filteredTags.length - 1 ? prev + 1 : prev,
          );
          break;
        case 'ArrowUp':
          e.preventDefault();
          setCurrentIndex((prev) => (prev > 0 ? prev - 1 : prev));
          break;
        case 'Enter':
          e.preventDefault();
          if (currentIndex >= 0 && currentIndex < filteredTags.length) {
            handleSelect(filteredTags[currentIndex]);
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
    },
    [filteredTags, currentIndex, handleSelect],
  );

  useEffect(() => {
    if (visible) {
      searchTags('');
      setSearchValue('');
      setTargetTag(null);
      setCurrentIndex(0);
      setDropdownVisible(false);
      setIsFocused(false);
      setHasSearched(false);
    }
    return () => {
      debouncedSearch.cancel();
    };
  }, [visible, searchTags, debouncedSearch]);

  useEffect(() => {
    if (filteredTags.length > 0 && isFocused) {
      setDropdownVisible(true);
    }
  }, [filteredTags, isFocused]);

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
          <div className="position-relative">
            <Dropdown
              show={
                dropdownVisible &&
                (filteredTags.length > 0 || Boolean(searchValue))
              }
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
              {filteredTags.length !== 0 && (
                <Dropdown.Menu className="w-100">
                  {filteredTags.map((tag, index) => (
                    <Dropdown.Item
                      key={tag.slug_name}
                      active={index === currentIndex}
                      onClick={() => handleSelect(tag)}>
                      {tag.display_name}
                    </Dropdown.Item>
                  ))}
                </Dropdown.Menu>
              )}
              {filteredTags.length === 0 && searchValue && hasSearched && (
                <Dropdown.Menu className="w-100">
                  <Dropdown.Item disabled>{t('no_results')}</Dropdown.Item>
                </Dropdown.Menu>
              )}
            </Dropdown>
          </div>
          <Form.Text className="text-muted">
            {t('target_tag_description')}
          </Form.Text>
        </Form.Group>
      </Form>
    </Modal>
  );
};

export default MergeTagModal;

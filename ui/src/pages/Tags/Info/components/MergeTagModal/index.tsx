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
import { Form, Dropdown, Modal, Button } from 'react-bootstrap';

import debounce from 'lodash/debounce';

import { TagInfo } from '@/common/interface';
import { queryTags } from '@/services';

import './index.scss';

const DEBOUNCE_DELAY = 400;

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
  const inputRef = useRef<HTMLInputElement>(null);

  const searchTags = useCallback(
    debounce((search) => {
      if (!search) {
        setTags([]);
        return;
      }
      queryTags(search).then((res) => {
        const filteredTags =
          res.filter((tag) => tag.slug_name !== sourceTag.slug_name) || [];
        setTags(filteredTags || []);
        setDropdownVisible(true);
      });
    }, DEBOUNCE_DELAY),
    [],
  );

  const handleConfirm = () => {
    if (!targetTag) return;
    onConfirm(sourceTag.tag_id, targetTag.tag_id);
  };

  const handleSelect = (tag: SearchTagResp) => {
    setTargetTag(tag);
    setDropdownVisible(false);
    setSearchValue(tag.display_name);
    inputRef.current?.blur();
  };

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const searchStr = e.currentTarget.value.trim();
    setSearchValue(searchStr);
    searchTags(searchStr);
  };

  const handleBlur = () => {
    setTimeout(() => {
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

  useEffect(() => {
    if (visible) {
      searchTags('');
      setSearchValue('');
      setTargetTag(null);
      setCurrentIndex(0);
      setDropdownVisible(false);
    }
  }, [visible]);

  return (
    <Modal show={visible} onCancel={onClose} className="mergeTagModal">
      <Modal.Header>
        <Modal.Title>{t('title')}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
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
                show={Boolean(dropdownVisible && searchValue)}
                onToggle={setDropdownVisible}>
                <Form.Control
                  ref={inputRef}
                  type="text"
                  value={searchValue}
                  onChange={handleSearch}
                  onKeyDown={handleKeyDown}
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
                  {!tags.length && searchValue && (
                    <Dropdown.Item disabled>{t('no_results')}</Dropdown.Item>
                  )}
                </Dropdown.Menu>
              </Dropdown>
            </div>
            <Form.Text className="text-muted">
              {t('target_tag_description')}
            </Form.Text>
          </Form.Group>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={onClose}>
          {t('close', { keyPrefix: 'btns' })}
        </Button>
        <Button onClick={handleConfirm} disabled={!targetTag}>
          {t('submit', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default MergeTagModal;

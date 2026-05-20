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

/* eslint-disable no-nested-ternary */
import { FC, useState, useEffect, useRef, useCallback } from 'react';
import { Dropdown, Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import debounce from 'lodash/debounce';
import { marked } from 'marked';
import classNames from 'classnames';

import { useTagModal, useToast } from '@/hooks';
import type * as Type from '@/common/interface';
import { queryTags, useUserPermission } from '@/services';
import { writeSettingStore } from '@/stores';

// import { OutsideClickListener } from '@/components';

import './index.scss';

interface IProps {
  value?: Type.Tag[];
  onChange?: (tags: Type.Tag[]) => void;
  hiddenDescription?: boolean;
  hiddenCreateBtn?: boolean;
  maxTagLength?: number;
  showRequiredTag?: boolean;
  autoFocus?: boolean;
  isInvalid?: boolean;
  tagStyleMode?: 'default' | 'simple';
  formText?: string;
  errMsg?: string;
}

let timer;

const TagSelector: FC<IProps> = ({
  value = [],
  onChange,
  hiddenDescription = false,
  hiddenCreateBtn = false,
  maxTagLength = 0,
  showRequiredTag = false,
  autoFocus = false,
  isInvalid = false,
  formText = '',
  tagStyleMode = 'default',
  errMsg = '',
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const [initialized, setInitialized] = useState(false);
  const [focusState, setFocusState] = useState(autoFocus);
  const [showMenu, setShowMenu] = useState(false);
  const [currentIndex, setCurrentIndex] = useState<number>(0);
  const [repeatIndex, setRepeatIndex] = useState(-1);
  const [searchValue, setSearchValue] = useState<string>('');
  const [tags, setTags] = useState<Type.Tag[] | null>([]);
  const [requiredTags, setRequiredTags] = useState<Type.Tag[] | null>(null);
  const writeInfo = writeSettingStore((state) => state.write);
  const { t } = useTranslation('translation', { keyPrefix: 'tag_selector' });
  const { data: userPermission } = useUserPermission('tag.add');
  const canAddTag =
    (maxTagLength > 0 && value?.length < maxTagLength) || maxTagLength === 0;
  const toast = useToast();
  const tagModal = useTagModal({
    onConfirm: (data) => {
      if (!(onChange instanceof Function)) {
        return;
      }
      const findIndex = value.findIndex(
        (item) => item.slug_name.toLowerCase() === data.slug_name.toLowerCase(),
      );
      if (findIndex === -1) {
        onChange([
          ...value,
          {
            ...data,
            parsed_text: marked(data.original_text),
          },
        ]);
        setSearchValue('');
      } else {
        setRepeatIndex(findIndex);
        clearTimeout(timer);
        timer = setTimeout(() => {
          setRepeatIndex(-1);
        }, 2000);
      }
    },
  });

  const filterTags = (result) => {
    const tagArray: Type.Tag[] = [];
    result?.forEach((item) => {
      const findIndex = value.findIndex((v) => {
        const tagName1 = v.slug_name.toLowerCase();
        const tagName2 =
          typeof item === 'string'
            ? item.toLowerCase()
            : item.slug_name.toLowerCase();

        return tagName1 === tagName2;
      });

      if (findIndex === -1) {
        tagArray.push(typeof item === 'string' ? { slug_name: item } : item);
      }
    });
    return tagArray;
  };

  const handleMenuShow = (bol: boolean) => {
    setShowMenu(bol);
    const ele = document.getElementById('a-dropdown-menu');
    if (ele) {
      if (bol) {
        ele.classList.add('show');
      } else {
        ele.classList.remove('show');
      }
    }
  };

  const handleTagSelectorFocus = () => {
    setFocusState(true);
    inputRef.current?.focus();
  };

  const handleTagSelectorBlur = () => {
    setFocusState(false);
    setCurrentIndex(0);
    handleMenuShow(false);
  };

  const fetchTags = useCallback(
    debounce((str) => {
      if (!showRequiredTag && !str) {
        setTags([]);
        return;
      }
      queryTags(str).then((res) => {
        const tagArray: Type.Tag[] = filterTags(res || []);
        if (str === '') {
          setRequiredTags(res);
        }
        handleMenuShow(tagArray.length > 0);
        setTags(tagArray);
      });
    }, 400),
    [],
  );

  const resetSearch = () => {
    setCurrentIndex(0);
    setSearchValue('');
    if (canAddTag) {
      const tagArray: Type.Tag[] = filterTags(requiredTags);
      setTags(tagArray.length > 0 ? tagArray : []);
    } else {
      setTags([]);
    }
  };
  const handleClick = (val: Type.Tag) => {
    const findIndex = value.findIndex(
      (item) => item.slug_name.toLowerCase() === val.slug_name.toLowerCase(),
    );
    if (onChange instanceof Function && findIndex === -1) {
      onChange([
        ...value,
        {
          original_text: '',
          parsed_text: '',
          ...val,
        },
      ]);
    } else {
      setRepeatIndex(findIndex);
      clearTimeout(timer);
      timer = setTimeout(() => {
        setRepeatIndex(-1);
      }, 2000);
    }
    resetSearch();
  };

  const handleRemove = (val: Type.Tag) => {
    if (onChange instanceof Function) {
      onChange(
        value.filter((v) => {
          if (v instanceof Object) {
            return v.slug_name.toLowerCase() !== val.slug_name.toLowerCase();
          }
          return v !== val;
        }),
      );
    }
  };

  const handleSearch = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const searchStr = e.currentTarget.value.replace(';', '');
    onChange?.([...value]);
    setSearchValue(searchStr);
    fetchTags(searchStr);
  };

  const scrollIntoView = (targetId) => {
    const container = document.getElementById('a-dropdown-menu') as HTMLElement;
    const ele = document.getElementById(targetId) as HTMLElement;
    if (ele?.offsetTop >= 104) {
      container.scrollTo({
        top: ele.offsetTop - 104,
        behavior: 'smooth',
      });
    }
  };

  const handleKeyDown = (e) => {
    e.stopPropagation();
    const { keyCode } = e;
    if (keyCode === 9) {
      // handleTagSelectorBlur();
      return;
    }
    if (value.length > 0 && keyCode === 8 && !searchValue) {
      handleRemove(value[value.length - 1]);
    }

    if (!tags) {
      return;
    }

    if (keyCode === 38 && currentIndex > 0) {
      scrollIntoView(tags[currentIndex - 1].slug_name);
      setCurrentIndex(currentIndex - 1);
    }
    if (keyCode === 40 && currentIndex < tags.length - 1) {
      scrollIntoView(tags[currentIndex + 1].slug_name);
      setCurrentIndex(currentIndex + 1);
    }

    if (keyCode === 13 && currentIndex > -1) {
      e.preventDefault();
      if (tags.length === 0 && searchValue) {
        tagModal.onShow(searchValue);
        return;
      }
      if (currentIndex <= tags.length - 1) {
        handleClick(tags[currentIndex]);
        // if (currentIndex === tags.length - 1 && currentIndex > 0) {
        //   setCurrentIndex(currentIndex - 1);
        // }
      }
    }
  };

  const handleCreate = () => {
    const tagAddPermission = userPermission?.['tag.add'];
    if (!tagAddPermission || tagAddPermission?.has_permission) {
      tagModal.onShow(searchValue);
    } else if (tagAddPermission?.no_permission_tip) {
      toast.onShow({
        msg: tagAddPermission.no_permission_tip,
        variant: 'danger',
      });
    }
  };

  const handleClickToggle = () => {
    const menuHasContent =
      (tags && tags?.length > 0) ||
      (searchValue && tags?.length === 0) ||
      (searchValue && !hiddenCreateBtn);
    if (canAddTag && menuHasContent) {
      handleMenuShow(true);
    } else {
      handleMenuShow(false);
    }
  };

  const handleTagHint = () => {
    if (!writeInfo || writeInfo.min_tags === undefined || !writeInfo.min_tags) {
      return t('hint_zero_tags');
    }

    if (writeInfo.min_tags === 1) {
      return t('hint');
    }

    return t(`hint_more_than_one_tag`, {
      min_tags_number: writeInfo.min_tags,
    });
  };

  useEffect(() => {
    if (canAddTag) {
      const tagArray: Type.Tag[] = filterTags(requiredTags);
      setTags(tagArray.length > 0 ? tagArray : []);
    } else {
      setTags([]);
    }
  }, [value]);

  useEffect(() => {
    if (focusState && showRequiredTag) {
      fetchTags(searchValue);
      inputRef.current?.focus();
    }
  }, [focusState]);

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
        handleTagSelectorBlur();
      }
    };
    document.addEventListener('click', handleOutsideClick);
    return () => {
      document.removeEventListener('click', handleOutsideClick);
    };
  }, [initialized]);

  useEffect(() => {
    // menu show
    const menuHasContent =
      (tags && tags?.length > 0) ||
      (searchValue && tags?.length === 0) ||
      (searchValue && !hiddenCreateBtn);
    if (focusState) {
      if (canAddTag && menuHasContent) {
        handleMenuShow(true);
      } else {
        handleMenuShow(false);
      }

      if ((tags && tags?.length < 5) || maxTagLength === 0) {
        inputRef.current?.focus();
      }
    }
  }, [focusState, tags, hiddenCreateBtn, searchValue, maxTagLength]);

  useEffect(() => {
    // set width of tag Form.Control
    const ele = document.querySelector('.a-input-width') as HTMLElement;
    const elePlaceholder = document.querySelector(
      '.a-placeholder-width',
    ) as HTMLElement;
    if (ele.offsetWidth > 60) {
      inputRef.current?.setAttribute(
        'style',
        `width:${ele.offsetWidth + 16}px`,
      );
    } else {
      inputRef.current?.setAttribute(
        'style',
        `width: ${elePlaceholder.offsetWidth + 7}px`,
      );
    }
  }, [searchValue]);

  return (
    <div ref={containerRef} className="position-relative">
      <div
        className={classNames(
          'tag-selector-wrap form-control position-relative p-0',
          focusState ? 'tag-selector-wrap--focus' : '',
          isInvalid ? 'is-invalid' : '',
        )}
        onClick={handleTagSelectorFocus}
        onKeyDown={handleKeyDown}>
        <div onClick={handleClickToggle}>
          <div
            className="d-flex flex-wrap m-n1"
            style={{ padding: '0.375rem 0.75rem' }}>
            {value?.map((item, index) => {
              return (
                <span
                  key={item.slug_name}
                  className={classNames(
                    'badge-tag rounded-1 m-1 flex-shrink-0',
                    tagStyleMode === 'default' &&
                      item.reserved &&
                      'badge-tag-reserved',
                    tagStyleMode === 'default' &&
                      item.recommend &&
                      'badge-tag-required',
                    index === repeatIndex && 'bg-fade-out',
                  )}>
                  {item.display_name}
                  <span
                    className="ms-1 hover-hand"
                    onMouseUp={() => handleRemove(item)}>
                    ×
                  </span>
                </span>
              );
            })}
            {canAddTag ? (
              <Form.Control
                // autoFocus
                autoComplete="off"
                style={{ width: '60px' }}
                ref={inputRef}
                className="a-input m-1"
                placeholder={t('add_btn')}
                value={searchValue}
                onChange={handleSearch}
                onFocus={handleTagSelectorFocus}
              />
            ) : (
              <Form.Control
                autoComplete="off"
                className="a-input"
                style={{ width: '60px', position: 'absolute', zIndex: -1 }}
                autoFocus
              />
            )}
            <span className="a-input-width">{searchValue}</span>
            <span className="a-placeholder-width">{t('add_btn')}</span>
          </div>
        </div>
        <Dropdown.Menu id="a-dropdown-menu" className="w-100" show={showMenu}>
          {!searchValue &&
            showRequiredTag &&
            tags &&
            tags.filter((v) => v.recommend)?.length > 0 && (
              <h6 className="dropdown-header">{t('tag_required_text')}</h6>
            )}

          {tags?.map((item, index) => {
            return (
              <Dropdown.Item
                key={item.slug_name}
                id={item.slug_name}
                active={index === currentIndex}
                onClick={() => handleClick(item)}>
                {item.display_name}
              </Dropdown.Item>
            );
          })}
          {searchValue && tags?.length === 0 && (
            <Dropdown.Item disabled className="text-secondary">
              {t('no_result')}
            </Dropdown.Item>
          )}
          {!hiddenCreateBtn && searchValue && (
            <Button
              variant="link"
              className="px-3 btn-no-border w-100 text-start"
              onClick={handleCreate}>
              + {t('create_btn')}
            </Button>
          )}
        </Dropdown.Menu>
      </div>
      {!hiddenDescription && (
        <Form.Text>{formText || handleTagHint()}</Form.Text>
      )}
      <Form.Control.Feedback type="invalid">{errMsg}</Form.Control.Feedback>
    </div>
  );
};

export default TagSelector;

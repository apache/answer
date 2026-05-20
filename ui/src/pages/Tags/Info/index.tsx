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

import { useState, useEffect } from 'react';
import { Alert, Row, Col, Button, Card } from 'react-bootstrap';
import { useParams, useNavigate, Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import classNames from 'classnames';

import { usePageTags } from '@/hooks';
import { Tag, TagSelector, FormatTime, Modal, htmlRender } from '@/components';
import {
  useTagInfo,
  useQuerySynonymsTags,
  saveSynonymsTags,
  deleteTag,
  editCheck,
  unDeleteTag,
  mergeTag,
} from '@/services';
import { pathFactory } from '@/router/pathFactory';
import { loggedUserInfoStore, toastStore } from '@/stores';

import MergeTagModal from './components/MergeTagModal';

const TagIntroduction = () => {
  const userInfo = loggedUserInfoStore((state) => state.user);
  const location = useLocation();
  const isLogged = Boolean(userInfo?.access_token);
  const [isEdit, setEditState] = useState(false);
  const { tagName } = useParams();
  const { data: tagInfo, mutate: refreshTagInfo } = useTagInfo({
    name: tagName,
  });
  const { t } = useTranslation('translation', { keyPrefix: 'tag_info' });
  const navigate = useNavigate();
  const { data: synonymsData, mutate } = useQuerySynonymsTags(
    tagInfo?.tag_id,
    tagInfo?.status,
  );
  const [showMergeModal, setShowMergeModal] = useState(false);

  let pageTitle = '';
  if (tagInfo) {
    pageTitle = `'${tagInfo.display_name}' ${t('tag_wiki', {
      keyPrefix: 'page_title',
    })}`;
  }
  usePageTags({
    title: pageTitle,
  });
  useEffect(() => {
    if (location.state?.isReview) {
      toastStore.getState().show({
        msg: t('review', { keyPrefix: 'toast' }),
        variant: 'warning',
      });

      // remove state isReview
      const newLocation = { ...location };
      delete newLocation.state;
      window.history.replaceState(null, '', newLocation.pathname);
    }
  }, [location.state]);

  useEffect(() => {
    const fmt = document.querySelector('.content.fmt') as HTMLElement;
    if (!fmt) {
      return;
    }
    htmlRender(fmt, {
      copySuccessText: t('copied', { keyPrefix: 'messages' }),
      copyText: t('copy', { keyPrefix: 'messages' }),
    });
  }, [tagInfo?.parsed_text]);

  if (!tagInfo) {
    return null;
  }
  if (tagInfo.main_tag_slug_name) {
    navigate(pathFactory.tagInfo(tagInfo.main_tag_slug_name), {
      replace: true,
    });
    return null;
  }

  const handleEdit = () => {
    setEditState(true);
  };

  const handleSave = () => {
    saveSynonymsTags({
      tag_id: tagInfo?.tag_id,
      synonym_tag_list: synonymsData?.synonyms,
    }).then(() => {
      mutate();
      setEditState(false);
    });
  };

  const handleTagsChange = (value) => {
    mutate(
      { ...synonymsData, synonyms: [...value] },
      {
        revalidate: false,
      },
    );
  };

  const handleEditTag = () => {
    editCheck(tagInfo?.tag_id).then(() => {
      navigate(pathFactory.tagEdit(tagInfo?.tag_id));
    });
  };
  const handleDeleteTag = () => {
    if (synonymsData?.synonyms && synonymsData.synonyms.length > 0) {
      Modal.confirm({
        title: t('delete.title'),
        content: t('delete.tip_with_synonyms'),
        showConfirm: false,
        cancelText: t('delete.close'),
      });
      return;
    }
    if (tagInfo.question_count > 0) {
      Modal.confirm({
        title: t('delete.title'),
        content: t('delete.tip_with_posts'),
        showConfirm: false,
        cancelText: t('delete.close'),
      });
      return;
    }

    Modal.confirm({
      title: t('delete.title'),
      content: t('delete.tip'),
      confirmText: t('delete', { keyPrefix: 'btns' }),
      confirmBtnVariant: 'danger',
      onConfirm: () => {
        deleteTag(tagInfo.tag_id).then(() => {
          navigate('/tags', { replace: true });
        });
      },
    });
  };
  const handleMergeTag = () => {
    setShowMergeModal(true);
  };

  const handleMergeConfirm = (sourceTagID: string, targetTagID: string) => {
    mergeTag({ source_tag_id: sourceTagID, target_tag_id: targetTagID }).then(
      () => {
        setShowMergeModal(false);
        navigate('/tags', { replace: true });
      },
    );
  };

  const onAction = (params) => {
    if (params.action === 'edit') {
      handleEditTag();
    }
    if (params.action === 'delete') {
      handleDeleteTag();
    }
    if (params.action === 'merge') {
      handleMergeTag();
    }
    if (params.action === 'undelete') {
      Modal.confirm({
        title: t('undelete_title', { keyPrefix: 'delete' }),
        content: t('undelete_desc', { keyPrefix: 'delete' }),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('undelete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          unDeleteTag(tagInfo.tag_id).then(() => {
            // undo
            refreshTagInfo();
          });
        },
      });
    }
  };

  return (
    <>
      <Row className="pt-4 mb-5">
        <Col className="page-main flex-auto">
          {tagInfo?.status === 'deleted' && (
            <Alert variant="danger" className="mb-4">
              {t('post_deleted', { keyPrefix: 'messages' })}
            </Alert>
          )}
          <h3 className="mb-3">
            <Link
              to={pathFactory.tagLanding(tagInfo.slug_name)}
              replace
              className="link-dark">
              {tagInfo.display_name}
            </Link>
          </h3>

          <div className="text-secondary mb-4 small">
            <FormatTime preFix={t('created_at')} time={tagInfo.created_at} />
            <FormatTime
              preFix={t('edited_at')}
              className="ms-3"
              time={tagInfo.updated_at}
            />
          </div>

          <div
            className="content text-break fmt"
            dangerouslySetInnerHTML={{ __html: tagInfo?.parsed_text }}
          />
          <div className="mt-4">
            {tagInfo?.member_actions.map((action, index) => {
              return (
                <Button
                  key={action.name}
                  variant="link"
                  size="sm"
                  className={classNames(
                    'link-secondary btn-no-border p-0',
                    index > 0 && 'ms-3',
                  )}
                  onClick={() => onAction(action)}>
                  {action.name}
                </Button>
              );
            })}
            {isLogged && (
              <Link
                to={`/tags/${tagInfo?.tag_id}/timeline`}
                className={classNames(
                  'link-secondary btn-no-border p-0 small',
                  tagInfo?.member_actions?.length > 0 && 'ms-3',
                )}>
                {t('history')}
              </Link>
            )}
          </div>
        </Col>
        <Col className="page-right-side mt-4 mt-xl-0">
          <Card>
            <Card.Header className="d-flex justify-content-between">
              <span>{t('synonyms.title')}</span>
              {isEdit ? (
                <Button
                  variant="link"
                  className="p-0 btn-no-border"
                  onClick={handleSave}>
                  {t('synonyms.btn_save')}
                </Button>
              ) : synonymsData?.member_actions?.find(
                  (v) => v.action === 'edit',
                ) ? (
                <Button
                  variant="link"
                  className="p-0 btn-no-border"
                  onClick={handleEdit}>
                  {t('synonyms.btn_edit')}
                </Button>
              ) : null}
            </Card.Header>
            <Card.Body>
              {isEdit && (
                <>
                  <div className="mb-3">
                    {t('synonyms.text')}{' '}
                    <Tag
                      data={{
                        slug_name: tagName || '',
                        main_tag_slug_name: '',
                        display_name:
                          tagInfo?.display_name || tagInfo?.slug_name || '',
                        recommend: false,
                        reserved: false,
                      }}
                    />
                  </div>
                  <TagSelector
                    value={synonymsData?.synonyms}
                    onChange={handleTagsChange}
                    hiddenDescription
                  />
                </>
              )}
              {!isEdit &&
                (synonymsData?.synonyms && synonymsData.synonyms.length > 0 ? (
                  <div className="m-n1">
                    {synonymsData.synonyms.map((item) => {
                      return (
                        <Tag key={item.tag_id} className="m-1" data={item} />
                      );
                    })}
                  </div>
                ) : (
                  <>
                    <div className="text-muted mb-3">{t('synonyms.empty')}</div>
                    {synonymsData?.member_actions?.find(
                      (v) => v.action === 'edit',
                    ) && (
                      <Button
                        variant="outline-primary"
                        size="sm"
                        onClick={handleEdit}>
                        {t('synonyms.btn_add')}
                      </Button>
                    )}
                  </>
                ))}
            </Card.Body>
          </Card>
        </Col>
      </Row>
      <MergeTagModal
        visible={showMergeModal}
        sourceTag={tagInfo}
        onClose={() => setShowMergeModal(false)}
        onConfirm={handleMergeConfirm}
      />
    </>
  );
};

export default TagIntroduction;

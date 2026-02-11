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
import { Table, Button, Nav, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { BaseUserCard, FormatTime, Pagination, Empty } from '@/components';
import { useToast } from '@/hooks';
import {
  getAiConfig,
  saveAiConfig,
  useQueryAdminConversationList,
} from '@/services';
import * as Type from '@/common/interface';

import DetailModal from './components/DetailModal';
import Action from './components/Action';

const getPromptByLang = (
  promptConfig: Type.AiConfig['prompt_config'] | undefined,
  lang: string,
) => {
  if (!promptConfig) {
    return '';
  }
  const isZh = lang?.toLowerCase().startsWith('zh');
  return isZh ? promptConfig.zh_cn || '' : promptConfig.en_us || '';
};

const Index = () => {
  const { t, i18n } = useTranslation('translation', {
    keyPrefix: 'admin.conversations',
  });
  const toast = useToast();
  const historyConfigRef = useRef<Type.AiConfig>();
  const [urlSearchParams] = useSearchParams();
  const curPage = Number(urlSearchParams.get('page') || '1');
  const PAGE_SIZE = 20;
  const [activeTab, setActiveTab] = useState<'conversations' | 'settings'>(
    'conversations',
  );
  const [isSaving, setIsSaving] = useState(false);
  const [promptForm, setPromptForm] = useState({
    value: '',
    isInvalid: false,
    errorMsg: '',
  });
  const [detailModal, setDetailModal] = useState({
    visible: false,
    id: '',
  });
  const {
    data: conversations,
    isLoading,
    mutate: refreshList,
  } = useQueryAdminConversationList({
    page: curPage,
    page_size: PAGE_SIZE,
  });
  const isZhLang = i18n.language?.toLowerCase().startsWith('zh');

  const getAiConfigData = async () => {
    const aiConfig = await getAiConfig();
    historyConfigRef.current = aiConfig;
    setPromptForm({
      value: getPromptByLang(aiConfig.prompt_config, i18n.language),
      isInvalid: false,
      errorMsg: '',
    });
  };

  const handleSavePrompt = (evt: FormEvent) => {
    evt.preventDefault();
    if (!historyConfigRef.current || isSaving) {
      return;
    }
    setIsSaving(true);
    setPromptForm((prev) => ({
      ...prev,
      isInvalid: false,
      errorMsg: '',
    }));

    const params: Type.AiConfig = {
      enabled: historyConfigRef.current.enabled || false,
      chosen_provider: historyConfigRef.current.chosen_provider || '',
      ai_providers: historyConfigRef.current.ai_providers || [],
      prompt_config: {
        zh_cn: isZhLang
          ? promptForm.value
          : historyConfigRef.current.prompt_config?.zh_cn || '',
        en_us: isZhLang
          ? historyConfigRef.current.prompt_config?.en_us || ''
          : promptForm.value,
      },
    };

    saveAiConfig(params)
      .then(() => {
        historyConfigRef.current = params;
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
    if (activeTab === 'settings') {
      getAiConfigData();
    }
  }, [activeTab]);

  useEffect(() => {
    if (!historyConfigRef.current || activeTab !== 'settings') {
      return;
    }
    setPromptForm((prev) => ({
      ...prev,
      value: getPromptByLang(
        historyConfigRef.current?.prompt_config,
        i18n.language,
      ),
      isInvalid: false,
      errorMsg: '',
    }));
  }, [i18n.language, activeTab]);

  const handleShowDetailModal = (data) => {
    setDetailModal({
      visible: true,
      id: data.id,
    });
  };

  const handleHideDetailModal = () => {
    setDetailModal({
      visible: false,
      id: '',
    });
  };

  return (
    <div className="d-flex flex-column flex-grow-1 position-relative">
      <h3 className="mb-4">{t('ai_assistant', { keyPrefix: 'nav_menus' })}</h3>
      <Nav variant="underline" className="mb-4 border-bottom">
        <Nav.Item>
          <Nav.Link
            className="px-0 me-4"
            active={activeTab === 'conversations'}
            onClick={() => setActiveTab('conversations')}>
            {t('tabs.conversations')}
          </Nav.Link>
        </Nav.Item>
        <Nav.Item>
          <Nav.Link
            className="px-0"
            active={activeTab === 'settings'}
            onClick={() => setActiveTab('settings')}>
            {t('tabs.settings')}
          </Nav.Link>
        </Nav.Item>
      </Nav>

      {activeTab === 'conversations' && (
        <>
          <Table responsive="md">
            <thead>
              <tr>
                <th className="min-w-15">{t('topic')}</th>
                <th style={{ width: '10%' }}>{t('helpful')}</th>
                <th style={{ width: '10%' }}>{t('unhelpful')}</th>
                <th style={{ width: '20%' }}>{t('created')}</th>
                <th style={{ width: '10%' }} className="text-end">
                  {t('action')}
                </th>
              </tr>
            </thead>
            <tbody className="align-middle">
              {conversations?.list.map((item) => {
                return (
                  <tr key={item.id}>
                    <td>
                      <Button
                        variant="link"
                        className="p-0 text-decoration-none text-truncate max-w-30"
                        onClick={() => handleShowDetailModal(item)}>
                        {item.topic}
                      </Button>
                    </td>
                    <td>{item.helpful_count}</td>
                    <td>{item.unhelpful_count}</td>
                    <td>
                      <div className="vstack">
                        <BaseUserCard data={item.user_info} avatarSize="20px" />
                        <FormatTime
                          className="small text-secondary"
                          time={item.created_at}
                        />
                      </div>
                    </td>
                    <td className="text-end">
                      <Action id={item.id} refreshList={refreshList} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </Table>
          {!isLoading && Number(conversations?.count) <= 0 && (
            <Empty>{t('empty')}</Empty>
          )}
          <div className="mt-4 mb-2 d-flex justify-content-center">
            <Pagination
              currentPage={curPage}
              totalSize={conversations?.count || 0}
              pageSize={PAGE_SIZE}
            />
          </div>
        </>
      )}

      {activeTab === 'settings' && (
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
      )}
      <DetailModal
        visible={detailModal.visible}
        id={detailModal.id}
        onClose={handleHideDetailModal}
      />
    </div>
  );
};
export default Index;

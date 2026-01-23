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

import { useEffect, useState, useRef } from 'react';
import { Form, InputGroup, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import {
  getAiConfig,
  useQueryAiProvider,
  checkAiConfig,
  saveAiConfig,
} from '@/services';
import { aiControlStore } from '@/stores';
import { handleFormError } from '@/utils';
import { useToast } from '@/hooks';
import * as Type from '@/common/interface';

const Index = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.ai_settings',
  });
  const toast = useToast();
  const historyConfigRef = useRef<Type.AiConfig>();
  // const [historyConfig, setHistoryConfig] = useState<Type.AiConfig>();
  const { data: aiProviders } = useQueryAiProvider();

  const [formData, setFormData] = useState({
    enabled: {
      value: false,
      isInvalid: false,
      errorMsg: '',
    },
    provider: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    api_host: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    api_key: {
      value: '',
      isInvalid: false,
      isValid: false,
      errorMsg: '',
    },
    model: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });
  const [apiHostPlaceholder, setApiHostPlaceholder] = useState('');
  const [modelsData, setModels] = useState<{ id: string }[]>([]);
  const [isChecking, setIsChecking] = useState(false);

  const getCurrentProviderData = (provider) => {
    const findHistoryProvider =
      historyConfigRef.current?.ai_providers.find(
        (v) => v.provider === provider,
      ) || historyConfigRef.current?.ai_providers[0];

    return findHistoryProvider;
  };

  const checkAiConfigData = (data) => {
    const params = data || {
      api_host: formData.api_host.value || apiHostPlaceholder,
      api_key: formData.api_key.value,
    };
    setIsChecking(true);

    checkAiConfig(params)
      .then((res) => {
        setModels(res);
        const findHistoryProvider = getCurrentProviderData(
          formData.provider.value,
        );

        setIsChecking(false);

        if (!data) {
          setFormData({
            ...formData,
            api_key: {
              ...formData.api_key,
              errorMsg: t('api_key.check_success'),
              isInvalid: false,
              isValid: true,
            },
            model: {
              value: findHistoryProvider?.model || res[0].id,
              errorMsg: '',
              isInvalid: false,
            },
          });
        }
      })
      .catch((err) => {
        console.error('Checking AI config:', err);
        setIsChecking(false);
      });
  };

  const handleProviderChange = (value) => {
    const findHistoryProvider = getCurrentProviderData(value);
    setFormData({
      ...formData,
      provider: {
        value,
        isInvalid: false,
        errorMsg: '',
      },
      api_host: {
        value: findHistoryProvider?.api_host || '',
        isInvalid: false,
        errorMsg: '',
      },
      api_key: {
        value: findHistoryProvider?.api_key || '',
        isInvalid: false,
        isValid: false,
        errorMsg: '',
      },
      model: {
        value: findHistoryProvider?.model || '',
        isInvalid: false,
        errorMsg: '',
      },
    });
    const provider = aiProviders?.find((item) => item.name === value);
    const host = findHistoryProvider?.api_host || provider?.default_api_host;
    if (findHistoryProvider?.model) {
      checkAiConfigData({
        api_host: host,
        api_key: findHistoryProvider.api_key,
      });
    } else {
      setModels([]);
    }
  };

  const handleValueChange = (value) => {
    setFormData((prev) => ({
      ...prev,
      ...value,
    }));
  };

  const checkValidate = () => {
    let bol = true;

    const { api_host, api_key, model } = formData;

    if (!api_host.value) {
      bol = false;
      formData.api_host = {
        value: '',
        isInvalid: true,
        errorMsg: t('api_host.msg'),
      };
    }

    if (!api_key.value) {
      bol = false;
      formData.api_key = {
        value: '',
        isInvalid: true,
        isValid: false,
        errorMsg: t('api_key.msg'),
      };
    }

    if (!model.value) {
      bol = false;
      formData.model = {
        value: '',
        isInvalid: true,
        errorMsg: t('model.msg'),
      };
    }

    setFormData({
      ...formData,
    });

    return bol;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!checkValidate()) {
      return;
    }
    const newProviders = historyConfigRef.current?.ai_providers.map((v) => {
      if (v.provider === formData.provider.value) {
        return {
          provider: formData.provider.value,
          api_host: formData.api_host.value,
          api_key: formData.api_key.value,
          model: formData.model.value,
        };
      }
      return v;
    });

    const params = {
      enabled: formData.enabled.value,
      chosen_provider: formData.provider.value,
      ai_providers: newProviders,
    };
    saveAiConfig(params)
      .then(() => {
        aiControlStore.getState().update({
          ai_enabled: formData.enabled.value,
        });

        historyConfigRef.current = {
          ...params,
          ai_providers: params.ai_providers || [],
        };

        toast.onShow({
          msg: t('add_success'),
          variant: 'success',
        });
      })
      .catch((err) => {
        const data = handleFormError(err, formData);
        setFormData({ ...data });
        const ele = document.getElementById(err.list[0].error_field);
        ele?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      });
  };

  const getAiConfigData = async () => {
    const aiConfig = await getAiConfig();
    historyConfigRef.current = aiConfig;

    const currentAiConfig = getCurrentProviderData(aiConfig.chosen_provider);
    if (currentAiConfig?.model) {
      const provider = aiProviders?.find(
        (item) => item.name === formData.provider.value,
      );
      const host = currentAiConfig.api_host || provider?.default_api_host;
      checkAiConfigData({
        api_host: host,
        api_key: currentAiConfig.api_key,
      });
    }

    setFormData({
      enabled: {
        value: aiConfig.enabled || false,
        isInvalid: false,
        errorMsg: '',
      },
      provider: {
        value: currentAiConfig?.provider || '',
        isInvalid: false,
        errorMsg: '',
      },
      api_host: {
        value: currentAiConfig?.api_host || '',
        isInvalid: false,
        errorMsg: '',
      },
      api_key: {
        value: currentAiConfig?.api_key || '',
        isInvalid: false,
        isValid: false,
        errorMsg: '',
      },
      model: {
        value: currentAiConfig?.model || '',
        isInvalid: false,
        errorMsg: '',
      },
    });
  };

  useEffect(() => {
    getAiConfigData();
  }, []);

  useEffect(() => {
    if (formData.provider.value) {
      const provider = aiProviders?.find(
        (item) => item.name === formData.provider.value,
      );
      if (provider) {
        setApiHostPlaceholder(provider.default_api_host || '');
      }
    }
    if (!formData.provider.value && aiProviders) {
      setFormData((prev) => ({
        ...prev,
        provider: {
          ...prev.provider,
          value: aiProviders[0].name,
        },
      }));
    }
  }, [aiProviders, formData]);

  return (
    <div>
      <h3 className="mb-4">{t('ai_settings', { keyPrefix: 'nav_menus' })}</h3>
      <div className="max-w-748">
        <Form noValidate onSubmit={handleSubmit}>
          <Form.Group className="mb-3" controlId="enabled">
            <Form.Label>{t('enabled.label')}</Form.Label>
            <Form.Switch
              type="switch"
              id="enabled"
              label={t('enabled.check')}
              checked={formData.enabled.value}
              onChange={(e) =>
                handleValueChange({
                  enabled: {
                    value: e.target.checked,
                    errorMsg: '',
                    isInvalid: false,
                  },
                })
              }
            />
            <Form.Text className="text-muted">{t('enabled.text')}</Form.Text>
            <Form.Control.Feedback type="invalid">
              {formData.enabled.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="provider">
            <Form.Label>{t('provider.label')}</Form.Label>
            <Form.Select
              isInvalid={formData.provider.isInvalid}
              value={formData.provider.value}
              onChange={(e) => handleProviderChange(e.target.value)}>
              {aiProviders?.map((provider) => (
                <option key={provider.name} value={provider.name}>
                  {provider.display_name}
                </option>
              ))}
            </Form.Select>
            <Form.Control.Feedback type="invalid">
              {formData.provider.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="api_host">
            <Form.Label>{t('api_host.label')}</Form.Label>
            <Form.Control
              type="text"
              autoComplete="off"
              placeholder={apiHostPlaceholder}
              isInvalid={formData.api_host.isInvalid}
              value={formData.api_host.value}
              onChange={(e) =>
                handleValueChange({
                  api_host: {
                    value: e.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                })
              }
            />
            <Form.Control.Feedback type="invalid">
              {formData.api_host.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group className="mb-3" controlId="api_key">
            <Form.Label>{t('api_key.label')}</Form.Label>
            <InputGroup>
              <Form.Control
                type="password"
                autoComplete="new-password"
                isInvalid={formData.api_key.isInvalid}
                isValid={formData.api_key.isValid}
                value={formData.api_key.value}
                onChange={(e) =>
                  handleValueChange({
                    api_key: {
                      value: e.target.value,
                      errorMsg: '',
                      isInvalid: false,
                      isValid: false,
                    },
                  })
                }
              />
              <Button
                variant="outline-secondary"
                className="rounded-end"
                onClick={() => checkAiConfigData(null)}
                disabled={isChecking}>
                {t('api_key.check')}
              </Button>
              <Form.Control.Feedback
                type={formData.api_key.isValid ? 'valid' : 'invalid'}>
                {formData.api_key.errorMsg}
              </Form.Control.Feedback>
            </InputGroup>
          </Form.Group>

          <div className="mb-3">
            <label htmlFor="model" className="form-label">
              {t('model.label')}
            </label>
            {/* <Form.Select
            list="datalistOptions"
            isInvalid={formData.model.isInvalid}
            value={formData.model.value}
            onChange={(e) =>
              handleValueChange({
                model: {
                  value: e.target.value,
                  errorMsg: '',
                  isInvalid: false,
                },
              })
            }>
            {modelsData?.map((model) => {
              return (
                <option key={model.id} value={model.id}>
                  {model.id}
                </option>
              );
            })}
          </Form.Select> */}
            <input
              className="form-control"
              list="datalistOptions"
              id="model"
              value={formData.model.value}
              onChange={(e) =>
                handleValueChange({
                  model: {
                    value: e.target.value,
                    errorMsg: '',
                    isInvalid: false,
                  },
                })
              }
            />
            <datalist id="datalistOptions">
              {modelsData?.map((model) => {
                return (
                  <option key={model.id} value={model.id}>
                    {model.id}
                  </option>
                );
              })}
            </datalist>

            <div className="invalid-feedback">{formData.model.errorMsg}</div>
          </div>

          <Button type="submit">{t('save', { keyPrefix: 'btns' })}</Button>
        </Form>
      </div>
    </div>
  );
};
export default Index;

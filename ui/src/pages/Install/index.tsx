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

import { FC, useState, useEffect } from 'react';
import { Container, Row, Col, Card, Alert } from 'react-bootstrap';
import { useTranslation, Trans } from 'react-i18next';
import { Helmet, HelmetProvider } from 'react-helmet-async';

import type { FormDataType } from '@/common/interface';
import {
  dbCheck,
  installInit,
  installBaseInfo,
  checkConfigFileExists,
} from '@/services';
import {
  Storage,
  handleFormError,
  scrollToDocTop,
  scrollToElementTop,
} from '@/utils';
import { CURRENT_LANG_STORAGE_KEY } from '@/common/constants';
import { BASE_ORIGIN } from '@/router/alias';
import { setupInstallLanguage } from '@/utils/localize';
import { loggedUserInfoStore } from '@/stores';

import {
  FirstStep,
  SecondStep,
  ThirdStep,
  FourthStep,
  Fifth,
} from './components';

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'install' });
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [errorData, setErrorData] = useState<{ [propName: string]: any }>({
    msg: '',
  });
  const [checkData, setCheckData] = useState({
    db_table_exist: false,
    db_connection_success: false,
  });

  const [formData, setFormData] = useState<FormDataType>({
    lang: {
      value: 'en_US',
      isInvalid: false,
      errorMsg: '',
    },
    db_type: {
      value: 'mysql',
      isInvalid: false,
      errorMsg: '',
    },
    db_username: {
      value: 'root',
      isInvalid: false,
      errorMsg: '',
    },
    db_password: {
      value: 'root',
      isInvalid: false,
      errorMsg: '',
    },
    db_host: {
      value: 'db:3306',
      isInvalid: false,
      errorMsg: '',
    },
    db_name: {
      value: 'answer',
      isInvalid: false,
      errorMsg: '',
    },
    db_file: {
      value: '/data/answer.db',
      isInvalid: false,
      errorMsg: '',
    },
    site_name: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    site_url: {
      value: BASE_ORIGIN,
      isInvalid: false,
      errorMsg: '',
    },
    contact_email: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    login_required: {
      value: false,
      isInvalid: false,
      errorMsg: '',
    },
    external_content_display: {
      value: 'always_display',
      isInvalid: false,
      errorMsg: '',
    },
    name: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    password: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    confirm_password: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    email: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    ssl_enabled: {
      value: false,
      isInvalid: false,
      errorMsg: '',
    },
    ssl_mode: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    ssl_key: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    ssl_root_cert: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
    ssl_cert: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });

  const { update, user } = loggedUserInfoStore();

  const updateFormData = (params: FormDataType) => {
    if (Object.keys(params)?.[0] === 'db_type') {
      let updatedFormData = formData;
      if (params.db_type.value === 'mysql') {
        updatedFormData = {
          ...updatedFormData,
          db_username: { ...updatedFormData.db_username, value: 'root' },
          db_password: { ...updatedFormData.db_password, value: 'root' },
          db_host: { ...updatedFormData.db_host, value: 'db:3306' },
        };
      } else if (params.db_type.value === 'postgres') {
        updatedFormData = {
          ...updatedFormData,
          db_username: { ...updatedFormData.db_username, value: 'postgres' },
          db_password: { ...updatedFormData.db_password, value: 'postgres' },
          db_host: { ...updatedFormData.db_host, value: 'db:5432' },
          ssl_enabled: { ...updatedFormData.ssl_enabled, value: false },
          ssl_mode: { ...updatedFormData.ssl_mode, value: '' },
        };
      }
      return updatedFormData;
    }
    return formData;
  };

  const handleChange = (params: FormDataType) => {
    setErrorData({
      msg: '',
    });
    const updatedFormData = updateFormData(params);
    setFormData({ ...updatedFormData, ...params });
  };

  const handleErr = (data) => {
    scrollToDocTop();
    setErrorData(data);
  };

  const handleNext = async () => {
    setErrorData({
      msg: '',
    });
    setStep((pre) => pre + 1);
  };

  const checkInstall = () => {
    const params = {
      lang: formData.lang.value,
      db_type: formData.db_type.value,
      db_username: formData.db_username.value,
      db_password: formData.db_password.value,
      db_host: formData.db_host.value,
      db_name: formData.db_name.value,
      db_file: formData.db_file.value,
      ssl_enabled: formData.ssl_enabled.value,
      ssl_mode: formData.ssl_mode.value,
      ssl_key: formData.ssl_key.value,
      ssl_root_cert: formData.ssl_root_cert.value,
      ssl_cert: formData.ssl_cert.value,
    };
    installInit(params)
      .then(() => {
        handleNext();
      })
      .catch((err) => {
        handleErr(err);
      });
  };

  const submitDatabaseForm = () => {
    const params = {
      lang: formData.lang.value,
      db_type: formData.db_type.value,
      db_username: formData.db_username.value,
      db_password: formData.db_password.value,
      db_host: formData.db_host.value,
      db_name: formData.db_name.value,
      db_file: formData.db_file.value,
      ssl_enabled: formData.ssl_enabled.value,
      ssl_mode: formData.ssl_mode.value,
      ssl_key: formData.ssl_key.value,
      ssl_root_cert: formData.ssl_root_cert.value,
      ssl_cert: formData.ssl_cert.value,
    };
    dbCheck(params)
      .then(() => {
        checkInstall();
      })
      .catch((err) => {
        handleErr(err);
      });
  };

  const submitSiteConfig = () => {
    const params = {
      lang: formData.lang.value,
      site_name: formData.site_name.value,
      site_url: formData.site_url.value,
      contact_email: formData.contact_email.value,
      login_required: formData.login_required.value,
      external_content_display: formData.external_content_display.value,
      name: formData.name.value,
      password: formData.password.value,
      email: formData.email.value,
    };

    installBaseInfo(params)
      .then(() => {
        handleNext();
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        } else {
          handleErr(err);
        }
      });
  };

  const handleStep = () => {
    if (step === 1) {
      Storage.set(CURRENT_LANG_STORAGE_KEY, formData.lang.value);
      update({
        ...user,
        language: formData.lang.value,
      });
      handleNext();
    }
    if (step === 2) {
      submitDatabaseForm();
    }
    if (step === 3) {
      if (errorData.msg) {
        checkInstall();
      } else {
        handleNext();
      }
    }
    if (step === 4) {
      submitSiteConfig();
    }
    if (step > 4) {
      handleNext();
    }
  };

  const handleInstallNow = (e) => {
    e.preventDefault();
    if (checkData.db_table_exist) {
      setStep(8);
    } else {
      setStep(4);
    }
  };

  const configYmlCheck = () => {
    checkConfigFileExists()
      .then((res) => {
        setCheckData({
          db_table_exist: res.db_table_exist,
          db_connection_success: res.db_connection_success,
        });
        if (res && res.config_file_exist) {
          setupInstallLanguage(Storage.get(CURRENT_LANG_STORAGE_KEY));
          if (res.db_connection_success) {
            setStep(6);
          } else {
            setStep(7);
          }
        }
      })
      .finally(() => {
        setLoading(false);
      });
  };

  useEffect(() => {
    configYmlCheck();
  }, []);

  if (loading) {
    return <div />;
  }

  return (
    <HelmetProvider>
      <Helmet>
        <title>{t('install', { keyPrefix: 'page_title' })}</title>
      </Helmet>
      <div className="bg-f5 py-5 flex-grow-1">
        <Container className="py-3">
          <Row className="justify-content-center">
            <Col lg={6}>
              <h2 className="mb-4 text-center">{t('title')}</h2>
              <Card>
                <Card.Body>
                  {errorData?.msg && (
                    <Alert variant="danger">{errorData?.msg}</Alert>
                  )}

                  <FirstStep
                    visible={step === 1}
                    data={formData.lang}
                    changeCallback={handleChange}
                    nextCallback={handleStep}
                  />

                  <SecondStep
                    visible={step === 2}
                    data={formData}
                    changeCallback={handleChange}
                    nextCallback={handleStep}
                  />

                  <ThirdStep
                    visible={step === 3}
                    nextCallback={handleStep}
                    errorMsg={errorData}
                  />

                  <FourthStep
                    visible={step === 4}
                    data={formData}
                    changeCallback={handleChange}
                    nextCallback={handleStep}
                  />

                  <Fifth
                    visible={step === 5}
                    siteUrl={formData.site_url.value}
                  />
                  {step === 6 && (
                    <div>
                      <h5>{t('warn_title')}</h5>
                      <p>
                        <Trans
                          i18nKey="install.warn_desc"
                          components={{ 1: <code /> }}
                        />{' '}
                        <Trans i18nKey="install.install_now">
                          You may try
                          <a href="###" onClick={(e) => handleInstallNow(e)}>
                            installing now
                          </a>
                          .
                        </Trans>
                      </p>
                    </div>
                  )}

                  {step === 7 && (
                    <div>
                      <h5>{t('db_failed')}</h5>
                      <p>
                        <Trans
                          i18nKey="install.db_failed_desc"
                          components={{ 1: <code /> }}
                        />
                      </p>
                    </div>
                  )}

                  {step === 8 && (
                    <div>
                      <h5>{t('installed')}</h5>
                      <p>{t('installed_desc')}</p>
                    </div>
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </Container>
      </div>
    </HelmetProvider>
  );
};

export default Index;

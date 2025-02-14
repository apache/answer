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

import { FC, FormEvent} from 'react';
import { Form, Button, InputGroup } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import Progress from '../Progress';
import type { FormDataType } from '@/common/interface';

interface Props {
  data: FormDataType;
  changeCallback: (value: FormDataType) => void;
  nextCallback: () => void;
  visible: boolean;
}

const sqlData = [
  {
    value: 'mysql',
    label: 'MariaDB/MySQL',
  },
  {
    value: 'sqlite3',
    label: 'SQLite',
  },
  {
    value: 'postgres',
    label: 'PostgreSQL',
  },
];

const sslModes = [
  {
    value: 'require',
  },
  {
    value: 'verify-ca',
  },
];

const Index: FC<Props> = ({ visible, data, changeCallback, nextCallback }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'install' });

  const checkValidated = (): boolean => {
    let bol = true;
    const { db_type, db_username, db_password, db_host, db_name, db_file,ssl_enabled,ssl_mode,key_file,cert_file,pem_file } =
      data;

    if (db_type.value !== 'sqlite3') {
      if (!db_username.value) {
        bol = false;
        data.db_username = {
          value: '',
          isInvalid: true,
          errorMsg: t('db_username.msg'),
        };
      }

      if (!db_password.value) {
        bol = false;
        data.db_password = {
          value: '',
          isInvalid: true,
          errorMsg: t('db_password.msg'),
        };
      }

      if (!db_host.value) {
        bol = false;
        data.db_host = {
          value: '',
          isInvalid: true,
          errorMsg: t('db_host.msg'),
        };
      }  
      if (!ssl_enabled.value) {
        bol = false;
        data.ssl_enabled = {
          value: '',
          isInvalid: true,
          errorMsg: t('ssl_enabled.msg'),
        };
      }
      if (!ssl_mode.value) {
        bol = false;
        data.ssl_mode = {
          value: '',
          isInvalid: true,
         errorMsg: '',
        };
      }
    if (ssl_mode.value ==="verify-ca") {
      if (!key_file.value) {
        bol = false;
        data.key_file = {
          value: '',
          isInvalid: true,
          errorMsg: t('key_file.msg'),
        };
      }
      if (!pem_file.value) {
        bol = false;
        data.pem_file = {
          value: '',
          isInvalid: true,
          errorMsg: t('pem_file.msg'),
        };
      }
      if (!cert_file.value) {
        bol = false;
        data.cert_file = {
          value: '',
          isInvalid: true,
          errorMsg: t('cert_file.msg'),
        };
      }
    }
      if (!db_name.value) {
        bol = false;
        data.db_name = {
          value: '',
          isInvalid: true,
          errorMsg: t('db_name.msg'),
        };
      }
    } else if (!db_file.value) {
      bol = false;
      data.db_file = {
        value: '',
        isInvalid: true,
        errorMsg: t('db_file.msg'),
      };
    }
    changeCallback({
      ...data,
    });
    return bol;
  };

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    nextCallback();
  };
 


  if (!visible) return null;
  return (
    <Form noValidate onSubmit={handleSubmit}>
      <Form.Group controlId="database_engine" className="mb-3">
        <Form.Label>{t('db_type.label')}</Form.Label>
        <Form.Select
          value={data.db_type.value}
          isInvalid={data.db_type.isInvalid}
          onChange={(e) => {
            changeCallback({
              db_type: {
                value: e.target.value,
                isInvalid: false,
                errorMsg: '',
              },
            });
          }}>
          {sqlData.map((item) => {
            return (
              <option key={item.value} value={item.value}>
                {item.label}
              </option>
            );
          })}
        </Form.Select>
      </Form.Group>
      {data.db_type.value !== 'sqlite3' ? (
        <>
          <Form.Group controlId="username" className="mb-3">
            <Form.Label>{t('db_username.label')}</Form.Label>
            <Form.Control
              required
              placeholder={t('db_username.placeholder')}
              value={data.db_username.value}
              isInvalid={data.db_username.isInvalid}
              onChange={(e) => {
                changeCallback({
                  db_username: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                });
              }}
            />
            <Form.Control.Feedback type="invalid">
              {data.db_username.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group controlId="db_password" className="mb-3">
            <Form.Label>{t('db_password.label')}</Form.Label>
            <Form.Control
              required
              value={data.db_password.value}
              isInvalid={data.db_password.isInvalid}
              onChange={(e) => {
                changeCallback({
                  db_password: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                });
              }}
            />
            <Form.Control.Feedback type="invalid">
              {data.db_password.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>
            {data.db_type.value === 'postgres' && (
              <Form.Group controlId="ssl_enabled" className="conditional-checkbox">
                <Form.Check type="checkbox" id="sslEnabled">
                  <Form.Check.Input
                    type="checkbox"
                    value={data.ssl_enabled.value}
                    onChange={(e) => {
                    changeCallback({
                      ssl_enabled: {
                        value: e.target.checked,
                        isInvalid: false,
                        errorMsg: '',
                      },
                      });
                    }}
                  />
                 <Form.Label htmlFor="ssl_enabled">{t('ssl_enabled.label')}</Form.Label>
                </Form.Check>
              </Form.Group>
              )
              }
                {data.db_type.value === 'postgres' && data.ssl_enabled.value && (
                  <Form.Group controlId="sslmodeOptionsDropdown" className="mb-3">
                      <Form.Label>{t('ssl_mode.label')}</Form.Label>
                          <Form.Select
                            value={data.ssl_mode.value}
                            onChange={(e) => {
                              changeCallback({
                                ssl_mode: {
                                  value: e.target.value,
                                  isInvalid: false,
                                  errorMsg: '',
                                },
                              });
                            }}>
                            {sslModes.map((item) => {
                              return(
                                <option value={item.value} >
                                  {item.value}
                                </option>
                              );
                            })}
                          </Form.Select>
                          </Form.Group>
                  )}
 <br/>
                          {data.db_type.value === 'postgres' && data.ssl_enabled.value &&  data.ssl_mode.value === 'verify-ca'   && (
                            
                           <InputGroup className="mb-3">
                              <Form.Control
                                placeholder={t('key_file.placeholder')}
                                aria-label="key_file"
                                aria-describedby="basic-addon1"
                                // value={data.key_file.value}
                                                  onChange={(e) => {
                                                    changeCallback({
                                                      key_file: {
                                                        value: e.target.value,
                                                        isInvalid: false,
                                                        errorMsg: '',
                                                      },
                                                    });
                                                  }}
                              />
                              
                              <Form.Control
                                placeholder={t('cert_file.placeholder')} 
                                aria-label="cert_file"
                                aria-describedby="basic-addon1"
                                // value={data.cert_file.value}
                                                  onChange={(e) => {
                                                    changeCallback({
                                                      cert_file: {
                                                        value: e.target.value,
                                                        isInvalid: false,
                                                        errorMsg: '',
                                                      },
                                                    });
                                                  }}
                              />
                              
                              <Form.Control
                                placeholder={t('pem_file.placeholder')}
                                aria-label="pem_file"
                                aria-describedby="basic-addon1"
                                // value={data.pem_file.value }
                                                  onChange={(e) => {
                                                    changeCallback({
                                                      pem_file: {
                                                        value: e.target.value,
                                                        isInvalid: false,
                                                        errorMsg: '',
                                                      },
                                                    });
                                                  }}
                              />
                            </InputGroup>
                
                                               ) }
          <Form.Group controlId="db_host" className="mb-3">
            <Form.Label>{t('db_host.label')}</Form.Label>
            <Form.Control
              required
              placeholder={t('db_host.placeholder')}
              value={data.db_host.value}
              isInvalid={data.db_host.isInvalid}
              onChange={(e) => {
                changeCallback({
                  db_host: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                });
              }}
            />
            <Form.Control.Feedback type="invalid">
              {data.db_host.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

          <Form.Group controlId="name" className="mb-3">
            <Form.Label>{t('db_name.label')}</Form.Label>
            <Form.Control
              required
              placeholder={t('db_name.placeholder')}
              value={data.db_name.value}
              isInvalid={data.db_name.isInvalid}
              onChange={(e) => {
                changeCallback({
                  db_name: {
                    value: e.target.value,
                    isInvalid: false,
                    errorMsg: '',
                  },
                });
              }}
            />
            <Form.Control.Feedback type="invalid">
              {data.db_name.errorMsg}
            </Form.Control.Feedback>
          </Form.Group>

        </>
      ) : (
        <Form.Group controlId="file" className="mb-3">
          <Form.Label>{t('db_file.label')}</Form.Label>
          <Form.Control
            required
            placeholder={t('db_file.placeholder')}
            value={data.db_file.value}
            isInvalid={data.db_file.isInvalid}
            onChange={(e) => {
              changeCallback({
                db_file: {
                  value: e.target.value,
                  isInvalid: false,
                  errorMsg: '',
                },
              });
            }}
          />
          <Form.Control.Feedback type="invalid">
            {data.db_file.errorMsg}
          </Form.Control.Feedback>
        </Form.Group>
      )}

      <div className="d-flex align-items-center justify-content-between">
        <Progress step={2} />
        <Button type="submit">{t('next')}</Button>
      </div>
    </Form>
  );
};

export default Index;

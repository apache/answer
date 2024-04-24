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

import { FC } from 'react';
import { Card, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import * as Type from '@/common/interface';

interface IProps {
  list: Type.ReviewTypeItem[] | undefined;
  checked: string;
  callback: (type: string) => void;
}

const Index: FC<IProps> = ({ list, checked, callback }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'page_review' });
  return (
    <Card>
      <Card.Header>{t('filter', { keyPrefix: 'btns' })}</Card.Header>
      <Card.Body>
        <Form.Group>
          <Form.Label>{t('filter_label')}</Form.Label>
          {list?.map((item) => {
            return (
              <Form.Check
                key={item.name}
                type="radio"
                id={item.name}
                disabled={item.todo_amount <= 0}
                label={`${item.label} (${item.todo_amount})`}
                checked={checked === item.name}
                onChange={() => callback(item.name)}
              />
            );
          })}
        </Form.Group>
      </Card.Body>
    </Card>
  );
};

export default Index;

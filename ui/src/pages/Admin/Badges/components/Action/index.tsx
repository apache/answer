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

import { Dropdown } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Icon } from '@/components';

interface Props {
  onSelect: (eventKey: string | null) => void;
}
const BadgeOperation = ({ onSelect }: Props) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.badges' });

  return (
    <td className="text-end">
      <Dropdown onSelect={onSelect}>
        <Dropdown.Toggle variant="link" className="no-toggle p-0">
          <Icon name="three-dots-vertical" title={t('action')} />
        </Dropdown.Toggle>
        <Dropdown.Menu align="end">
          <Dropdown.Item eventKey="active">{t('active')}</Dropdown.Item>
          <Dropdown.Item eventKey="inactive">{t('deactivate')}</Dropdown.Item>
          <Dropdown.Divider />
          <Dropdown.Item>{t('show_logs')}</Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown>
    </td>
  );
};

export default BadgeOperation;

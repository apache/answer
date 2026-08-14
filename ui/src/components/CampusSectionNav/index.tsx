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
import { Accordion, Form, Placeholder } from 'react-bootstrap';
import { Link, useLocation, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useForumSections } from '@/services';
import Icon from '@/components/Icon';

import './index.scss';

interface Props {
  mobile?: boolean;
}

const SECTION_ICONS: Record<string, string> = {
  'career-future': 'briefcase-fill',
  'hainanu-campus': 'building-fill',
  'technology-life': 'laptop-fill',
  'life-information': 'basket-fill',
  'site-management': 'megaphone-fill',
};

const CampusSectionNav: FC<Props> = ({ mobile = false }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'campus_forum' });
  const { data, isLoading } = useForumSections();
  const { pathname } = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const current = searchParams.get('section') || '';

  const sectionHref = (slug = '') => {
    const params = new URLSearchParams(searchParams);
    params.delete('page');
    if (slug) params.set('section', slug);
    else params.delete('section');
    return `/questions?${params.toString()}`;
  };

  if (mobile) {
    return (
      <Form.Select
        className="mb-3 d-lg-none"
        aria-label={t('sections')}
        value={current}
        onChange={(event) => {
          const params = new URLSearchParams(searchParams);
          params.delete('page');
          if (event.target.value) params.set('section', event.target.value);
          else params.delete('section');
          setSearchParams(params);
        }}>
        <option value="">{t('all_sections')}</option>
        {data.map((parent) => (
          <optgroup key={parent.id} label={parent.name}>
            {parent.children.map((child) => (
              <option key={child.id} value={child.slug}>
                {child.name}
              </option>
            ))}
          </optgroup>
        ))}
      </Form.Select>
    );
  }

  return (
    <div className="campus-section-nav mt-3">
      <div className="campus-section-title px-3 pb-2 small fw-bold text-secondary">
        {t('sections')}
      </div>
      {isLoading ? (
        <Placeholder as="div" animation="glow" className="px-3 pb-3">
          <Placeholder xs={8} />
        </Placeholder>
      ) : (
        <Accordion
          alwaysOpen
          flush
          defaultActiveKey={data.map((item) => String(item.id))}>
          {data.map((parent) => (
            <Accordion.Item
              key={parent.id}
              eventKey={String(parent.id)}
              className="campus-section-group">
              <Accordion.Header>
                <Icon
                  name={SECTION_ICONS[parent.slug] || 'folder-fill'}
                  className="me-2"
                />
                <span>{parent.name}</span>
              </Accordion.Header>
              <Accordion.Body>
                {parent.children.map((child) => (
                  <Link
                    key={child.id}
                    to={sectionHref(child.slug)}
                    aria-current={
                      pathname === '/questions' && current === child.slug
                        ? 'page'
                        : undefined
                    }
                    className={`campus-section-link campus-section-child d-block ${pathname === '/questions' && current === child.slug ? 'active' : ''}`}>
                    {child.name}
                    {child.admin_only ? (
                      <span className="campus-section-note ms-1">
                        · {t('admin_only')}
                      </span>
                    ) : null}
                  </Link>
                ))}
              </Accordion.Body>
            </Accordion.Item>
          ))}
        </Accordion>
      )}
    </div>
  );
};

export default CampusSectionNav;

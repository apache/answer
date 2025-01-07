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
import { Table, Dropdown, Stack } from 'react-bootstrap';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useTranslation, Trans } from 'react-i18next';

import classNames from 'classnames';

import { Empty, QueryGroup, Icon } from '@/components';
import * as Type from '@/common/interface';
import { useQueryPlugins, updatePluginStatus } from '@/services';
import PluginKit from '@/utils/pluginKit';

const InstalledPluginsFilterKeys: Type.InstalledPluginsFilterBy[] = [
  'all',
  'active',
  'inactive',
];

const bgMap = {
  active: 'text-bg-success',
  inactive: 'text-bg-secondary',
};

const Users: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.installed_plugins',
  });
  const navigate = useNavigate();
  const [urlSearchParams] = useSearchParams();
  const curFilter =
    urlSearchParams.get('filter') || InstalledPluginsFilterKeys[0];
  const {
    data,
    isLoading,
    mutate: updatePlugins,
  } = useQueryPlugins({
    status: curFilter === 'all' ? undefined : curFilter,
  });
  const emitPluginChange = (type) => {
    window.postMessage({
      msgType: type,
    });
  };
  const handleStatus = (plugin) => {
    updatePluginStatus({
      enabled: !plugin.enabled,
      plugin_slug_name: plugin.slug_name,
    }).then(() => {
      updatePlugins();
      PluginKit.refresh();
      if (plugin.have_config) {
        emitPluginChange('refreshConfigurablePlugins');
      }
    });
  };
  const handleSettings = (plugin) => {
    const url = `/admin/${plugin.slug_name}`;
    navigate(url);
  };

  return (
    <>
      <h3>{t('title')}</h3>
      <div className="mb-4">
        <Trans i18nKey="admin.installed_plugins.plugin_link">
          Plugins extend and expand the functionality. You may find plugins in
          the
          <a
            href="https://github.com/apache/answer-plugins"
            target="_blank"
            rel="noreferrer">
            Plugin Repository
          </a>
          .
        </Trans>
      </div>
      <div className="d-flex justify-content-between align-items-center mb-3">
        <Stack direction="horizontal" gap={3}>
          <QueryGroup
            data={InstalledPluginsFilterKeys}
            currentSort={curFilter}
            sortKey="filter"
            i18nKeyPrefix="admin.installed_plugins.filter"
          />
        </Stack>
      </div>
      <Table responsive="md">
        <thead>
          <tr>
            <th className="min-w-15">{t('name')}</th>
            <th style={{ width: '17%' }}>{t('version')}</th>
            <th style={{ width: '11%' }}>{t('status')}</th>
            {curFilter !== 'deleted' ? (
              <th style={{ width: '11%' }} className="text-end">
                {t('action')}
              </th>
            ) : null}
          </tr>
        </thead>
        <tbody className="align-middle">
          {data?.map((plugin) => {
            return (
              <tr key={plugin.slug_name}>
                <td>
                  <div>
                    {plugin.link ? (
                      <a href={plugin.link} target="_blank" rel="noreferrer">
                        {plugin.name}
                      </a>
                    ) : (
                      plugin.name
                    )}
                  </div>
                  <div className="small">{plugin.description}</div>
                </td>
                <td className="text-break">{plugin.version}</td>
                <td>
                  <span
                    className={classNames(
                      'badge',
                      bgMap[plugin.enabled ? 'active' : 'inactive'],
                    )}>
                    {t(`filter.${plugin.enabled ? 'active' : 'inactive'}`)}
                  </span>
                </td>
                {curFilter !== 'deleted' ? (
                  <td className="text-end">
                    <Dropdown>
                      <Dropdown.Toggle variant="link" className="no-toggle">
                        <Icon name="three-dots-vertical" title={t('action')} />
                      </Dropdown.Toggle>
                      <Dropdown.Menu>
                        {plugin.enabled ? (
                          <Dropdown.Item onClick={() => handleStatus(plugin)}>
                            {t('deactivate')}
                          </Dropdown.Item>
                        ) : (
                          <Dropdown.Item onClick={() => handleStatus(plugin)}>
                            {t('activate')}
                          </Dropdown.Item>
                        )}
                        {plugin.enabled && plugin.have_config && (
                          <Dropdown.Item onClick={() => handleSettings(plugin)}>
                            {t('settings')}
                          </Dropdown.Item>
                        )}
                      </Dropdown.Menu>
                    </Dropdown>
                  </td>
                ) : null}
              </tr>
            );
          })}
        </tbody>
      </Table>
      {Number(data?.length) <= 0 && !isLoading && <Empty />}
    </>
  );
};

export default Users;

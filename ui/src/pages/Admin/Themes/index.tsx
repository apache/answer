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

import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import type * as Type from '@/common/interface';
import { getThemeSetting, putThemeSetting } from '@/services';
import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import { useToast } from '@/hooks';
import { handleFormError, scrollToElementTop } from '@/utils';
import { themeSettingStore } from '@/stores';
import { setupAppTheme } from '@/utils/localize';
import { DEFAULT_THEME_COLOR } from '@/common/constants';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.themes',
  });
  const Toast = useToast();
  const [themeSetting, setThemeSetting] = useState<Type.AdminSettingsTheme>();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      themes: {
        type: 'string',
        title: t('themes.label'),
        description: t('themes.text'),
        enum: themeSetting?.theme_options?.map((_) => _.value),
        enumNames: themeSetting?.theme_options?.map((_) => _.label),
        default: themeSetting?.theme_options?.[0]?.value,
      },
      color_scheme: {
        type: 'string',
        title: t('color_scheme.label'),
        enum: ['system', 'light', 'dark'],
        enumNames: [
          t('system_setting', { keyPrefix: 'btns' }),
          t('light', { keyPrefix: 'btns' }),
          t('dark', { keyPrefix: 'btns' }),
        ],
        default: themeSetting?.color_scheme,
      },
      navbar_style: {
        type: 'string',
        title: t('navbar_style.label'),
        default: DEFAULT_THEME_COLOR,
      },
      primary_color: {
        type: 'string',
        title: t('primary_color.label'),
        description: t('primary_color.text'),
        default: DEFAULT_THEME_COLOR,
      },
    },
  };
  const uiSchema: UISchema = {
    themes: {
      'ui:widget': 'select',
    },
    color_scheme: {
      'ui:widget': 'select',
    },
    navbar_style: {
      'ui:widget': 'input_group',
      'ui:options': {
        inputType: 'color',
        suffixBtnOptions: {
          text: '',
          variant: 'outline-secondary',
          iconName: 'arrow-counterclockwise',
          actionType: 'click',
          title: t('reset', { keyPrefix: 'btns' }),
          // eslint-disable-next-line @typescript-eslint/no-use-before-define
          clickCallback: () => resetNavbarStyle(),
        },
      },
    },
    primary_color: {
      'ui:widget': 'input_group',
      'ui:options': {
        inputType: 'color',
        suffixBtnOptions: {
          text: '',
          variant: 'outline-secondary',
          iconName: 'arrow-counterclockwise',
          actionType: 'click',
          title: t('reset', { keyPrefix: 'btns' }),
          // eslint-disable-next-line @typescript-eslint/no-use-before-define
          clickCallback: () => resetPrimaryScheme(),
        },
      },
    },
  };

  const [formData, setFormData] = useState(initFormData(schema));
  const { update: updateThemeSetting } = themeSettingStore((_) => _);

  const resetNavbarStyle = () => {
    const formMeta = { ...formData };
    formMeta.navbar_style.value = DEFAULT_THEME_COLOR;
    setFormData({ ...formMeta });
  };

  const resetPrimaryScheme = () => {
    const formMeta = { ...formData };
    formMeta.primary_color.value = DEFAULT_THEME_COLOR;
    setFormData({ ...formMeta });
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    const themeName = formData.themes.value;
    const reqParams: Type.AdminSettingsTheme = {
      theme: themeName,
      color_scheme: formData.color_scheme.value,
      theme_config: {
        [themeName]: {
          navbar_style: formData.navbar_style.value,
          primary_color: formData.primary_color.value,
        },
      },
    };

    putThemeSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        updateThemeSetting(reqParams);
        setupAppTheme();
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
          const ele = document.getElementById(err.list[0].error_field);
          scrollToElementTop(ele);
        }
      });
  };

  useEffect(() => {
    getThemeSetting().then((setting) => {
      if (setting) {
        setThemeSetting(setting);
        const themeName = setting.theme;
        const themeConfig = setting.theme_config[themeName];
        const formMeta = { ...formData };
        formMeta.themes.value = themeName;
        formMeta.navbar_style.value = themeConfig?.navbar_style.startsWith('#')
          ? themeConfig?.navbar_style
          : DEFAULT_THEME_COLOR;
        formMeta.primary_color.value = themeConfig?.primary_color;
        formData.color_scheme.value = setting?.color_scheme || 'system';
        setFormData({ ...formMeta });
      }
    });
  }, []);

  const handleOnChange = (cd) => {
    setFormData(cd);
    const themeConfig = themeSetting?.theme_config[cd.themes.value];
    if (themeConfig) {
      themeConfig.navbar_style = cd.navbar_style.value;
      themeConfig.primary_color = cd.primary_color.value;
      setThemeSetting({
        ...themeSetting,
        theme: themeSetting?.theme,
        theme_config: themeSetting?.theme_config,
      });
    }
  };

  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <SchemaForm
        schema={schema}
        formData={formData}
        onSubmit={onSubmit}
        uiSchema={uiSchema}
        onChange={handleOnChange}
      />
    </>
  );
};

export default Index;

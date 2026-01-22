import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import type * as Type from '@/common/interface';
import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import { siteSecurityStore } from '@/stores';
import {
  getSecuritySetting,
  putSecuritySetting,
} from '@/services/admin/settings';
import { handleFormError, scrollToElementTop } from '@/utils';
import { useToast } from '@/hooks';

const Security = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.security',
  });
  const Toast = useToast();
  const externalContent = [
    {
      value: 'always_display',
      label: t('external_content_display.always_display', {
        keyPrefix: 'admin.legal',
      }),
    },
    {
      value: 'ask_before_display',
      label: t('external_content_display.ask_before_display', {
        keyPrefix: 'admin.legal',
      }),
    },
  ];

  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      login_required: {
        type: 'boolean',
        title: t('private.title', { keyPrefix: 'admin.login' }),
        description: t('private.text', { keyPrefix: 'admin.login' }),
        default: false,
      },
      external_content_display: {
        type: 'string',
        title: t('external_content_display.label', {
          keyPrefix: 'admin.legal',
        }),
        description: t('external_content_display.text', {
          keyPrefix: 'admin.legal',
        }),
        enum: externalContent?.map((lang) => lang.value),
        enumNames: externalContent?.map((lang) => lang.label),
        default: 0,
      },
      check_update: {
        type: 'boolean',
        title: t('check_update.label', { keyPrefix: 'admin.general' }),
        default: true,
      },
    },
  };
  const uiSchema: UISchema = {
    login_required: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('private.label', { keyPrefix: 'admin.login' }),
      },
    },
    external_content_display: {
      'ui:widget': 'select',
      'ui:options': {
        label: t('external_content_display.label', {
          keyPrefix: 'admin.legal',
        }),
      },
    },
    check_update: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('check_update.label', { keyPrefix: 'admin.general' }),
      },
    },
  };
  const [formData, setFormData] = useState(initFormData(schema));

  const handleValueChange = (data: Type.FormDataType) => {
    setFormData(data);
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    const reqParams = {
      login_required: formData.login_required.value,
      external_content_display: formData.external_content_display.value,
      check_update: formData.check_update.value,
    };
    putSecuritySetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        siteSecurityStore.getState().update(reqParams);
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
    getSecuritySetting().then((setting) => {
      if (setting) {
        const formMeta = { ...formData };
        formMeta.login_required.value = setting.login_required;
        formMeta.external_content_display.value =
          setting.external_content_display;
        formMeta.check_update.value = setting.check_update;
        setFormData(formMeta);
      }
    });
  }, []);

  return (
    <>
      <h3 className="mb-4">{t('security', { keyPrefix: 'nav_menus' })}</h3>
      <div className="max-w-748">
        <SchemaForm
          schema={schema}
          uiSchema={uiSchema}
          formData={formData}
          onChange={handleValueChange}
          onSubmit={onSubmit}
        />
      </div>
    </>
  );
};

export default Security;

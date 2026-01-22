import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import {
  SchemaForm,
  JSONSchema,
  UISchema,
  initFormData,
  TabNav,
} from '@/components';
import { ADMIN_TAGS_NAV_MENUS } from '@/common/constants';
import * as Type from '@/common/interface';
import { handleFormError, scrollToElementTop } from '@/utils';
import { writeSettingStore } from '@/stores';
import { getAdminTagsSetting, updateAdminTagsSetting } from '@/services/admin';
import { useToast } from '@/hooks';

const QaSettings = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.write',
  });
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      reserved_tags: {
        type: 'string',
        title: t('reserved_tags.label'),
        description: t('reserved_tags.text'),
      },
      recommend_tags: {
        type: 'string',
        title: t('recommend_tags.label'),
        description: t('recommend_tags.text'),
      },
      required_tag: {
        type: 'boolean',
        title: t('required_tag.title'),
        description: t('required_tag.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    reserved_tags: {
      'ui:widget': 'tag_selector',
      'ui:options': {
        label: t('reserved_tags.label'),
      },
    },
    recommend_tags: {
      'ui:widget': 'tag_selector',
      'ui:options': {
        label: t('recommend_tags.label'),
      },
    },
    required_tag: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('required_tag.label'),
      },
    },
  };
  const [formData, setFormData] = useState<Type.FormDataType>(
    initFormData(schema),
  );

  const handleValueChange = (data: Type.FormDataType) => {
    setFormData(data);
  };

  const checkValidated = (): boolean => {
    let bol = true;
    const { recommend_tags, reserved_tags } = formData;
    // 找出 recommend_tags 和 reserved_tags 中是否有重复的标签
    // 通过标签中的 slug_name 来去重
    const repeatTag = recommend_tags.value.filter((tag) =>
      reserved_tags.value.some((rTag) => rTag?.slug_name === tag?.slug_name),
    );
    if (repeatTag.length > 0) {
      handleValueChange({
        ...formData,
        recommend_tags: {
          ...recommend_tags,
          errorMsg: t('recommend_tags.msg.contain_reserved'),
          isInvalid: true,
        },
      });
      bol = false;
      const ele = document.getElementById('recommend_tags');
      scrollToElementTop(ele);
    } else {
      handleValueChange({
        ...formData,
        recommend_tags: {
          ...recommend_tags,
          errorMsg: '',
          isInvalid: false,
        },
      });
    }
    return bol;
  };

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    if (!checkValidated()) {
      return;
    }
    const reqParams: Type.AdminTagsSetting = {
      recommend_tags: formData.recommend_tags.value,
      reserved_tags: formData.reserved_tags.value,
      required_tag: formData.required_tag.value,
    };
    updateAdminTagsSetting(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        writeSettingStore.getState().update({ ...reqParams });
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
    getAdminTagsSetting().then((res) => {
      if (res) {
        const formMeta = { ...formData };
        if (Array.isArray(res.recommend_tags)) {
          formData.recommend_tags.value = res.recommend_tags;
        } else {
          formData.recommend_tags.value = [];
        }
        if (Array.isArray(res.reserved_tags)) {
          formData.reserved_tags.value = res.reserved_tags;
        } else {
          formData.reserved_tags.value = [];
        }
        formMeta.required_tag.value = res.required_tag;
        setFormData(formMeta);
      }
    });
  }, []);

  return (
    <>
      <h3 className="mb-4">{t('tags', { keyPrefix: 'nav_menus' })}</h3>
      <TabNav menus={ADMIN_TAGS_NAV_MENUS} />
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

export default QaSettings;

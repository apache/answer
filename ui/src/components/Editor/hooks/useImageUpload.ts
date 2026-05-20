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

import { useTranslation } from 'react-i18next';

import { Modal as AnswerModal } from '@/components';
import { uploadImage } from '@/services';
import { writeSettingStore } from '@/stores';

export const useImageUpload = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const {
    max_image_size = 4,
    max_attachment_size = 8,
    authorized_image_extensions = [],
    authorized_attachment_extensions = [],
  } = writeSettingStore((state) => state.write);

  const verifyImageSize = (files: FileList | File[]): boolean => {
    const fileArray = Array.isArray(files) ? files : Array.from(files);

    if (fileArray.length === 0) {
      return false;
    }

    const canUploadAttachment = authorized_attachment_extensions.length > 0;
    const allowedAllType = [
      ...authorized_image_extensions,
      ...authorized_attachment_extensions,
    ];

    const unSupportFiles = fileArray.filter((file) => {
      const fileName = file.name.toLowerCase();
      return canUploadAttachment
        ? !allowedAllType.find((v) => fileName.endsWith(v))
        : file.type.indexOf('image') === -1;
    });

    if (unSupportFiles.length > 0) {
      AnswerModal.confirm({
        content: canUploadAttachment
          ? t('file.not_supported', { file_type: allowedAllType.join(', ') })
          : t('image.form_image.fields.file.msg.only_image'),
        showCancel: false,
      });
      return false;
    }

    const otherFiles = fileArray.filter((file) => {
      return file.type.indexOf('image') === -1;
    });

    if (canUploadAttachment && otherFiles.length > 0) {
      const attachmentOverSizeFiles = otherFiles.filter(
        (file) => file.size / 1024 / 1024 > max_attachment_size,
      );
      if (attachmentOverSizeFiles.length > 0) {
        AnswerModal.confirm({
          content: t('file.max_size', { size: max_attachment_size }),
          showCancel: false,
        });
        return false;
      }
    }

    const imageFiles = fileArray.filter(
      (file) => file.type.indexOf('image') > -1,
    );
    const oversizedImages = imageFiles.filter(
      (file) => file.size / 1024 / 1024 > max_image_size,
    );
    if (oversizedImages.length > 0) {
      AnswerModal.confirm({
        content: t('image.form_image.fields.file.msg.max_size', {
          size: max_image_size,
        }),
        showCancel: false,
      });
      return false;
    }

    return true;
  };

  const uploadFiles = (
    files: FileList | File[],
  ): Promise<{ url: string; name: string; type: string }[]> => {
    const fileArray = Array.isArray(files) ? files : Array.from(files);
    const promises = fileArray.map(async (file) => {
      const type = file.type.indexOf('image') > -1 ? 'post' : 'post_attachment';
      const url = await uploadImage({ file, type });

      return {
        name: file.name,
        url,
        type,
      };
    });

    return Promise.all(promises);
  };

  const uploadSingleFile = async (file: File): Promise<string> => {
    const type = file.type.indexOf('image') > -1 ? 'post' : 'post_attachment';
    return uploadImage({ file, type });
  };

  return {
    verifyImageSize,
    uploadFiles,
    uploadSingleFile,
  };
};

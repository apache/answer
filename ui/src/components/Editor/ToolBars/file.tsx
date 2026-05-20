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

import { memo, useRef, useContext } from 'react';
import { useTranslation } from 'react-i18next';

import { Modal as AnswerModal } from '@/components';
import ToolItem from '../toolItem';
import { EditorContext } from '../EditorContext';
import { uploadImage } from '@/services';
import { writeSettingStore } from '@/stores';

const File = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const { max_attachment_size = 8, authorized_attachment_extensions = [] } =
    writeSettingStore((state) => state.write);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const editor = useContext(EditorContext);

  const item = {
    label: 'paperclip',
    tip: `${t('file.text')}`,
  };

  const addLink = () => {
    fileInputRef.current?.click?.();
  };

  const verifyFileSize = (files: FileList) => {
    if (files.length === 0) {
      return false;
    }
    const unSupportFiles = Array.from(files).filter((file) => {
      const fileName = file.name.toLowerCase();
      return !authorized_attachment_extensions.find((v) =>
        fileName.endsWith(v),
      );
    });

    if (unSupportFiles.length > 0) {
      AnswerModal.confirm({
        content: t('file.not_supported', {
          file_type: authorized_attachment_extensions.join(', '),
        }),
        showCancel: false,
      });
      return false;
    }

    const attachmentOverSizeFiles = Array.from(files).filter(
      (file) => file.size / 1024 / 1024 > max_attachment_size,
    );
    if (attachmentOverSizeFiles.length > 0) {
      AnswerModal.confirm({
        content: t('file.max_size', { size: max_attachment_size }),
        showCancel: false,
      });
      return false;
    }

    return true;
  };

  const onUpload = async (e) => {
    if (!editor) {
      return;
    }
    const files = e.target?.files || [];
    const bool = verifyFileSize(files);

    if (!bool) {
      return;
    }
    const fileName = files[0].name;
    const loadingText = `![${t('image.uploading')} ${fileName}...]()`;
    const startPos = editor.getCursor();

    const endPos = { ...startPos, ch: startPos.ch + loadingText.length };
    editor.replaceSelection(loadingText);
    editor.setReadOnly(true);

    uploadImage({ file: e.target.files[0], type: 'post_attachment' })
      .then((url) => {
        const text = `[${fileName}](${url})`;
        editor.replaceRange(text, startPos, endPos);
      })
      .catch(() => {
        editor.replaceRange('', startPos, endPos);
      })
      .finally(() => {
        editor.setReadOnly(false);
        editor.focus();
      });
  };

  if (!authorized_attachment_extensions?.length) {
    return null;
  }

  return (
    <ToolItem {...item} onClick={addLink}>
      <input
        type="file"
        className="d-none"
        accept={`.${authorized_attachment_extensions
          .join(',.')
          .toLocaleLowerCase()}`}
        ref={fileInputRef}
        onChange={onUpload}
      />
    </ToolItem>
  );
};

export default memo(File);

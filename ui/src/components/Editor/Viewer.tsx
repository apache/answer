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

import {
  forwardRef,
  useEffect,
  useRef,
  useState,
  memo,
  useImperativeHandle,
} from 'react';
import { useTranslation } from 'react-i18next';

import { markdownToHtml } from '@/services';
import ImgViewer from '@/components/ImgViewer';

import { htmlRender } from './utils';

let scrollTop = 0;
let renderTimer;

const Index = ({ value }, ref) => {
  const [html, setHtml] = useState('');
  const previewRef = useRef<HTMLDivElement>(null);
  const { t } = useTranslation('translation', { keyPrefix: 'messages' });

  const renderMarkdown = (markdown) => {
    clearTimeout(renderTimer);
    const timeout = renderTimer ? 1000 : 0;
    renderTimer = setTimeout(() => {
      markdownToHtml(markdown).then((resp) => {
        scrollTop = previewRef.current?.scrollTop || 0;
        setHtml(resp);
      });
    }, timeout);
  };
  useEffect(() => {
    renderMarkdown(value);
  }, [value]);

  useEffect(() => {
    if (!html) {
      return;
    }

    previewRef.current?.scrollTo(0, scrollTop);

    htmlRender(previewRef.current, {
      copySuccessText: t('copied', { keyPrefix: 'messages' }),
      copyText: t('copy', { keyPrefix: 'messages' }),
    });
  }, [html]);

  useImperativeHandle(ref, () => {
    return {
      getHtml: () => html,
      element: previewRef.current,
    };
  });

  return (
    <ImgViewer>
      <div
        ref={previewRef}
        className="preview-wrap position-relative p-3 rounded text-break text-wrap mt-2 fmt"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </ImgViewer>
  );
};

export default memo(forwardRef(Index));

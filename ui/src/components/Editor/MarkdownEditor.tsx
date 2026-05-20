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

import { useEffect, useRef } from 'react';

import { EditorView } from '@codemirror/view';

import { BaseEditorProps } from './types';
import { useEditor } from './utils';

interface MarkdownEditorProps extends BaseEditorProps {}

const MarkdownEditor: React.FC<MarkdownEditorProps> = ({
  value,
  onChange,
  onFocus,
  onBlur,
  placeholder,
  autoFocus,
  onEditorReady,
}) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const lastSyncedValueRef = useRef<string>(value);
  const isInitializedRef = useRef<boolean>(false);

  const editor = useEditor({
    editorRef,
    onChange,
    onFocus,
    onBlur,
    placeholder,
    autoFocus,
    initialValue: value,
  });

  useEffect(() => {
    if (!editor || isInitializedRef.current) {
      return;
    }

    isInitializedRef.current = true;
    onEditorReady?.(editor);
  }, [editor, onEditorReady]);

  useEffect(() => {
    if (!editor || value === lastSyncedValueRef.current) {
      return;
    }

    const currentValue = editor.getValue();
    if (currentValue !== value) {
      editor.setValue(value || '');
      lastSyncedValueRef.current = value || '';
    }
  }, [editor, value]);

  useEffect(() => {
    lastSyncedValueRef.current = value;
    isInitializedRef.current = false;

    return () => {
      if (editor) {
        const view = editor as unknown as EditorView;
        if (view.destroy) {
          view.destroy();
        }
      }
      isInitializedRef.current = false;
    };
  }, []);

  return (
    <div className="content-wrap">
      <div
        className="md-editor position-relative w-100 h-100"
        ref={editorRef}
      />
    </div>
  );
};

export default MarkdownEditor;

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

import { StateEffect } from '@codemirror/state';
import { EditorView } from '@codemirror/view';

import { Editor } from '../../types';

/**
 * Creates event methods module
 *
 * Provides event listener registration and removal for the editor.
 * Handles various DOM events including focus, blur, drag, drop, and paste.
 *
 * @param editor - CodeMirror editor instance
 * @returns Object containing event methods (on, off)
 */
export function createEventMethods(editor: Editor) {
  return {
    on: (event, callback) => {
      if (event === 'change') {
        const change = EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            callback();
          }
        });

        editor.dispatch({
          effects: StateEffect.appendConfig.of(change),
        });
      }
      if (event === 'focus') {
        editor.contentDOM.addEventListener('focus', callback);
      }
      if (event === 'blur') {
        editor.contentDOM.addEventListener('blur', callback);
      }

      if (event === 'dragenter') {
        editor.contentDOM.addEventListener('dragenter', callback);
      }

      if (event === 'dragover') {
        editor.contentDOM.addEventListener('dragover', callback);
      }

      if (event === 'drop') {
        editor.contentDOM.addEventListener('drop', callback);
      }

      if (event === 'paste') {
        editor.contentDOM.addEventListener('paste', callback);
      }
    },

    off: (event, callback) => {
      if (event === 'focus') {
        editor.contentDOM.removeEventListener('focus', callback);
      }

      if (event === 'blur') {
        editor.contentDOM.removeEventListener('blur', callback);
      }

      if (event === 'dragenter') {
        editor.contentDOM.removeEventListener('dragenter', callback);
      }

      if (event === 'dragover') {
        editor.contentDOM.removeEventListener('dragover', callback);
      }

      if (event === 'drop') {
        editor.contentDOM.removeEventListener('drop', callback);
      }

      if (event === 'paste') {
        editor.contentDOM.removeEventListener('paste', callback);
      }
    },
  };
}

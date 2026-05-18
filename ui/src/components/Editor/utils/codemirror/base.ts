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

import { EditorSelection, StateEffect } from '@codemirror/state';
import { keymap, KeyBinding, Command } from '@codemirror/view';

import { Editor, Position } from '../../types';

/**
 * Creates base methods module
 *
 * Provides core base methods for the editor, including:
 * - Content getter and setter (getValue, setValue)
 * - Selection operations (getSelection, replaceSelection)
 * - Cursor and selection position (getCursor, setSelection)
 * - Focus and keyboard mapping (focus, addKeyMap)
 *
 * @param editor - CodeMirror editor instance
 * @returns Object containing base methods
 */
export function createBaseMethods(editor: Editor) {
  return {
    focus: () => {
      editor.contentDOM.focus();
    },

    getCursor: () => {
      const range = editor.state.selection.ranges[0];
      const line = editor.state.doc.lineAt(range.from).number;
      const { from, to } = editor.state.doc.line(line);
      return { from, to, ch: range.from - from, line };
    },

    addKeyMap: (keyMap: Record<string, Command>) => {
      const array = Object.entries(keyMap).map(([key, value]) => {
        const keyBinding: KeyBinding = {
          key,
          preventDefault: true,
          run: value,
        };
        return keyBinding;
      });

      editor.dispatch({
        effects: StateEffect.appendConfig.of(keymap.of(array)),
      });
    },

    getSelection: () => {
      return editor.state.sliceDoc(
        editor.state.selection.main.from,
        editor.state.selection.main.to,
      );
    },

    replaceSelection: (value: string) => {
      editor.dispatch({
        changes: [
          {
            from: editor.state.selection.main.from,
            to: editor.state.selection.main.to,
            insert: value,
          },
        ],
        selection: EditorSelection.cursor(
          editor.state.selection.main.from + value.length,
        ),
      });
    },

    setSelection: (anchor: Position, head?: Position) => {
      editor.dispatch({
        selection: EditorSelection.create([
          EditorSelection.range(
            editor.state.doc.line(anchor.line).from + anchor.ch,
            head
              ? editor.state.doc.line(head.line).from + head.ch
              : editor.state.doc.line(anchor.line).from + anchor.ch,
          ),
        ]),
      });
    },

    replaceRange: (value: string, from: Position, to: Position) => {
      const fromOffset = editor.state.doc.line(from.line).from + from.ch;
      const toOffset = editor.state.doc.line(to.line).from + to.ch;
      editor.dispatch({
        changes: { from: fromOffset, to: toOffset, insert: value },
      });
    },

    getValue: () => {
      return editor.state.doc.toString();
    },

    setValue: (value: string) => {
      editor.dispatch({
        changes: { from: 0, to: editor.state.doc.length, insert: value },
      });
    },
  };
}

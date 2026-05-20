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

import { EditorSelection } from '@codemirror/state';

import { Editor, Level } from '../../types';

/**
 * Creates command methods module
 *
 * Provides semantic command methods and low-level text manipulation methods:
 * - Semantic methods: insertBold, insertHeading, insertImage, etc. (for toolbar use)
 * - Low-level methods: wrapText, replaceLines, appendBlock (for internal use)
 * - State query methods: isBold, isHeading, etc.
 *
 * @param editor - CodeMirror editor instance
 * @returns Object containing all command methods
 */
export function createCommandMethods(editor: Editor) {
  // Create methods object that allows self-reference
  const methods = {
    wrapText: (before: string, after = before, defaultText) => {
      const range = editor.state.selection.ranges[0];
      const selectedText = editor.state.sliceDoc(range.from, range.to);
      const text = selectedText || defaultText || '';
      const wrappedText = before + text + after;
      const insertFrom = range.from;
      const insertTo = range.to;

      editor.dispatch({
        changes: [
          {
            from: insertFrom,
            to: insertTo,
            insert: wrappedText,
          },
        ],
        selection: selectedText
          ? EditorSelection.cursor(insertFrom + before.length + text.length)
          : EditorSelection.range(
              insertFrom + before.length,
              insertFrom + before.length + text.length,
            ),
      });
    },

    replaceLines: (replace: Parameters<Array<string>['map']>[0]) => {
      const { doc } = editor.state;
      const lines: string[] = [];
      for (let i = 1; i <= doc.lines; i += 1) {
        lines.push(doc.line(i).text);
      }

      const newLines = lines.map(replace) as string[];
      const newText = newLines.join('\n');
      editor.dispatch({
        changes: {
          from: 0,
          to: editor.state.doc.length,
          insert: newText,
        },
      });
    },

    appendBlock: (content: string) => {
      const { doc } = editor.state;
      const currentText = doc.toString();
      const newText = currentText ? `${currentText}\n\n${content}` : content;
      editor.dispatch({
        changes: {
          from: 0,
          to: editor.state.doc.length,
          insert: newText,
        },
      });
    },

    insertBold: (text?: string) => {
      methods.wrapText('**', '**', text || 'bold text');
    },

    insertItalic: (text?: string) => {
      methods.wrapText('*', '*', text || 'italic text');
    },

    insertCode: (text?: string) => {
      methods.wrapText('`', '`', text || 'code');
    },

    insertStrikethrough: (text?: string) => {
      methods.wrapText('~~', '~~', text || 'strikethrough text');
    },

    insertHeading: (level: Level, text?: string) => {
      const headingText = '#'.repeat(level);
      methods.wrapText(`${headingText} `, '', text || 'heading');
    },

    insertBlockquote: (text?: string) => {
      methods.wrapText('> ', '', text || 'quote');
    },

    insertCodeBlock: (language?: string, code?: string) => {
      const lang = language || '';
      const codeText = code || '';
      const block = `\`\`\`${lang}\n${codeText}\n\`\`\``;
      methods.appendBlock(block);
    },

    insertHorizontalRule: () => {
      methods.appendBlock('---');
    },

    insertOrderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      if (/^\d+\.\s/.test(lineText)) {
        return;
      }
      methods.replaceLines((lineItem) => {
        if (lineItem.trim() === '') {
          return lineItem;
        }
        return `1. ${lineItem}`;
      });
    },

    insertUnorderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      if (/^[-*+]\s/.test(lineText)) {
        return;
      }
      methods.replaceLines((lineItem) => {
        if (lineItem.trim() === '') {
          return lineItem;
        }
        return `- ${lineItem}`;
      });
    },

    toggleOrderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      if (/^\d+\.\s/.test(lineText)) {
        methods.replaceLines((lineItem) => {
          return lineItem.replace(/^\d+\.\s/, '');
        });
      } else {
        methods.insertOrderedList();
      }
    },

    toggleUnorderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      if (/^[-*+]\s/.test(lineText)) {
        methods.replaceLines((lineItem) => {
          return lineItem.replace(/^[-*+]\s/, '');
        });
      } else {
        methods.insertUnorderedList();
      }
    },

    insertLink: (url: string, text?: string) => {
      const linkText = text || url;
      methods.wrapText('[', `](${url})`, linkText);
    },

    insertImage: (url: string, alt?: string) => {
      const altText = alt || '';
      methods.wrapText('![', `](${url})`, altText);
    },

    insertTable: (rows = 3, cols = 3) => {
      const table: string[] = [];
      for (let i = 0; i < rows; i += 1) {
        const row: string[] = [];
        for (let j = 0; j < cols; j += 1) {
          row.push(i === 0 ? 'Header' : 'Cell');
        }
        table.push(`| ${row.join(' | ')} |`);
        if (i === 0) {
          table.push(`| ${'---'.repeat(cols).split('').join(' | ')} |`);
        }
      }
      methods.appendBlock(table.join('\n'));
    },

    indent: () => {
      methods.replaceLines((line) => {
        if (line.trim() === '') {
          return line;
        }
        return `  ${line}`;
      });
    },

    outdent: () => {
      methods.replaceLines((line) => {
        if (line.trim() === '') {
          return line;
        }
        return line.replace(/^ {2}/, '');
      });
    },

    isBold: () => {
      const selection = editor.getSelection();
      return /^\*\*.*\*\*$/.test(selection) || /^__.*__$/.test(selection);
    },

    isItalic: () => {
      const selection = editor.getSelection();
      return /^\*.*\*$/.test(selection) || /^_.*_$/.test(selection);
    },

    isHeading: (level?: number) => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      if (level) {
        return new RegExp(`^#{${level}}\\s`).test(lineText);
      }
      return /^#{1,6}\s/.test(lineText);
    },

    isBlockquote: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      return /^>\s/.test(lineText);
    },

    isCodeBlock: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      return /^```/.test(lineText);
    },

    isOrderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      return /^\d+\.\s/.test(lineText);
    },

    isUnorderedList: () => {
      const cursor = editor.getCursor();
      const line = editor.state.doc.line(cursor.line);
      const lineText = line.text.trim();
      return /^[-*+]\s/.test(lineText);
    },
  };

  return methods;
}

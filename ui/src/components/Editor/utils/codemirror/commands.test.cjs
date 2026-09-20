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

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const ts = require('typescript');

require.extensions['.ts'] = (module, filename) => {
  const source = fs.readFileSync(filename, 'utf8');
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      esModuleInterop: true,
    },
    fileName: filename,
  }).outputText;
  module._compile(output, filename);
};

const { EditorState } = require('@codemirror/state');
const { createCommandMethods } = require('./commands.ts');

function createEditor(doc = '') {
  let state = EditorState.create({ doc });

  return {
    get state() {
      return state;
    },
    dispatch(spec) {
      state = state.update(spec).state;
    },
    getCursor() {
      const range = state.selection.ranges[0];
      return { line: state.doc.lineAt(range.from).number };
    },
  };
}

for (const [name, command, expected] of [
  ['unordered', 'insertUnorderedList', '- '],
  ['ordered', 'insertOrderedList', '1. '],
]) {
  test(`${name} list command inserts a marker in an empty editor`, () => {
    const editor = createEditor();
    const commands = createCommandMethods(editor);

    commands[command]();

    assert.equal(editor.state.doc.toString(), expected);
    assert.equal(editor.state.selection.main.head, expected.length);
  });
}

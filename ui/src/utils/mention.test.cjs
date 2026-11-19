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
const Module = require('node:module');
const test = require('node:test');

let transpileTypeScript;
try {
  const ts = require('typescript');
  transpileTypeScript = (source, filename) =>
    ts.transpileModule(source, {
      compilerOptions: {
        module: ts.ModuleKind.CommonJS,
        target: ts.ScriptTarget.ES2020,
        esModuleInterop: true,
      },
      fileName: filename,
    }).outputText;
} catch {
  transpileTypeScript = (source) => {
    const stripped = Module.stripTypeScriptTypes(source);
    return stripped.replace(
      /export\s*\{\s*([^}]+)\s*\};?/g,
      'module.exports = { $1 };',
    );
  };
}

const previousTypeScriptLoader = require.extensions['.ts'];
require.extensions['.ts'] = (mod, filename) => {
  const source = fs.readFileSync(filename, 'utf8');
  const output = transpileTypeScript(source, filename);
  mod._compile(output, filename);
};

let matchedUsers;
let parseUserInfo;
let parseEditMentionUser;
try {
  ({ matchedUsers, parseUserInfo, parseEditMentionUser } = require('./mention.ts'));
} finally {
  if (previousTypeScriptLoader) {
    require.extensions['.ts'] = previousTypeScriptLoader;
  } else {
    delete require.extensions['.ts'];
  }
}

test('matches usernames containing dots, hyphens, and underscores', () => {
  const markdown = 'cc @jane-doe and @john.doe and @dev_user-1.2';

  assert.deepEqual(matchedUsers(markdown), [
    { userName: 'jane-doe' },
    { userName: 'john.doe' },
    { userName: 'dev_user-1.2' },
  ]);
  assert.equal(
    parseUserInfo(markdown),
    'cc [@jane-doe](/users/jane-doe) and [@john.doe](/users/john.doe) and [@dev_user-1.2](/users/dev_user-1.2)',
  );
});

test('treats trailing sentence period as punctuation unless username with trailing dot is known', () => {
  const sentence = 'Thanks @john.';

  assert.deepEqual(matchedUsers(sentence), [{ userName: 'john' }]);
  assert.equal(parseUserInfo(sentence), 'Thanks [@john](/users/john).');

  const knownUsers = [{ userName: 'john.' }];
  assert.deepEqual(matchedUsers(sentence, knownUsers), [{ userName: 'john.' }]);
  assert.equal(
    parseUserInfo(sentence, knownUsers),
    'Thanks [@john.](/users/john.)',
  );

  const sentenceWithKnownTrailingDot = 'Thanks @john..';
  assert.deepEqual(matchedUsers(sentenceWithKnownTrailingDot, knownUsers), [
    { userName: 'john.' },
  ]);
  assert.equal(
    parseUserInfo(sentenceWithKnownTrailingDot, knownUsers),
    'Thanks [@john.](/users/john.).',
  );
});

test('excludes trailing punctuation from both notification targets and rendered links', () => {
  const markdown =
    'Ping @john.doe., @jane-doe, @alice! (@bob?) @carol...';

  assert.deepEqual(matchedUsers(markdown), [
    { userName: 'john.doe' },
    { userName: 'jane-doe' },
    { userName: 'alice' },
    { userName: 'bob' },
    { userName: 'carol' },
  ]);
  assert.equal(
    parseUserInfo(markdown),
    'Ping [@john.doe](/users/john.doe)., [@jane-doe](/users/jane-doe), [@alice](/users/alice)! ([@bob](/users/bob)?) [@carol](/users/carol)...',
  );
});

test('stops before invalid username characters and ignores non-mentions', () => {
  assert.deepEqual(matchedUsers('@john\\doe'), [{ userName: 'john' }]);
  assert.equal(parseUserInfo('@john\\doe'), '[@john](/users/john)\\doe');

  assert.deepEqual(matchedUsers('@john|doe'), [{ userName: 'john' }]);
  assert.equal(parseUserInfo('@john|doe'), '[@john](/users/john)|doe');

  assert.deepEqual(matchedUsers('@\\doe @! @. @... user@example.com'), []);
  assert.equal(
    parseUserInfo('@\\doe @! @. @... user@example.com'),
    '@\\doe @! @. @... user@example.com',
  );
});

test('handles multiple mentions and already-rendered mention links consistently', () => {
  const knownUsers = [{ userName: 'ops.lead.' }];
  const markdown =
    '@jane-doe, please check with @ops.lead. and [@john.](/users/john.) about @john.doe.';

  assert.deepEqual(matchedUsers(markdown, knownUsers), [
    { userName: 'jane-doe' },
    { userName: 'ops.lead.' },
    { userName: 'john.' },
    { userName: 'john.doe' },
  ]);
  const rendered = parseUserInfo(markdown, knownUsers);
  assert.equal(
    rendered,
    '[@jane-doe](/users/jane-doe), please check with [@ops.lead.](/users/ops.lead.) and [@john.](/users/john.) about [@john.doe](/users/john.doe).',
  );
  assert.equal(
    parseEditMentionUser(rendered),
    '@jane-doe, please check with @ops.lead. and @john. about @john.doe.',
  );
});

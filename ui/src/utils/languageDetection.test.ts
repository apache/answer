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

import { describe, expect, it } from 'vitest';

import {
  doesTextNeedTranslation,
  normalizeLanguageDetectionText,
} from './languageDetection';

describe('doesTextNeedTranslation', () => {
  it('detects a different script in short text', async () => {
    await expect(doesTextNeedTranslation('你好世界', 'en_US')).resolves.toBe(
      true,
    );
  });

  it('does not offer translation for the target language', async () => {
    await expect(
      doesTextNeedTranslation(
        'This is a sufficiently long English question about software testing.',
        'en_US',
      ),
    ).resolves.toBe(false);
  });

  it('detects a confidently different Latin language', async () => {
    await expect(
      doesTextNeedTranslation(
        'Wie kann ich dieses Problem in meiner Anwendung zuverlässig lösen?',
        'en_US',
      ),
    ).resolves.toBe(true);
  });

  it('does not guess between Latin languages when text is too short', async () => {
    await expect(doesTextNeedTranslation('Hello', 'de_DE')).resolves.toBe(
      false,
    );
  });

  it('detects a short input written in a different script', async () => {
    await expect(doesTextNeedTranslation('Hello', 'zh_CN')).resolves.toBe(true);
  });

  it('ignores code, URLs, and Markdown links', () => {
    expect(
      normalizeLanguageDetectionText(
        '```js\nconst greeting = "你好";\n``` https://example.com [docs](https://example.com)',
      ),
    ).toBe('docs');
  });
});

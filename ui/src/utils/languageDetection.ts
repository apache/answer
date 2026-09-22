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

const localeToISO3: Record<string, string> = {
  en: 'eng',
  es: 'spa',
  pt: 'por',
  de: 'deu',
  fr: 'fra',
  ja: 'jpn',
  it: 'ita',
  ru: 'rus',
  zh: 'cmn',
  ko: 'kor',
  vi: 'vie',
  sk: 'slk',
  fa: 'pes',
};

const supportedLanguages = Array.from(new Set(Object.values(localeToISO3)));
const minimumScore = 0.8;
const minimumLead = 0.15;
const minimumLatinSampleLength = 10;

type Script = 'arabic' | 'cyrillic' | 'han' | 'hangul' | 'japanese' | 'latin';

const localeToScript: Record<string, Script> = {
  en: 'latin',
  es: 'latin',
  pt: 'latin',
  de: 'latin',
  fr: 'latin',
  ja: 'japanese',
  it: 'latin',
  ru: 'cyrillic',
  zh: 'han',
  ko: 'hangul',
  vi: 'latin',
  sk: 'latin',
  fa: 'arabic',
};

const detectScript = (
  text: string,
  letterCount: number,
): Script | undefined => {
  const japanese =
    text.match(/[\p{Script=Hiragana}\p{Script=Katakana}]/gu)?.length || 0;
  if (japanese > 0) {
    return 'japanese';
  }

  const scripts: Array<[Script, number]> = [
    ['arabic', text.match(/\p{Script=Arabic}/gu)?.length || 0],
    ['cyrillic', text.match(/\p{Script=Cyrillic}/gu)?.length || 0],
    ['han', text.match(/\p{Script=Han}/gu)?.length || 0],
    ['hangul', text.match(/\p{Script=Hangul}/gu)?.length || 0],
    ['latin', text.match(/\p{Script=Latin}/gu)?.length || 0],
  ];
  const [script, count] = scripts.sort((a, b) => b[1] - a[1])[0];
  return count >= 3 && count / letterCount >= 0.5 ? script : undefined;
};

export const normalizeLanguageDetectionText = (value: string) =>
  value
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`[^`]*`/g, ' ')
    .replace(/https?:\/\/\S+/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/!?(\[([^\]]+)\])\([^)]*\)/g, '$2')
    .replace(/[@#][\w-]+/g, ' ')
    .replace(/[\s*_~>|=[\]{}()-]+/g, ' ')
    .trim();

export const doesTextNeedTranslation = async (
  value: string,
  targetLocale: string,
): Promise<boolean> => {
  const locale = targetLocale.split(/[-_]/)[0];
  const targetLanguage = localeToISO3[locale];
  const targetScript = localeToScript[locale];
  const text = normalizeLanguageDetectionText(value);
  const letters = text.match(/\p{L}/gu)?.length || 0;

  if (!targetLanguage || letters < 3) {
    return false;
  }

  const detectedScript = detectScript(text, letters);
  if (
    detectedScript &&
    targetScript &&
    detectedScript !== targetScript &&
    !(targetScript === 'japanese' && detectedScript === 'han')
  ) {
    return true;
  }
  if (detectedScript === 'latin' && letters < minimumLatinSampleLength) {
    return false;
  }

  const { francAll } = await import('franc-min');
  const [best, second] = francAll(text, {
    only: supportedLanguages,
    minLength: 3,
  });

  if (!best || best[0] === 'und' || best[0] === targetLanguage) {
    return false;
  }

  return best[1] >= minimumScore && best[1] - (second?.[1] ?? 0) >= minimumLead;
};

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

import i18next from 'i18next';

import pattern from '@/common/pattern';
import { USER_AGENT_NAMES } from '@/common/constants';
import type * as Type from '@/common/interface';

const Diff = require('diff');

function thousandthDivision(num) {
  const reg = /\d{1,3}(?=(\d{3})+$)/g;
  return `${num}`.replace(reg, '$&,');
}

function formatCount($num: number): string {
  let res = String($num);
  if (!Number.isFinite($num)) {
    res = '0';
  } else if ($num < 10000) {
    res = thousandthDivision($num);
  } else if ($num < 1000000) {
    res = `${Math.round($num / 100) / 10}k`;
  } else if ($num >= 1000000) {
    res = `${Math.round($num / 100000) / 10}m`;
  }
  return res;
}

function scrollToElementTop(element) {
  if (!element) {
    return;
  }
  const offset = 120;
  const bodyRect = document.body.getBoundingClientRect().top;
  const elementRect = element.getBoundingClientRect().top;
  const elementPosition = elementRect - bodyRect;
  const offsetPosition = elementPosition - offset;

  window.scrollTo({
    top: offsetPosition,
    behavior: 'instant' as ScrollBehavior,
  });
}

function scrollElementIntoView(element) {
  if (!element) {
    return;
  }
  element.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
    inline: 'center',
  });
}

const scrollToDocTop = () => {
  setTimeout(() => {
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'instant' as ScrollBehavior,
    });
  });
};

const bgFadeOut = (el) => {
  if (el && !el.classList.contains('bg-fade-out')) {
    el.classList.add('bg-fade-out');
    setTimeout(() => {
      el.classList.remove('bg-fade-out');
    }, 3200);
  }
};

/**
 * Extract user info from markdown
 * @param markdown string
 * @returns Array<{displayName: string, userName: string}>
 */
function matchedUsers(markdown) {
  const globalReg = /\B@([\w|]+)/g;
  const reg = /\B@([\w\\_\\.]+)/;

  const users = markdown.match(globalReg);
  if (!users) {
    return [];
  }
  return users.map((user) => {
    const matched = user.match(reg);
    return {
      userName: matched[1],
    };
  });
}

/**
 * Identify user information from markdown
 * @param markdown string
 * @returns string
 */
function parseUserInfo(markdown) {
  const globalReg = /\B@([\w\\_\\.\\-]+)/g;
  return markdown.replace(globalReg, '[@$1](/users/$1)');
}

function parseEditMentionUser(markdown) {
  const globalReg = /\[@([^\]]+)\]\([^)]+\)/g;
  return markdown.replace(globalReg, '@$1');
}

function formatUptime(value) {
  const t = i18next.t.bind(i18next);
  const second = parseInt(value, 10);

  if (second > 60 * 60 && second < 60 * 60 * 24) {
    const hour = second / 3600;
    return `${Math.floor(hour)} ${
      hour > 1 ? t('dates.hours') : t('dates.hour')
    }`;
  }
  if (second > 60 * 60 * 24) {
    const day = second / 3600 / 24;
    return `${Math.floor(day)} ${day > 1 ? t('dates.days') : t('dates.day')}`;
  }

  return `< 1 ${t('dates.hour')}`;
}

function escapeRemove(str: string) {
  if (!str || typeof str !== 'string') return str;
  let temp: HTMLDivElement | null = document.createElement('div');
  temp.innerHTML = str;
  const output = temp?.innerText || temp.textContent;
  temp = null;
  return output;
}

function handleFormError(
  error: { list: Type.FieldError[] },
  data: any,
  keymap?: Array<{ from: string; to: string }>,
) {
  if (error.list?.length > 0) {
    error.list.forEach((item) => {
      if (keymap?.length) {
        const key = keymap.find((k) => k.from === item.error_field);
        if (key) {
          item.error_field = key.to;
        }
      }
      const errorFieldObject = data[item.error_field];
      if (errorFieldObject) {
        errorFieldObject.isInvalid = true;
        errorFieldObject.errorMsg = item.error_msg;
      }
    });
  }
  return data;
}

function escapeHtml(str: string) {
  const tagsToReplace = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
    '`': '&#96;',
  };
  return str.replace(/[&<>"'`]/g, (tag) => tagsToReplace[tag] || tag);
}

function formatDiffPart(part: any, className: string): string {
  if (part.value.replace(/\n/g, '').length <= 0) {
    if (part.value.match(/\n/g)?.length > 1) {
      const value = part.value.replace(/\n/, '');
      return `<span class="${className}"> </span><div><span class="${className}">${value.replace(
        /\n/g,
        ' ',
      )}</span></div>`;
    }
    return `<div><span class="${className}">${part.value.replace(
      /\n/g,
      ' ',
    )}</span></div>`;
  }
  return `<span class="${className}">${part.value}</span>`;
}

function diffText(newText: string, oldText?: string): string {
  if (!newText) {
    return '';
  }

  if (typeof oldText !== 'string') {
    return escapeHtml(newText);
  }
  let result = [];
  const diff = Diff.diffChars(escapeHtml(oldText), escapeHtml(newText));
  result = diff.map((part) => {
    if (part.added) {
      return formatDiffPart(part, 'review-text-add');
    }
    if (part.removed) {
      return formatDiffPart(part, 'review-text-delete text-decoration-none');
    }

    return part.value;
  });

  return result.join('');
}

function base64ToSvg(base64: string, svgClassName?: string) {
  try {
    // base64 to svg xml
    const svgxml = atob(base64);

    // svg add class
    const parser = new DOMParser();
    const doc = parser.parseFromString(svgxml, 'image/svg+xml');
    const parseError = doc.querySelector('parsererror');
    const svg = doc.querySelector('svg');
    let str = '';
    if (svg && !parseError) {
      if (svgClassName) {
        svg.setAttribute('class', svgClassName);
      }
      // svg.classList.add('me-2');

      // transform svg to string
      const serializer = new XMLSerializer();
      str = serializer.serializeToString(doc);
    }
    return str;
  } catch (error) {
    return '';
  }
}

// Determine whether the user is in WeChat or Enterprise WeChat or DingTalk, and return the corresponding type

function getUaType() {
  const ua = navigator.userAgent.toLowerCase();
  if (pattern.uaWeCom.test(ua)) {
    return USER_AGENT_NAMES.WeCom;
  }
  if (pattern.uaWeChat.test(ua)) {
    return USER_AGENT_NAMES.WeChat;
  }
  if (pattern.uaDingTalk.test(ua)) {
    return USER_AGENT_NAMES.DingTalk;
  }
  return null;
}

function changeTheme(mode: 'default' | 'light' | 'dark' | 'system') {
  const htmlTag = document.querySelector('html') as HTMLHtmlElement;
  if (mode === 'system') {
    const systemThemeQuery = window.matchMedia('(prefers-color-scheme: dark)');

    if (systemThemeQuery.matches) {
      htmlTag.setAttribute('data-bs-theme', 'dark');
    } else {
      htmlTag.setAttribute('data-bs-theme', 'light');
    }
  } else {
    htmlTag.setAttribute('data-bs-theme', mode);
  }
}

function isDarkTheme() {
  const htmlTag = document.querySelector('html') as HTMLHtmlElement;
  return htmlTag.getAttribute('data-bs-theme') === 'dark';
}

function pageTitleType() {
  const { pathname } = window.location;
  if (pathname.endsWith('/articles')) {
    return 'articles';
  }
  if (pathname.endsWith('/questions')) {
    return 'questions';
  }
  return 'posts';
}

export {
  thousandthDivision,
  formatCount,
  scrollElementIntoView,
  scrollToElementTop,
  scrollToDocTop,
  bgFadeOut,
  matchedUsers,
  parseUserInfo,
  parseEditMentionUser,
  formatUptime,
  escapeRemove,
  handleFormError,
  diffText,
  base64ToSvg,
  getUaType,
  changeTheme,
  isDarkTheme,
  pageTitleType,
};

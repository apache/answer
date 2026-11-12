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

type KnownUser = { userName?: string } | string;

const MENTION_REG = /\B@([\w.-]+)/g;

function toKnownUserSet(knownUsers: KnownUser[] = []): Set<string> {
  const set = new Set<string>();
  knownUsers.forEach((user) => {
    const userName = typeof user === 'string' ? user : user?.userName;
    if (userName) {
      set.add(userName);
    }
  });
  return set;
}

function isRenderedMentionLink(
  source: string,
  offset: number,
  matchLength: number,
): boolean {
  return (
    offset > 0 &&
    source[offset - 1] === '[' &&
    source.slice(offset + matchLength).startsWith('](/users/')
  );
}

function resolveMentionUsername(
  raw: string,
  knownUserSet: Set<string>,
): string {
  let candidate = raw;
  while (candidate.endsWith('.') && !knownUserSet.has(candidate)) {
    candidate = candidate.slice(0, -1);
  }
  return candidate;
}

/**
 * Extract user info from markdown
 * @param markdown string
 * @param knownUsers Array<{ userName?: string } | string>
 * @returns Array<{userName: string}>
 */
function matchedUsers(markdown = '', knownUsers: KnownUser[] = []) {
  if (!markdown) {
    return [];
  }
  const knownUserSet = toKnownUserSet(knownUsers);
  const users: Array<{ userName: string }> = [];

  markdown.replace(
    MENTION_REG,
    (match: string, raw: string, offset: number, source: string) => {
      const userName = isRenderedMentionLink(source, offset, match.length)
        ? raw
        : resolveMentionUsername(raw, knownUserSet);
      if (userName) {
        users.push({ userName });
      }
      return match;
    },
  );

  return users;
}

/**
 * Identify user information from markdown
 * @param markdown string
 * @param knownUsers Array<{ userName?: string } | string>
 * @returns string
 */
function parseUserInfo(markdown = '', knownUsers: KnownUser[] = []) {
  if (!markdown) {
    return '';
  }
  const knownUserSet = toKnownUserSet(knownUsers);

  return markdown.replace(
    MENTION_REG,
    (match: string, raw: string, offset: number, source: string) => {
      if (isRenderedMentionLink(source, offset, match.length)) {
        return match;
      }
      const userName = resolveMentionUsername(raw, knownUserSet);
      if (!userName) {
        return match;
      }
      const trailing = raw.slice(userName.length);
      return `[@${userName}](/users/${userName})${trailing}`;
    },
  );
}

function parseEditMentionUser(markdown) {
  const globalReg = /\[@([^\]]+)\]\([^)]+\)/g;
  return markdown.replace(globalReg, '@$1');
}

export { matchedUsers, parseUserInfo, parseEditMentionUser };

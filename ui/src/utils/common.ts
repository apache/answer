function getQueryString(name: string): string {
  const reg = new RegExp(`(^|&)${name}=([^&]*)(&|$)`);
  const r = window.location.search.substr(1).match(reg);
  if (r != null) return unescape(r[2]);
  return '';
}

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

function scrollTop(element) {
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
  });
}

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
  return markdown.replace(globalReg, '[@$1](/u/$1)');
}

export {
  getQueryString,
  thousandthDivision,
  formatCount,
  scrollTop,
  matchedUsers,
  parseUserInfo,
};

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

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const UI_DIR = path.resolve(__dirname, '..');  // UI directory
const ROOT_DIR = path.resolve(UI_DIR, '..');   // Project root directory
const GIT_DIR = getGitDir(ROOT_DIR);           // Git root directory
const HUSKY_DIR = path.join(GIT_DIR, '.husky');

// Find Git directory
function getGitDir(startDir) {
  let currentDir = startDir;

  while (currentDir !== path.parse(currentDir).root) {
    const gitDir = path.join(currentDir, '.git');
    if (fs.existsSync(gitDir)) {
      return currentDir;
    }
    currentDir = path.dirname(currentDir);
  }

  throw new Error('Could not find Git directory');
}


if (!fs.existsSync(HUSKY_DIR)) {
  console.log(`Creating husky directory: ${HUSKY_DIR}`);
  fs.mkdirSync(HUSKY_DIR, { recursive: true });
}

if (!fs.existsSync(path.join(HUSKY_DIR, '_'))) {
  console.log(`Creating husky _ directory: ${path.join(HUSKY_DIR, '_')}`);
  fs.mkdirSync(path.join(HUSKY_DIR, '_'), { recursive: true });
}

// init husky
try {
  console.log('Initializing husky...');
  execSync('npx husky install', { cwd: GIT_DIR, stdio: 'inherit' });
} catch (error) {
  console.error(`❌ Failed to initialize husky: ${error.message}`);
  process.exit(1);
}

// create lint-staged config file
const lintStagedConfig = {
  "src/**/*.{ts,tsx}": [
    "eslint --fix",
    "prettier --write"
  ],
  "src/**/*.{scss,md}": [
    "prettier --write"
  ]
};

console.log(`Creating lint-staged config: ${path.join(UI_DIR, '.lintstagedrc.json')}`);
fs.writeFileSync(
  path.join(UI_DIR, '.lintstagedrc.json'),
  JSON.stringify(lintStagedConfig, null, 2)
);

// create pre-commit hook
const preCommitContent = `#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

echo "🔍 Start running the code check..."

# Getting the Git Root Directory
GIT_ROOT=$(git rev-parse --show-toplevel)

# Get a list of staging files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACMR)

# Check for files in the ui/ directory
UI_FILES=$(echo "$STAGED_FILES" | grep '^ui/' || echo "")

if [ -n "$UI_FILES" ]; then
  echo "🔎 Discover ui file changes, run code checks..."

  # Switch to the ui directory
  cd "$GIT_ROOT/ui" || {
    echo "❌ Unable to access the UI catalog"
    exit 1
  }

  # 运行 lint-staged
  echo "✨ Running ESLint and Prettier Formatting..."
  npx lint-staged --verbose

  LINT_STAGED_RESULT=$?
  if [ $LINT_STAGED_RESULT -ne 0 ]; then
    echo "❌ Code check failed, please fix the above problem"
    exit $LINT_STAGED_RESULT
  fi

  echo "✅ Code check passes！"
else
  echo "ℹ️ No front-end file changes found, skip code checking"
fi

echo "🎉 Pre-submission check completed"
`;

console.log(`Creating pre-commit hook: ${path.join(HUSKY_DIR, 'pre-commit')}`);
fs.writeFileSync(path.join(HUSKY_DIR, 'pre-commit'), preCommitContent);
execSync(`chmod +x ${path.join(HUSKY_DIR, 'pre-commit')}`);

// create husky.sh
const huskyShContent = `#!/bin/sh
if [ -z "$husky_skip_init" ]; then
  debug () {
    if [ "$HUSKY_DEBUG" = "1" ]; then
      echo "husky (debug) - $1"
    fi
  }

  readonly hook_name="$(basename "$0")"
  debug "starting $hook_name..."

  if [ "$HUSKY" = "0" ]; then
    debug "HUSKY=0, skip hook"
    exit 0
  fi

  if [ -f ~/.huskyrc ]; then
    debug "sourcing ~/.huskyrc"
    . ~/.huskyrc
  fi

  export readonly husky_skip_init=1
  sh -e "$0" "$@"
  exitCode="$?"

  if [ $exitCode != 0 ]; then
    echo "husky - $hook_name hook exited with code $exitCode (error)"
  fi

  exit $exitCode
fi
`;

console.log(`Creating husky.sh: ${path.join(HUSKY_DIR, '_', 'husky.sh')}`);
fs.writeFileSync(
  path.join(HUSKY_DIR, '_', 'husky.sh'),
  huskyShContent
);
execSync(`chmod +x ${path.join(HUSKY_DIR, '_', 'husky.sh')}`);

console.log('Lint setup complete! Husky and lint-staged have been configured.');
console.log(`Git root directory: ${GIT_DIR}`);
console.log(`Husky directory: ${HUSKY_DIR}`);

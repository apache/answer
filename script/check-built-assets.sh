#!/bin/bash
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Builds the frontend and asserts the server can still find the built assets
# inside index.html. See internal/controller/template_controller_test.go for
# why that is not implied by a successful build.
#
# --skip-build   reuse an existing ui/build, do not rebuild
# --self-check   additionally rewrite ui/build/index.html so a script
#                or stylesheet tag is missing, and confirm the check
#                fails on the missing asset. Restores the real build
#                output afterwards.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INDEX_HTML="$REPO_ROOT/ui/build/index.html"
TEST_PACKAGE="./internal/controller/"
TEST_NAME="TestGetStyleResolvesBuiltAssets"

list_output="$(cd "$REPO_ROOT" && go test "$TEST_PACKAGE" -list "^${TEST_NAME}$" 2>&1)"
if ! grep -qx "$TEST_NAME" <<<"$list_output"; then
  echo "no test named $TEST_NAME in $TEST_PACKAGE; go test -run with a stale/renamed test name matches nothing and still exits 0, which would turn this check into a silent no-op" >&2
  exit 1
fi

skip_build=0
self_check=0
for arg in "$@"; do
  case "$arg" in
    --skip-build) skip_build=1 ;;
    --self-check) self_check=1 ;;
    *) echo "unknown option: $arg" >&2; exit 2 ;;
  esac
done

run_check() {
  (cd "$REPO_ROOT" && go test -count=1 "$TEST_PACKAGE" -run "$TEST_NAME" "$@")
}

if [ "$skip_build" -eq 0 ]; then
  echo "==> building frontend"
  (cd "$REPO_ROOT/ui" && pnpm build)
fi

if [ ! -f "$INDEX_HTML" ]; then
  echo "no built index.html at $INDEX_HTML; run without --skip-build" >&2
  exit 1
fi

echo "==> checking the server can parse the built asset tags"
run_check -v

if [ "$self_check" -eq 0 ]; then
  exit 0
fi

# Confirm the check actually fails when a required asset is missing from
# the build output. Without this, a check that silently stopped asserting
# anything would look identical to a passing one.
backup="$(mktemp)"
cp "$INDEX_HTML" "$backup"
trap 'cp "$backup" "$INDEX_HTML"; rm -f "$backup"' EXIT

expect_failure() {
  local label="$1"
  local html="$2"
  printf '%s' "$html" > "$INDEX_HTML"
  echo "==> self-check: expecting failure on $label"
  if run_check >/dev/null 2>&1; then
    echo "SELF-CHECK FAILED: the check passed on $label, so it is not guarding anything" >&2
    exit 1
  fi
  echo "    check failed as expected"
}

expect_failure "stylesheet link but no script src" \
  '<!doctype html><html><head><meta charset="utf-8"/><link rel="stylesheet" crossorigin href="/static/css/index-e5f6a7b8.css"></head><body><div id="root"></div></body></html>'

expect_failure "script src but no stylesheet link" \
  '<!doctype html><html><head><meta charset="utf-8"/><script type="module" crossorigin src="/static/js/index-a1b2c3d4.js"></script></head><body><div id="root"></div></body></html>'

expect_failure "only an inline script, no src" \
  '<!doctype html><html><head><meta charset="utf-8"/><link rel="stylesheet" crossorigin href="/static/css/index-e5f6a7b8.css"></head><body><div id="root"></div><script>window.__inline=1;</script></body></html>'

expect_failure "manifest link but no stylesheet link" \
  '<!doctype html><html><head><meta charset="utf-8"/><script type="module" crossorigin src="/static/js/index-a1b2c3d4.js"></script><link rel="manifest" href="/manifest.json"></head><body><div id="root"></div></body></html>'

echo "==> self-check passed"

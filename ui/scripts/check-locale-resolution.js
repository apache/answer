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

/*
 * src/utils/localize.ts loads a language with a template-literal dynamic
 * import through the @i18n alias, which points outside the frontend root. A
 * bundler that cannot enumerate that pattern still produces a clean build and
 * still serves a working app; the resources simply never arrive. The failure
 * is confined to runtime and to languages other than the default one, so
 * neither a green build nor a smoke test in the default language sees it.
 *
 * So: bundle that same import with the project's own bundler configuration,
 * run it, and require two different languages to come back as different, real
 * translated content.
 */

const fs = require('fs');
const os = require('os');
const path = require('path');
const util = require('util');
const Module = require('module');
const { createRequire } = require('module');

const UI_DIR = path.resolve(__dirname, '..');
const PROBE = path.resolve(__dirname, 'locale-probe.js');
const LOCALIZE_SOURCE = path.resolve(UI_DIR, 'src/utils/localize.ts');
const OVERRIDES_CONFIG = path.resolve(UI_DIR, 'config-overrides.js');

// A language written in a script the default language does not use, so
// "resolved the language that was asked for" cannot be mistaken for "fell
// back to the default language".
const TARGET_LANG = 'zh_CN';
const DEFAULT_LANG = 'en_US';
const TARGET_SCRIPT = /[一-鿿]/;

const rel = (file) => path.relative(UI_DIR, file);

function fail(message) {
  console.error(`FAIL: ${message}`);
  process.exit(1);
}

// The probe carries a copy of the application's import expression, so it is
// only evidence for as long as the two agree.
function assertProbeStillMatchesApplication() {
  const source = fs.readFileSync(LOCALIZE_SOURCE, 'utf8');
  if (!/import\(\s*`@i18n\/\$\{[^}]+\}\.yaml`\s*\)/.test(source)) {
    fail(
      `${rel(LOCALIZE_SOURCE)} no longer loads languages with a template-literal import ` +
        `through @i18n, so ${rel(PROBE)} is exercising a shape the application does not use. ` +
        `Update the probe to match the application, then re-run.`,
    );
  }
}

function bundleWithProjectConfig(outDir) {
  const uiRequire = createRequire(path.join(UI_DIR, 'package.json'));
  const reactScriptsRequire = createRequire(
    uiRequire.resolve('react-scripts/package.json'),
  );

  // config-overrides.js requires webpack directly, and the package layout does
  // not expose webpack at the frontend root. Add the copy react-scripts
  // depends on to the global module search path so the project's real
  // configuration loads instead of an approximation of it.
  process.env.NODE_PATH = path.resolve(
    path.dirname(reactScriptsRequire.resolve('webpack')),
    '..',
    '..',
  );
  Module._initPaths();

  const webpack = reactScriptsRequire('webpack');
  const overrides = require(OVERRIDES_CONFIG);

  // config-overrides.js mutates a config in place, so hand it the minimum
  // shape it reaches into and let it install the real alias table and loaders.
  const projectConfig = overrides.webpack(
    {
      plugins: [],
      resolve: { alias: {}, plugins: [] },
      module: { rules: [{ oneOf: [] }] },
      optimization: {},
    },
    'development',
  );

  const compiler = webpack({
    ...projectConfig,
    mode: 'development',
    devtool: false,
    target: 'node',
    context: UI_DIR,
    entry: PROBE,
    output: {
      path: outDir,
      filename: 'probe.js',
      library: { type: 'commonjs2' },
    },
    // Chunk splitting is a production concern and would scatter the probe.
    optimization: {
      ...projectConfig.optimization,
      splitChunks: false,
      runtimeChunk: false,
    },
  });

  return new Promise((resolve, reject) => {
    compiler.run((err, stats) => {
      compiler.close(() => {});
      if (err) {
        reject(err);
        return;
      }
      if (stats.hasErrors()) {
        reject(new Error(stats.toString({ all: false, errors: true })));
        return;
      }
      resolve();
    });
  });
}

function bundleProbe(outDir) {
  if (fs.existsSync(OVERRIDES_CONFIG)) {
    return bundleWithProjectConfig(outDir);
  }
  fail(
    `no bundler configuration this check knows how to drive was found in ${rel(UI_DIR) || '.'}; ` +
      `teach it how to bundle ${rel(PROBE)} with the current one before relying on it again`,
  );
  return Promise.resolve();
}

function resourcesOf(resConf, langName) {
  // The application reads .ui off the loaded file and registers it with i18next.
  const resources = resConf && resConf.ui;
  if (
    !resources ||
    typeof resources !== 'object' ||
    Object.keys(resources).length === 0
  ) {
    fail(
      `${langName} resolved to ${util.inspect(resConf, { depth: 1 })}, which carries no ui ` +
        `section; the application would register no translations for it`,
    );
  }
  return resources;
}

async function main() {
  assertProbeStillMatchesApplication();

  const outDir = fs.mkdtempSync(path.join(os.tmpdir(), 'answer-locale-check-'));
  try {
    await bundleProbe(outDir);

    const probe = require(path.join(outDir, 'probe.js'));
    const target = resourcesOf(await probe.loadLocaleResource(TARGET_LANG), TARGET_LANG);
    const fallback = resourcesOf(await probe.loadLocaleResource(DEFAULT_LANG), DEFAULT_LANG);

    if (JSON.stringify(target) === JSON.stringify(fallback)) {
      fail(
        `${TARGET_LANG} and ${DEFAULT_LANG} resolved to identical resources; the language ` +
          `name is not selecting a file, so every language would render as ${DEFAULT_LANG}`,
      );
    }

    if (!TARGET_SCRIPT.test(JSON.stringify(target))) {
      fail(
        `${TARGET_LANG} resolved without a single character of its own script; the content ` +
          `is not the translated file`,
      );
    }

    console.log(
      `OK: ${TARGET_LANG} and ${DEFAULT_LANG} both resolve at runtime to distinct translated resources`,
    );
  } finally {
    fs.rmSync(outDir, { recursive: true, force: true });
  }
}

main().catch((err) => fail(err.stack || String(err)));

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
 * build tool that cannot resolve that shape still produces a clean build and
 * still serves a working app; the resources simply never arrive. The failure
 * is confined to runtime and to languages other than the default one, so
 * neither a green build nor a smoke test in the default language sees it.
 *
 * So: start the project's own dev server, load that same import through it,
 * and require two different languages to come back as different, real
 * translated content.
 *
 * Two separate things have to hold, and only one of them is about resolution:
 *
 *   1. The module graph resolves the alias and parses the file. Covered by
 *      importing the probe through the server's module runner.
 *   2. A browser is actually allowed to fetch it. The languages live outside
 *      the frontend root, so they are reachable only if the dev server's
 *      filesystem allow-list includes their directory. The module runner does
 *      not go through that allow-list. A real request does, so this makes one.
 *
 * Checking only the first would pass while every language 403s in a browser.
 */

const fs = require('fs');
const path = require('path');
const util = require('util');

const UI_DIR = path.resolve(__dirname, '..');
const PROBE = path.resolve(__dirname, 'locale-probe.js');
const LOCALIZE_SOURCE = path.resolve(UI_DIR, 'src/utils/localize.ts');
const DEV_SERVER_CONFIG = path.resolve(UI_DIR, 'vite.config.mts');

// A language written in a script the default language does not use, so
// "resolved the language that was asked for" cannot be mistaken for "fell
// back to the default language".
const TARGET_LANG = 'zh_CN';
const DEFAULT_LANG = 'en_US';
const TARGET_SCRIPT = /[一-鿿]/;

const rel = (file) => path.relative(UI_DIR, file);

function fail(message) {
  throw new Error(`FAIL: ${message}`);
}

function withTimeout(promise, ms, message) {
  let timer;
  const timeout = new Promise((_, reject) => {
    timer = setTimeout(() => reject(new Error(`FAIL: ${message}`)), ms);
  });
  return Promise.race([promise, timeout]).finally(() => clearTimeout(timer));
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

async function openProbe() {
  if (!fs.existsSync(DEV_SERVER_CONFIG)) {
    fail(
      `no dev server configuration this check knows how to drive was found at ` +
        `${rel(DEV_SERVER_CONFIG)}; teach it how to load ${rel(PROBE)} with the current ` +
        `tooling before relying on it again`,
    );
  }

  const { createServer, createServerModuleRunner } = await import('vite');
  const server = await createServer({
    root: UI_DIR,
    configFile: DEV_SERVER_CONFIG,
    logLevel: 'warn',
  });
  try {
    await withTimeout(
      server.listen(),
      30000,
      'dev server did not start within 30s',
    );
  } catch (err) {
    // listen() failed or timed out, but the server (and the watcher and
    // websocket server it created before listen() ever ran) already exists.
    // Nothing else will hold a reference to it once this throws, so close it
    // here or it keeps the event loop alive forever.
    await server.close().catch(() => {});
    throw err;
  }

  const ssr = server.environments.ssr;
  const baseUrl = (server.resolvedUrls.local[0] || '').replace(/\/$/, '');

  // Loaded on demand rather than up front, because the reachability check
  // below is only meaningful before anything pulls a language into the module
  // graph. See the comment on assertBrowserCanFetch.
  let probe = null;
  const importProbe = async () => {
    if (!probe) {
      probe = await withTimeout(
        createServerModuleRunner(ssr).import(PROBE),
        30000,
        'module runner did not import the probe within 30s',
      );
    }
    return probe;
  };

  return {
    load: async (langName) =>
      (await importProbe()).loadLocaleResource(langName),

    // Ask for the language file the way the browser will: over HTTP, at the
    // path the dev server assigns to a file outside the frontend root.
    //
    // ORDER MATTERS. Run this before any language is loaded through the module
    // graph. Once a module is in the graph the dev server answers from the
    // transform pipeline instead of reading the file, and the request stops
    // passing through the filesystem allow-list. Checking afterwards returns
    // 200 even when the allow-list would give a browser a 403, which is to say
    // it checks nothing. Resolve the path without loading it, then ask.
    async assertBrowserCanFetch(langName) {
      const specifier = `@i18n/${langName}.yaml`;
      const resolved = await ssr.pluginContainer.resolveId(specifier, PROBE);

      if (!resolved || !resolved.id) {
        fail(
          `${specifier} does not resolve at all, so there is nothing for a browser to request`,
        );
      }

      const url = `${baseUrl}/@fs${resolved.id.split('?')[0]}`;
      let response;
      try {
        response = await fetch(url);
      } catch (err) {
        fail(
          `requesting ${langName} at ${url} failed outright: ${err.message}`,
        );
      }

      if (!response.ok) {
        fail(
          `the dev server answered ${response.status} for ${langName} at ${url}; the language ` +
            `files sit outside the frontend root, so a browser would get this too and every ` +
            `language other than the default would fail to load`,
        );
      }

      const body = await response.text();
      if (!TARGET_SCRIPT.test(body)) {
        fail(
          `the dev server served ${langName} at ${url} but the response carries none of its ` +
            `script; a browser would receive something that is not the translated file`,
        );
      }
    },

    close: () => server.close(),
  };
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

  const probe = await openProbe();
  try {
    // Before any load, while the filesystem allow-list still governs the request.
    await probe.assertBrowserCanFetch(TARGET_LANG);

    const target = resourcesOf(await probe.load(TARGET_LANG), TARGET_LANG);
    const fallback = resourcesOf(await probe.load(DEFAULT_LANG), DEFAULT_LANG);

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
      `OK: ${TARGET_LANG} and ${DEFAULT_LANG} resolve at runtime to distinct translated ` +
        `resources, and ${TARGET_LANG} is fetchable over the dev server`,
    );
  } finally {
    await probe.close();
  }
}

main().catch((err) => {
  console.error(err.message || err.stack || String(err));
  process.exitCode = 1;
});

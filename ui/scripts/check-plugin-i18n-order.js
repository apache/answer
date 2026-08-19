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
 * Plugin i18n modules call initI18nResource while they are being evaluated,
 * and i18next only attaches its resource-store methods to the instance while
 * init runs. So whether registering a plugin's translations works depends on
 * whether that module evaluated before or after init, which is decided by how
 * the bundler groups and orders chunks. Nothing in the application controls
 * it, and when it goes wrong the entry module throws while still evaluating:
 * the build succeeds, the server returns a page, the console stays empty, and
 * the application never mounts.
 *
 * Registering BEFORE init is the order that breaks, so do exactly that. Then
 * initialise, then require the translations to actually be present. That last
 * step is the point: a fix that merely swallowed the error would leave every
 * plugin string untranslated and would otherwise look identical to a working
 * one.
 */

const fs = require('fs');
const path = require('path');
const util = require('util');

const UI_DIR = path.resolve(__dirname, '..');
const PLUGIN_UTILS = path.resolve(UI_DIR, 'src/utils/pluginKit/utils.ts');
const DEV_SERVER_CONFIG = path.resolve(UI_DIR, 'vite.config.mts');

const LANG = 'en_US';
const PLUGIN_NS = 'plugin';
const SLUG = 'check_only_plugin';
const SENTINEL = 'registered before init';

const rel = (file) => path.relative(UI_DIR, file);

function fail(message) {
  console.error(`FAIL: ${message}`);
  process.exit(1);
}

async function main() {
  if (!fs.existsSync(DEV_SERVER_CONFIG)) {
    fail(
      `no dev server configuration this check knows how to drive was found at ` +
        `${rel(DEV_SERVER_CONFIG)}; teach it how to load ${rel(PLUGIN_UTILS)} with the ` +
        `current tooling before relying on it again`,
    );
  }

  const { createServer, createServerModuleRunner } = await import('vite');
  const server = await createServer({
    root: UI_DIR,
    configFile: DEV_SERVER_CONFIG,
    logLevel: 'warn',
    // Without this the helper under test and this check each resolve their own
    // copy of i18next, and the check ends up inspecting an instance nobody
    // registered anything into. It reads as a failure with a confusing message
    // rather than as a broken harness, so the single-instance assertion below
    // guards it too.
    ssr: { noExternal: ['i18next'] },
  });
  await server.listen();

  try {
    const runner = createServerModuleRunner(server.environments.ssr);

    // Deliberately load the plugin helper first and never touch the app's own
    // i18n bootstrap, so i18next is guaranteed to be uninitialised here. This
    // is the ordering the bundler is free to produce.
    const pluginUtils = await runner.import(PLUGIN_UTILS);
    const i18next = (await runner.import('i18next')).default;

    // If the helper and this check hold different copies, every assertion
    // below is measuring an object nothing under test ever touched.
    const loaded = [...server.environments.ssr.moduleGraph.idToModuleMap.keys()].filter(
      (id) => /[\\/]i18next[\\/]/.test(id) && !/[\\/]\.vite[\\/]/.test(id),
    );
    if (loaded.length !== 1) {
      fail(
        `expected exactly one i18next module to be loaded, found ${loaded.length}: ` +
          `${util.inspect(loaded)}. This check can only observe what the helper registers ` +
          `if both resolve the same copy.`,
      );
    }

    if (i18next.isInitialized) {
      fail(
        `i18next was already initialised before this check registered anything, so the ` +
          `ordering the check exists to exercise was never exercised; the check is not ` +
          `proving what it claims`,
      );
    }

    const resource = {
      [LANG]: { plugin: { [SLUG]: { ui: { title: SENTINEL } } } },
    };

    try {
      pluginUtils.initI18nResource(resource);
    } catch (err) {
      fail(
        `registering a plugin's translations before i18next.init threw: ${err.message}\n` +
          `  This is the ordering a bundler is free to produce. When it happens the entry ` +
          `module throws while evaluating and the application never mounts, with no console ` +
          `output and a successful build.`,
      );
    }

    await i18next.init({ lng: LANG, fallbackLng: LANG, resources: {} });

    const bundle = i18next.getResourceBundle(LANG, PLUGIN_NS);
    const title = bundle && bundle[SLUG] && bundle[SLUG].ui && bundle[SLUG].ui.title;

    if (title !== SENTINEL) {
      fail(
        `registering before init did not throw, but the translations never arrived: ` +
          `expected ${util.inspect(SENTINEL)} at ${PLUGIN_NS}.${SLUG}.ui.title, got ` +
          `${util.inspect(bundle)}.\n` +
          `  Surviving the call is not enough. Plugin strings have to actually be ` +
          `registered once i18next is up, or every plugin renders untranslated.`,
      );
    }

    console.log(
      `OK: plugin translations registered before i18next.init survive and are present afterwards`,
    );
  } finally {
    await server.close();
  }
}

main().catch((err) => fail(err.stack || String(err)));

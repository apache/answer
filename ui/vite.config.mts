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

import path from 'path';
import { fileURLToPath } from 'url';

import react from '@vitejs/plugin-react';
import yaml from '@modyfi/vite-plugin-yaml';
import { CORE_SCHEMA } from 'js-yaml';
import { defineConfig, loadEnv } from 'vite';

// This file is loaded as a real ES module, where __dirname does not exist.
const rootDir = path.dirname(fileURLToPath(import.meta.url));
const i18nDir = path.resolve(rootDir, '../i18n');

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, rootDir, 'REACT_APP_');

  // configs/config.yaml ui.public_url, as written by scripts/env.js, may or
  // may not already carry a trailing slash (the root value is exactly "/").
  // Vite requires base to end with one, so add it only when missing rather
  // than concatenating blindly and risking "//".
  //
  // Vite keeps an absolute external base (e.g. a CDN URL) exactly as given
  // only for `vite build`. `vite dev` and `vite preview` reduce the same
  // base to its bare pathname, dropping the scheme and host. That split is
  // intentional here, not a bug to unify: those two commands only serve
  // this app locally, and the Go server only ever embeds `vite build`'s
  // output, so the reduction never reaches anything a real deployment
  // serves.
  const publicUrl = env.REACT_APP_PUBLIC_URL || '/';
  const base = publicUrl.endsWith('/') ? publicUrl : `${publicUrl}/`;

  return {
    // The previous yaml-loader (yaml@2.6.1 core schema) kept bare dates as
    // strings and left merge keys unresolved. @modyfi/vite-plugin-yaml
    // defaults to js-yaml's DEFAULT_SCHEMA, which resolves bare YYYY-MM-DD
    // scalars to JS Date objects and enables merge keys. Pin CORE_SCHEMA so
    // yaml imports keep parsing the way they did before the migration.
    plugins: [react(), yaml({ schema: CORE_SCHEMA })],

    css: {
      preprocessorOptions: {
        // bootstrap 5.3.3's own scss internals emit dozens of deprecation
        // warnings (color functions, mixed-decls) on every build. They are
        // unactionable here and bury warnings that point at our own code.
        scss: { quietDeps: true },
      },
    },

    // scripts/env.js generates .env.production from the server's own
    // configs/config.yaml using REACT_APP_ names. Reading that prefix keeps the
    // generator as the single source of truth for both sides.
    envPrefix: 'REACT_APP_',

    base,

    resolve: {
      alias: {
        '@': path.resolve(rootDir, 'src'),
        '@i18n': i18nDir,
      },
    },

    build: {
      // ui/static.go embeds this directory, and internal/router/ui.go serves
      // /static from it. Neither path is configurable from here.
      outDir: 'build',
      assetsDir: 'static',
      // Matches the previous build so before/after size comparisons measure the
      // bundler rather than a change of sourcemap setting.
      sourcemap: true,
      rollupOptions: {
        output: {
          // Keep emitted files grouped under static/js, static/css and
          // static/media. The analyze script globs that layout, and a flat
          // static/ directory silently matches nothing.
          entryFileNames: 'static/js/[name].[hash].js',
          chunkFileNames: 'static/js/[name].[hash].chunk.js',
          assetFileNames: (assetInfo) => {
            const name = assetInfo.names?.[0] ?? '';
            if (name.endsWith('.css')) {
              return 'static/css/[name].[hash][extname]';
            }
            return 'static/media/[name].[hash][extname]';
          },
        },
      },
    },

    server: {
      port: 3000,
      proxy: {
        '/answer': {
          target: env.REACT_APP_API_URL,
          changeOrigin: true,
          secure: false,
        },
        '/installation': {
          target: env.REACT_APP_API_URL,
          changeOrigin: true,
          secure: false,
        },
        '/custom.css': {
          target: env.REACT_APP_API_URL,
        },
      },
      fs: {
        // Languages live outside this root and are loaded through @i18n.
        allow: [rootDir, i18nDir],
      },
    },
  };
});

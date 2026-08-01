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

import react from '@vitejs/plugin-react';
import yaml from '@modyfi/vite-plugin-yaml';
import { defineConfig, loadEnv } from 'vite';

const i18nDir = path.resolve(__dirname, '../i18n');

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, 'REACT_APP_');

  return {
    plugins: [react(), yaml()],

    // scripts/env.js generates .env.production from the server's own
    // configs/config.yaml using REACT_APP_ names. Reading that prefix keeps the
    // generator as the single source of truth for both sides.
    envPrefix: 'REACT_APP_',

    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
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
          // static/media. The repository's .gitignore and the analyze script
          // both match on that layout, and a flat static/ directory silently
          // escapes both.
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
        allow: [__dirname, i18nDir],
      },
    },
  };
});

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

const {
  addWebpackModuleRule,
  addWebpackAlias,
  setWebpackOptimizationSplitChunks,
} = require("customize-cra");

const path = require("path");
const i18nPath = path.resolve(__dirname, "../i18n");

module.exports = {
  webpack: function(config, env) {
    addWebpackAlias({
      "@": path.resolve(__dirname, "src"),
      "@i18n": i18nPath
    })(config);

    addWebpackModuleRule({
      test: /\.ya?ml$/,
      use: "yaml-loader"
    })(config);

    setWebpackOptimizationSplitChunks({
      maxInitialRequests: 20,
      minSize: 20 * 1024,
      minChunks: 2,
      cacheGroups: {
        automaticNamePrefix: 'chunk',
        components: {
          test: /[\\/]components[\\/]/,
          name: 'components',
          priority: 14,
          reuseExistingChunk: true,
          minChunks: process.env.NODE_ENV === 'production' ? 1 : 2,
          chunks: 'initial',
        },
        i18next: {
          name: 'i18next',
          test: /[\/]node_modules[\/](i18next)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 12,
          reuseExistingChunk: true,
          minChunks: 1,
          chunks: 'initial',
        },
        reactBootstrap: {
          name: 'react-bootstrap',
          test: /[\/]node_modules[\/](react-bootstrap)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 11,
          minChunks: 1,
          chunks: 'initial',
          reuseExistingChunk: true,
        },
        lodash: {
          name: 'lodash',
          test: /[\/]node_modules[\/](lodash)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 10,
          reuseExistingChunk: true,
          minChunks: 1,
          chunks: 'initial',
        },
        codemirror: {
          name: 'codemirror',
          test: /[\/]node_modules[\/](codemirror)[\/]/,
          priority: 9,
          reuseExistingChunk: true,
          enforce: true,
        },
        nextShare: {
          name: 'next-share',
          test: /[\/]node_modules[\/](next-share)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 8,
          reuseExistingChunk: true,
          minChunks: 1,
          chunks: 'initial',
        },
        marked: {
          name: 'marked',
          test: /[\/]node_modules[\/](marked)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 7,
          reuseExistingChunk: true,
          minChunks: 1,
          chunks: 'initial',
        },
        reactDom: {
          name: 'react-dom',
          test: /[\/]node_modules[\/](react-dom)[\/]/,
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          priority: 7,
          reuseExistingChunk: true,
          chunks: 'all',
          enforce: true,
        },
        nodesAsync: {
          name: 'chunk-nodesAsync',
          test: /[\/]node_modules[\/]/,
          priority: 2,
          minChunks: 2,
          chunks: 'async', // only package dependencies that are referenced asynchronously
          reuseExistingChunk: true, // reuse an existing block
        },
        nodesInitial: {
          name: 'chunk-nodesInitial',
          filename: 'static/js/[name].[contenthash:8].chunk.js',
          test: /[\/]node_modules[\/]/,
          priority: 1,
          minChunks: 1,
          chunks: 'initial',
          reuseExistingChunk: true,
        },
      },
    })(config);

    // add i18n dir to ModuleScopePlugin allowedPaths
    const moduleScopePlugin = config.resolve.plugins.find(_ => _.constructor.name === "ModuleScopePlugin");
    if (moduleScopePlugin) {
      moduleScopePlugin.allowedPaths.push(i18nPath);
    }

    return config;
  },
  devServer: function(configFunction) {
    return function(proxy, allowedHost) {
      const config = configFunction(proxy, allowedHost);
      config.proxy = [
        {
          context: ['/answer', '/installation'],
          target: process.env.REACT_APP_API_URL,
          changeOrigin: true,
          secure: false,
        },
        {
          context: ['/custom.css'],
          target: process.env.REACT_APP_API_URL,
        }
      ];
      return config;
    };
  }
};

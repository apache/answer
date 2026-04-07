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

const path = require('path');
const fs = require('fs');
const yaml = require('js-yaml');

const pluginPath = path.join(__dirname, '../src/plugins');
const pluginFolders = fs.readdirSync(pluginPath);

function pascalize(str) {
  return str.split(/[_-]/).map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join('');
}

function resetIndexTs() {
  const indexTsPath = path.join(pluginPath, 'index.ts');
  const existing = fs.readFileSync(indexTsPath, 'utf-8');
  const licenseEnd = existing.indexOf('*/');
  const header = licenseEnd >= 0 ? existing.slice(0, licenseEnd + 2) : '';
  fs.writeFileSync(indexTsPath, `${header}\n\nexport default null;`);
}

function addPluginToIndexTs(packageName, pluginFolder) {
  const indexTsPath = path.join(pluginPath, 'index.ts');
  const indexTsContent = fs.readFileSync(indexTsPath, 'utf-8');
  const lines = indexTsContent.split('\n');
  const ComponentName = pascalize(packageName);

  const importLine = `const load${ComponentName} = () => import('${packageName}').then(module => module.default);`;
  const info = yaml.load(fs.readFileSync(path.join(pluginFolder, 'info.yaml'), 'utf8'));
  const exportLine = `export const ${info.slug_name} = load${ComponentName}`;

  if (!lines.includes(exportLine)) {
    lines.push(importLine);
    lines.push(exportLine);
  }

  fs.writeFileSync(indexTsPath, lines.join('\n'));
}

const pluginLength = pluginFolders.filter((folder) => {
  const pluginFolder = path.join(pluginPath, folder);
  const stat = fs.statSync(pluginFolder);
  return stat.isDirectory() && folder !== 'builtin';
}).length;

if (pluginLength > 0) {
  resetIndexTs();
}

pluginFolders.forEach((folder) => {
  const pluginFolder = path.join(pluginPath, folder);
  const stat = fs.statSync(pluginFolder);

  if (stat.isDirectory() && folder !== 'builtin') {
    if (!fs.existsSync(path.join(pluginFolder, 'index.ts'))) {
      return;
    }
    const packageJson = require(path.join(pluginFolder, 'package.json'));
    const packageName = packageJson.name;

    addPluginToIndexTs(packageName, pluginFolder);
  }
});


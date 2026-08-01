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
const yaml = require('js-yaml');

const configFilePath = path.resolve(__dirname, '../../configs/config.yaml');
const envFilePath = path.resolve(__dirname, '../.env.production');

// Read config.yaml file
const config = yaml.load(fs.readFileSync(configFilePath, 'utf8'));

// Generate .env file content
let envContent = '';
for (const key in config.ui) {
  const value = config.ui[key];
  envContent += `REACT_APP_${key.toUpperCase()}=${value}\n`;
}

// Write .env file
fs.writeFileSync(envFilePath, envContent);
